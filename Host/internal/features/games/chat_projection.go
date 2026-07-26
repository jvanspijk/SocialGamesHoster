package games

import (
	"net/http"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	chatfeature "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/chat"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func visibleRoomsForPlayer(app core.App, game, participant *core.Record, definition rulesets.DefinitionV1) ([]map[string]any, error) {
	memberships, err := app.FindRecordsByFilter("chat_memberships", "participant = {:participant}", "", 200, 0,
		dbx.Params{"participant": participant.Id})
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(memberships))
	for _, membership := range memberships {
		room, err := app.FindRecordById("chat_rooms", membership.GetString("room"))
		if err != nil || room.GetString("game") != game.Id {
			continue
		}
		base, override := roomPolicy(definition, game.GetString("phase_key"), room)
		if room.GetString("kind") == "announcements" || room.GetString("kind") == "gm_dm" {
			base = rulesets.RoomPermission{Visible: true, Readable: true, Sendable: room.GetString("kind") == "gm_dm", SenderDisplay: rulesets.SenderProfileName}
			override = nil
		}
		state := chatfeature.ParticipantState{
			IsMember:       membership.GetDateTime("left_at").IsZero(),
			IsActive:       participant.GetString("status") == "active",
			HistoricalRead: membership.GetBool("historical_access"),
		}
		policy := chatfeature.EffectivePolicy(base, override, state, chatfeature.RoomState{
			ManuallyLocked: room.GetBool("manually_locked"), ManualVisibilityOverride: room.GetString("manual_visibility_override"),
		})
		if game.GetString("status") == string(StatusReview) || game.GetString("status") == string(StatusArchived) {
			policy.Sendable = false
		}
		if game.GetString("status") == string(StatusPaused) && room.GetString("kind") != "gm_dm" {
			explicitlyAllowed := override != nil && override.Sendable != nil && *override.Sendable
			if !explicitlyAllowed {
				policy.Sendable = false
			}
		}
		if !policy.Visible && !policy.Readable {
			continue
		}
		result = append(result, map[string]any{
			"id": room.Id, "key": room.GetString("room_key"), "kind": room.GetString("kind"),
			"label": room.GetString("label"), "visible": policy.Visible, "readable": policy.Readable,
			"sendable": policy.Sendable, "senderDisplay": policy.SenderDisplay,
		})
	}
	return result, nil
}

func roomPolicy(definition rulesets.DefinitionV1, phaseKey string, room *core.Record) (rulesets.RoomPermission, *rulesets.PartialRoomPermission) {
	var base rulesets.RoomPermission
	var override *rulesets.PartialRoomPermission
	switch room.GetString("kind") {
	case "general":
		if definition.Chat.DefaultPolicy.General != nil {
			base = *definition.Chat.DefaultPolicy.General
		}
		if phase, ok := definition.Chat.PhaseOverrides[phaseKey]; ok {
			override = phase.General
		}
	case "player_dm":
		if definition.Chat.DefaultPolicy.PlayerDM != nil {
			base = *definition.Chat.DefaultPolicy.PlayerDM
		}
		if phase, ok := definition.Chat.PhaseOverrides[phaseKey]; ok {
			override = phase.PlayerDM
		}
	case "team":
		base = definition.Chat.DefaultPolicy.Teams[room.GetString("team_key")]
		if phase, ok := definition.Chat.PhaseOverrides[phaseKey]; ok {
			value, exists := phase.Teams[room.GetString("team_key")]
			if exists {
				override = &value
			}
		}
	}
	return base, override
}

type announcementRequest struct {
	Content  string `json:"content"`
	CueKey   string `json:"cueKey"`
	Audience string `json:"audience"`
	TargetID string `json:"targetId"`
}

