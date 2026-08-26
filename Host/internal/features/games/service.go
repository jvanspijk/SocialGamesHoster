package games

import (
	"fmt"
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

func findLiveGame(app core.App) (*core.Record, error) {
	records, err := app.FindRecordsByFilter("games", "status = 'lobby' || status = 'running' || status = 'paused'", "", 1, 0)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no live game")
	}
	return records[0], nil
}

func findGame(event *core.RequestEvent) (*core.Record, error) {
	record, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return nil, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound}
	}
	return record, nil
}

func snapshot(record *core.Record) (rulesets.DefinitionV1, error) {
	return rulesets.DecodeSnapshot(record.Get("ruleset_snapshot"))
}

func projectGame(record *core.Record) map[string]any {
	return map[string]any{
		"id":                     record.Id,
		"name":                   record.GetString("name"),
		"status":                 record.GetString("status"),
		"rulesetVersion":         record.GetString("ruleset_version"),
		"joiningOpen":            record.GetBool("joining_open"),
		"rolesVisible":           record.GetBool("roles_visible"),
		"roleVisibilityRevision": record.GetInt("role_visibility_revision"),
		"revision":               record.GetInt("revision"),
		"roundNumber":            record.GetInt("round_number"),
		"phaseKey":               record.GetString("phase_key"),
		"phaseStartedAt":         dateValue(record, "phase_started_at"),
		"abilityPhaseLockedAt":   dateValue(record, "ability_phase_locked_at"),
		"startedAt":              dateValue(record, "started_at"),
		"endedAt":                dateValue(record, "ended_at"),
		"createdAt":              dateValue(record, "created"),
		"updatedAt":              dateValue(record, "updated"),
	}
}

func projectParticipant(record *core.Record, includeRole bool) map[string]any {
	projected := map[string]any{
		"id":                  record.Id,
		"profileId":           record.GetString("profile"),
		"displayNameSnapshot": record.GetString("display_name_snapshot"),
		"gameAlias":           record.GetString("game_alias"),
		"seatNumber":          record.GetInt("seat_number"),
		"status":              record.GetString("status"),
		"joinedAt":            dateValue(record, "joined_at"),
		"eliminatedAt":        dateValue(record, "eliminated_at"),
	}
	if includeRole {
		projected["roleKey"] = record.GetString("role_key")
		projected["roleRevision"] = record.GetInt("role_revision")
		projected["outcome"] = record.GetString("outcome")
	}
	return projected
}

func dateValue(record *core.Record, field string) any {
	value := record.GetDateTime(field)
	if value.IsZero() {
		return nil
	}
	return value.Time().UTC()
}

func gameParticipants(app core.App, gameID string) ([]*core.Record, error) {
	return app.FindRecordsByFilter("participants", "game = {:game}", "seat_number", 30, 0, dbx.Params{"game": gameID})
}

func currentParticipants(app core.App, gameID string) ([]*core.Record, error) {
	return gamepolicyapp.CurrentParticipantsByGame(app, gameID)
}

func publishGame(app core.App, record *core.Record, kind string, payload any) {
	_ = realtime.Publish(app, "game:"+record.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: record.Id, Revision: record.GetInt("revision"),
		Kind: kind, Payload: payload,
	}, func(auth *core.Record) bool {
		if auth == nil || !auth.GetBool("active") {
			return false
		}
		if actorauth.IsGameMaster(auth) {
			return true
		}
		if !actorauth.IsPlayer(auth) {
			return false
		}
		return gamepolicyapp.ProfileParticipatesInGame(app, record.Id, auth.Id)
	})
}

// publishLobbyOpened lets signed-in profiles that have not yet joined the new
// lobby refresh their waiting screen. Game events remain participant-scoped.
func publishLobbyOpened(app core.App, record *core.Record) {
	profiles, err := app.FindRecordsByFilter(actorauth.PlayerProfilesCollection, "active = true", "", 10000, 0)
	if err != nil {
		return
	}
	for _, profile := range profiles {
		_ = realtime.Publish(app, "profile:"+profile.Id, realtime.Event[any]{
			EventID: realtime.NewEventID(), GameID: record.Id, Revision: record.GetInt("revision"),
			Kind: "game.lobby_opened", Payload: projectGame(record),
		}, func(auth *core.Record) bool {
			return actorauth.IsActivePlayer(auth) && auth.Id == profile.Id
		})
	}
}

func publishGameMasters(app core.App, record *core.Record, kind string, payload any) {
	_ = realtime.Publish(app, "game:"+record.Id+":game-masters", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: record.Id, Revision: record.GetInt("revision"),
		Kind: kind, Payload: payload,
	}, func(auth *core.Record) bool {
		return actorauth.IsActiveGameMaster(auth)
	})
}

func incrementRevision(record *core.Record) {
	record.Set("revision", record.GetInt("revision")+1)
}

func timePointer(record *core.Record, field string) *time.Time {
	value := record.GetDateTime(field)
	if value.IsZero() {
		return nil
	}
	parsed := value.Time().UTC()
	return &parsed
}
