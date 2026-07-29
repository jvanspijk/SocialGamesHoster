package games

import (
	"net/http"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/abilities"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func startCompletion(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	status := game.GetString("status")
	if status != string(StatusRunning) && status != string(StatusPaused) {
		return httpx.WriteError(event, result.Conflict("game.completion_not_allowed", "Only a running or paused game can be finished."))
	}
	if _, err := abilities.FinalizePhase(event.App, game.Id, time.Now().UTC()); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	game, err = event.App.FindRecordById("games", game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	game.Set("completion_previous_status", status)
	if status == string(StatusRunning) && game.GetString("timer_state") == "running" {
		remaining := game.GetDateTime("timer_ends_at").Time().Sub(time.Now().UTC()).Milliseconds()
		if remaining < 0 {
			remaining = 0
		}
		game.Set("timer_remaining_ms", remaining)
		game.Set("timer_ends_at", nil)
		if remaining == 0 {
			game.Set("timer_state", "completed")
		} else {
			game.Set("timer_state", "paused")
		}
		game.Set("timer_revision", game.GetInt("timer_revision")+1)
	}
	next, err := ApplyTransition(stateFromRecord(game), BeginReview, time.Now().UTC())
	if err != nil {
		return httpx.WriteError(event, result.Conflict("game.completion_not_allowed", err.Error()))
	}
	applyState(game, next)
	if err := event.App.Save(game); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = audit(event.App, event.Auth, game.Id, "game.completion_started", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "game.completion_started", projectGame(game))
	return event.JSON(http.StatusOK, projectGame(game))
}

func cancelCompletion(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if game.GetString("status") != string(StatusReview) {
		return httpx.WriteError(event, result.Conflict("game.completion_not_active", "This game is not in the completion flow."))
	}
	previous := game.GetString("completion_previous_status")
	if previous != string(StatusPaused) {
		previous = string(StatusRunning)
	}
	game.Set("status", previous)
	game.Set("completion_previous_status", "")
	if previous == string(StatusRunning) && game.GetString("timer_state") == "paused" && game.GetInt("timer_remaining_ms") > 0 {
		game.Set("timer_state", "running")
		game.Set("timer_ends_at", time.Now().UTC().Add(time.Duration(game.GetInt("timer_remaining_ms"))*time.Millisecond))
		game.Set("timer_revision", game.GetInt("timer_revision")+1)
	}
	incrementRevision(game)
	if err := event.App.Save(game); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = audit(event.App, event.Auth, game.Id, "game.completion_cancelled", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "game.completion_cancelled", projectGame(game))
	return event.JSON(http.StatusOK, projectGame(game))
}

func gameSummary(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if game.GetString("status") != string(StatusReview) && game.GetString("status") != string(StatusArchived) {
		return httpx.WriteError(event, result.Conflict("game.summary_not_available", "The summary is available while finishing or after archiving a game."))
	}
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	participants, err := gameParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	awards, err := event.App.FindRecordsByFilter("achievement_awards", "game = {:game}", "created", 500, 0, dbx.Params{"game": game.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projectedParticipants := make([]map[string]any, 0, len(participants))
	for _, participant := range participants {
		item := projectParticipant(participant, true)
		playerAwards := make([]map[string]any, 0)
		for _, award := range awards {
			if award.GetString("profile") == participant.GetString("profile") {
				playerAwards = append(playerAwards, map[string]any{
					"id": award.Id, "title": award.GetString("title_snapshot"),
					"description": award.GetString("description_snapshot"),
					"points":      award.GetInt("points_snapshot"),
				})
			}
		}
		item["achievements"] = playerAwards
		projectedParticipants = append(projectedParticipants, item)
	}
	durationMS := int64(0)
	if started := timePointer(game, "started_at"); started != nil {
		end := time.Now().UTC()
		if ended := timePointer(game, "ended_at"); ended != nil {
			end = *ended
		}
		durationMS = end.Sub(*started).Milliseconds()
	}
	return event.JSON(http.StatusOK, map[string]any{
		"game": projectGame(game),
		"ruleset": map[string]any{
			"name":          definition.Metadata.Name,
			"coverAssetKey": definition.Metadata.CoverAssetKey,
		},
		"durationMs":   durationMS,
		"participants": projectedParticipants,
		"immutable":    game.GetString("status") == string(StatusArchived),
	})
}