func createAnnouncement(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	if game.GetString("status") != string(StatusLobby) && game.GetString("status") != string(StatusRunning) && game.GetString("status") != string(StatusPaused) {
		return httpx.WriteError(event, result.Conflict("chat.read_only", "Announcements are only available during a live game."))
	}
	var request announcementRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("chat.invalid_message", "Enter an announcement of at most 1000 characters.", nil))
	}
	request.Content = strings.TrimSpace(request.Content)
	if len([]rune(request.Content)) > 1000 || request.Content == "" || containsControlCharacter(request.Content) {
		return httpx.WriteError(event, result.Invalid("chat.invalid_message", "Enter an announcement of at most 1000 characters.", nil))
	}
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	var cue *rulesets.AudioCue
	if request.CueKey != "" {
		for index := range definition.AudioCues {
			if definition.AudioCues[index].ID == request.CueKey {
				cue = &definition.AudioCues[index]
				break
			}
		}
		if cue == nil {
			return httpx.WriteError(event, result.Invalid("audio.invalid_cue", "That sound cue is not part of this game.", nil))
		}
		if request.Audience == "" {
			request.Audience = cue.DefaultAudience
		}
		if appError := validateCueAudience(event.App, game, definition, request.Audience, request.TargetID); appError != nil {
			return httpx.WriteError(event, *appError)
		}
	}
	room, err := findRoom(event.App, game.Id, "announcements")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	collection, err := event.App.FindCollectionByNameOrId("chat_messages")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	message := core.NewRecord(collection)
	message.Set("room", room.Id)
	message.Set("message_kind", "announcement")
	message.Set("sender_type", "game_master")
	message.Set("sender_id", event.Auth.Id)
	message.Set("sender_label_snapshot", event.Auth.GetString("display_name"))
	message.Set("content", request.Content)
	message.Set("cue_key", request.CueKey)
	if err := event.App.Save(message); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projected := map[string]any{
		"id": message.Id, "roomId": room.Id, "kind": "announcement",
		"senderLabel": message.GetString("sender_label_snapshot"), "content": request.Content,
		"cueKey": request.CueKey, "createdAt": dateValue(message, "created"),
	}
	_ = realtime.Publish(event.App, "room:"+room.Id, realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: "chat.announcement", Payload: projected,
	}, func(auth *core.Record) bool { return auth != nil && auth.GetBool("active") })
	if cue != nil {
		publishAudioCue(event.App, game, definition, *cue, request.Audience, request.TargetID)
	}
	_ = audit(event.App, event.Auth, game.Id, "chat.announcement_sent", "chat_message", message.Id,
		map[string]any{"cueKey": request.CueKey, "audience": request.Audience, "targetId": request.TargetID}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusCreated, projected)
}

func containsControlCharacter(value string) bool {
	for _, character := range value {
		if (character < 0x20 && character != '\n' && character != '\t') || character == 0x7f {
			return true
		}
	}
	return false
}

func validateCueAudience(app core.App, game *core.Record, definition rulesets.DefinitionV1, audience, targetID string) *result.AppError {
	switch audience {
	case "all", "game_masters":
		if targetID != "" {
			value := result.Invalid("audio.invalid_audience", "This audience does not accept a target.", nil)
			return &value
		}
	case "team":
		found := false
		for _, team := range definition.Teams {
			found = found || team.ID == targetID
		}
		if !found {
			value := result.Invalid("audio.invalid_target", "Choose a team from this ruleset.", nil)
			return &value
		}
	case "player":
		participant, err := app.FindRecordById("participants", targetID)
		if err != nil || participant.GetString("game") != game.Id ||
			participant.GetString("status") == "kicked" || participant.GetString("status") == "left" {
			value := result.Invalid("audio.invalid_target", "Choose a player in this game.", nil)
			return &value
		}
	default:
		value := result.Invalid("audio.invalid_audience", "Choose who should hear the sound cue.", nil)
		return &value
	}
	return nil
}

func publishAudioCue(app core.App, game *core.Record, definition rulesets.DefinitionV1, cue rulesets.AudioCue, audience, targetID string) {
	assets, err := app.FindRecordsByFilter(
		"ruleset_assets",
		"ruleset_version = {:version} && asset_key = {:key} && kind = 'audio'",
		"",
		1,
		0,
		dbx.Params{"version": game.GetString("ruleset_version"), "key": cue.AssetKey},
	)
	if err != nil || len(assets) != 1 {
		return
	}
	asset := assets[0]
	payload := map[string]any{
		"cueKey": cue.ID, "name": cue.Name, "assetId": asset.Id,
		"preview": "/api/app/v1/ruleset-assets/" + asset.Id,
	}
	_ = realtime.Publish(app, "game:"+game.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: "audio.cue", Payload: payload,
	}, func(auth *core.Record) bool {
		if auth == nil || !auth.GetBool("active") {
			return false
		}
		if audience == "game_masters" {
			return auth.Collection().Name == "game_masters"
		}
		if auth.Collection().Name != "player_profiles" {
			return false
		}
		participants, err := app.FindRecordsByFilter(
			"participants",
			"game = {:game} && profile = {:profile} && status != 'kicked' && status != 'left'",
			"",
			1,
			0,
			dbx.Params{"game": game.Id, "profile": auth.Id},
		)
		if err != nil || len(participants) != 1 {
			return false
		}
		participant := participants[0]
		switch audience {
		case "all":
			return true
		case "player":
			return participant.Id == targetID
		case "team":
			for _, role := range definition.Roles {
				if role.ID == participant.GetString("role_key") {
					return role.TeamID == targetID
				}
			}
		}
		return false
	})
}
