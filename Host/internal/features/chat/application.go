package chat

import (
	"net/http"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

// EnsureLobbyRoom creates the optional ruleset-owned general room. The caller
// supplies its transaction so room setup composes with the game transition.
func EnsureLobbyRoom(app core.App, gameID string, definition rulesets.DefinitionV1) error {
	if definition.Chat.DefaultPolicy.General == nil {
		return nil
	}
	_, err := ensureRoom(app, gameID, "general", "general", "General", "")
	return err
}

// AddParticipant creates the participant's game-master room and joins any
// general room. The caller supplies its transaction.
func AddParticipant(app core.App, gameID string, participant *core.Record) error {
	dm, err := ensureRoom(
		app,
		gameID,
		"gm:"+participant.Id,
		"gm_dm",
		"Game master · "+participant.GetString("display_name_snapshot"),
		"",
	)
	if err != nil {
		return err
	}
	if err := ensureChatMembership(app, dm.Id, participant.Id); err != nil {
		return err
	}
	if general, findErr := findRoomByKey(app, gameID, "general"); findErr == nil {
		return ensureChatMembership(app, general.Id, participant.Id)
	}
	return nil
}

// PrepareRoleRooms materializes team and custom-channel membership from the
// frozen ruleset snapshot. It never starts a transaction.
func PrepareRoleRooms(app core.App, gameID string, definition rulesets.DefinitionV1, participants []*core.Record) error {
	roleTeam := map[string]string{}
	roles := map[string]rulesets.Role{}
	teamName := map[string]string{}
	for _, role := range definition.Roles {
		roleTeam[role.ID] = role.TeamID
		roles[role.ID] = role
	}
	for _, team := range definition.Teams {
		teamName[team.ID] = team.Name
	}
	for teamID := range definition.Chat.DefaultPolicy.Teams {
		room, err := ensureRoom(app, gameID, "team:"+teamID, "team", teamName[teamID], teamID)
		if err != nil {
			return err
		}
		for _, participant := range participants {
			if roleTeam[participant.GetString("role_key")] == teamID {
				if err := ensureChatMembership(app, room.Id, participant.Id); err != nil {
					return err
				}
			}
		}
	}
	for _, channel := range definition.Chat.Channels {
		room, err := ensureRoom(app, gameID, rulesets.CustomChatRoomPrefix+channel.ID, "custom", channel.Name, "")
		if err != nil {
			return err
		}
		for _, participant := range participants {
			role, ok := roles[participant.GetString("role_key")]
			if ok && rulesets.ChatChannelAudienceMatches(channel, role, false) {
				if err := ensureChatMembership(app, room.Id, participant.Id); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CloseParticipantMemberships records that a removed participant can no
// longer use their current rooms. It never starts a transaction.
func CloseParticipantMemberships(app core.App, participantID string, leftAt time.Time) error {
	memberships, err := app.FindRecordsByFilter(
		"chat_memberships",
		"participant = {:participant} && left_at = ''",
		"",
		200,
		0,
		dbx.Params{"participant": participantID},
	)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		membership.Set("left_at", leftAt)
		if err := app.Save(membership); err != nil {
			return err
		}
	}
	return nil
}

// FreezeHistoricalAccess closes current memberships while retaining their
// archived-history read grant. It never starts a transaction.
func FreezeHistoricalAccess(app core.App, gameID string, leftAt time.Time) error {
	memberships, err := app.FindRecordsByFilter(
		"chat_memberships",
		"room.game = {:game}",
		"",
		1000,
		0,
		dbx.Params{"game": gameID},
	)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		membership.Set("historical_access", true)
		if membership.GetDateTime("left_at").IsZero() {
			membership.Set("left_at", leftAt)
		}
		if err := app.Save(membership); err != nil {
			return err
		}
	}
	return nil
}

// ClearGameSession removes all chat and announcement-owned state for a game.
// The aggregate orchestrator supplies the transaction.
func ClearGameSession(app core.App, gameID string) error {
	items, err := app.FindRecordsByFilter("attention_items", "game = {:game}", "", 10000, 0, dbx.Params{"game": gameID})
	if err != nil {
		return err
	}
	for _, item := range items {
		attachments, err := app.FindRecordsByFilter("announcement_attachments", "announcement = {:item}", "", 10, 0, dbx.Params{"item": item.Id})
		if err != nil {
			return err
		}
		for _, attachment := range attachments {
			if err := app.Delete(attachment); err != nil {
				return err
			}
		}
		receipts, err := app.FindRecordsByFilter("attention_receipts", "attention_item = {:item}", "", 10000, 0, dbx.Params{"item": item.Id})
		if err != nil {
			return err
		}
		for _, receipt := range receipts {
			if err := app.Delete(receipt); err != nil {
				return err
			}
		}
		if err := app.Delete(item); err != nil {
			return err
		}
	}
	rooms, err := app.FindRecordsByFilter("chat_rooms", "game = {:game}", "", 500, 0, dbx.Params{"game": gameID})
	if err != nil {
		return err
	}
	for _, room := range rooms {
		for _, collection := range []string{"chat_messages", "chat_memberships"} {
			dependents, err := app.FindRecordsByFilter(collection, "room = {:room}", "", 10000, 0, dbx.Params{"room": room.Id})
			if err != nil {
				return err
			}
			for _, dependent := range dependents {
				if err := app.Delete(dependent); err != nil {
					return err
				}
			}
		}
		if err := app.Delete(room); err != nil {
			return err
		}
	}
	return nil
}

func AdminViewData(app core.App, gameID string) ([]map[string]any, []map[string]any) {
	rooms, _ := app.FindRecordsByFilter("chat_rooms", "game = {:game} && kind != 'announcements'", "kind,label", 200, 0, dbx.Params{"game": gameID})
	items, _ := app.FindRecordsByFilter("attention_items", "game = {:game}", "-created,-id", 50, 0, dbx.Params{"game": gameID})
	attention := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if summary, err := projectAdminAttentionSummary(app, item); err == nil {
			attention = append(attention, summary)
		}
	}
	return projectAdminRooms(app, rooms), attention
}

func ensureRoom(app core.App, gameID, key, kind, label, teamKey string) (*core.Record, error) {
	if room, err := findRoomByKey(app, gameID, key); err == nil {
		return room, nil
	}
	collection, err := app.FindCollectionByNameOrId("chat_rooms")
	if err != nil {
		return nil, err
	}
	room := core.NewRecord(collection)
	room.Set("game", gameID)
	room.Set("room_key", key)
	room.Set("kind", kind)
	room.Set("label", label)
	room.Set("team_key", teamKey)
	room.Set("players_can_post", true)
	room.Set("manual_visibility_override", "default")
	room.Set("sender_display", "profile_name")
	if err := app.Save(room); err != nil {
		return nil, err
	}
	return room, nil
}

func projectAdminRooms(app core.App, rooms []*core.Record) []map[string]any {
	result := make([]map[string]any, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, map[string]any{
			"id": room.Id, "key": room.GetString("room_key"), "kind": room.GetString("kind"),
			"label": room.GetString("label"), "teamKey": room.GetString("team_key"),
			"playersCanPost": room.GetBool("players_can_post"),
			"latestMessage":  latestMessageSummary(app, room.Id),
		})
	}
	return result
}

func findAnnouncementGame(event *core.RequestEvent) (*core.Record, error) {
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return nil, resultNotFoundGame()
	}
	return game, nil
}

func resultNotFoundGame() error {
	return result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound}
}

func currentAnnouncementParticipants(app core.App, gameID string) ([]*core.Record, error) {
	return gamepolicyapp.CurrentParticipantsByGame(app, gameID)
}

func recordDateValue(record *core.Record, field string) any {
	value := record.GetDateTime(field)
	if value.IsZero() {
		return nil
	}
	return value.Time().UTC()
}

func publishAnnouncementGameMasters(app core.App, game *core.Record, kind string, payload any) {
	_ = realtime.Publish(app, "game:"+game.Id+":game-masters", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: kind, Payload: payload,
	}, func(auth *core.Record) bool {
		return actorauth.IsActiveGameMaster(auth)
	})
}
