package games

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
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
	data, err := json.Marshal(record.Get("ruleset_snapshot"))
	if err != nil {
		return rulesets.DefinitionV1{}, err
	}
	var definition rulesets.DefinitionV1
	if err := json.Unmarshal(data, &definition); err != nil {
		return rulesets.DefinitionV1{}, err
	}
	return definition, nil
}

func projectGame(record *core.Record) map[string]any {
	return map[string]any{
		"id":             record.Id,
		"name":           record.GetString("name"),
		"status":         record.GetString("status"),
		"rulesetVersion": record.GetString("ruleset_version"),
		"joinCode":       record.GetString("join_code"),
		"joiningOpen":    record.GetBool("joining_open"),
		"revision":       record.GetInt("revision"),
		"roundNumber":    record.GetInt("round_number"),
		"phaseKey":       record.GetString("phase_key"),
		"phaseStartedAt": dateValue(record, "phase_started_at"),
		"startedAt":      dateValue(record, "started_at"),
		"endedAt":        dateValue(record, "ended_at"),
		"createdAt":      dateValue(record, "created"),
		"updatedAt":      dateValue(record, "updated"),
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

func activeParticipants(app core.App, gameID string) ([]*core.Record, error) {
	return app.FindRecordsByFilter("participants", "game = {:game} && status != 'kicked' && status != 'left'", "seat_number", 30, 0, dbx.Params{"game": gameID})
}

func audit(app core.App, actor *core.Record, gameID, action, targetType, targetID string, detail any, traceID any) error {
	collection, err := app.FindCollectionByNameOrId("game_audit")
	if err != nil {
		return err
	}
	record := core.NewRecord(collection)
	actorType := "system"
	actorID := ""
	actorLabel := "System"
	if actor != nil {
		actorID = actor.Id
		if actor.Collection().Name == "game_masters" {
			actorType = "game_master"
			actorLabel = actor.GetString("display_name")
		} else {
			actorType = "player"
			actorLabel = actor.GetString("display_name")
		}
	}
	record.Set("game", gameID)
	record.Set("actor_type", actorType)
	record.Set("actor_id", actorID)
	record.Set("actor_label", actorLabel)
	record.Set("action", action)
	record.Set("target_type", targetType)
	record.Set("target_id", targetID)
	record.Set("detail", detail)
	if trace, ok := traceID.(string); ok {
		record.Set("request_id", trace)
	}
	return app.Save(record)
}

func publishGame(app core.App, record *core.Record, kind string, payload any) {
	_ = realtime.Publish(app, "game:"+record.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: record.Id, Revision: record.GetInt("revision"),
		Kind: kind, Payload: payload,
	}, func(auth *core.Record) bool {
		if auth == nil || !auth.GetBool("active") {
			return false
		}
		if auth.Collection().Name == "game_masters" {
			return true
		}
		if auth.Collection().Name != "player_profiles" {
			return false
		}
		records, err := app.FindRecordsByFilter("participants",
			"game = {:game} && profile = {:profile} && status != 'kicked' && status != 'left'",
			"", 1, 0, dbx.Params{"game": record.Id, "profile": auth.Id})
		return err == nil && len(records) == 1
	})
}

func publishGameMasters(app core.App, record *core.Record, kind string, payload any) {
	_ = realtime.Publish(app, "game:"+record.Id+":game-masters", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: record.Id, Revision: record.GetInt("revision"),
		Kind: kind, Payload: payload,
	}, func(auth *core.Record) bool {
		return auth != nil && auth.Collection().Name == "game_masters" && auth.GetBool("active")
	})
}

func writeGameError(event *core.RequestEvent, err error) error {
	if appError, ok := err.(result.AppError); ok {
		return httpx.WriteError(event, appError)
	}
	return httpx.WriteError(event, result.Internal(err))
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
