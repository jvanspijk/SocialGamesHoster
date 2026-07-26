package timer

import (
	"net/http"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	platformauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/auth"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type commandRequest struct {
	DurationMS int64 `json:"durationMs"`
	DeltaMS    int64 `json:"deltaMs"`
}

func Register(event *core.ServeEvent, service *Service) {
	group := event.Router.Group("/api/app/v1")
	group.GET("/games/{id}/timer", getTimer)
	group.POST("/games/{id}/timer/start", timerCommand(service, "start")).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/timer/pause", timerCommand(service, "pause")).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/timer/resume", timerCommand(service, "resume")).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/timer/adjust", timerCommand(service, "adjust")).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/timer/stop", timerCommand(service, "stop")).BindFunc(platformauth.RequireGameMaster)
}

func getTimer(event *core.RequestEvent) error {
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
	}
	if !canViewTimer(event, game) {
		return httpx.WriteError(event, result.Forbidden("timer.forbidden", "This timer is not available."))
	}
	now := time.Now().UTC()
	state := stateFromRecord(game)
	reconciled := Reconcile(state, now)
	if reconciled.Status != state.Status || reconciled.Remaining != state.Remaining {
		saveState(game, reconciled)
		if err := event.App.Save(game); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	return event.JSON(http.StatusOK, Project(reconciled, now))
}

func canViewTimer(event *core.RequestEvent, game *core.Record) bool {
	auth := event.Auth
	if auth == nil || !auth.GetBool("active") {
		return false
	}
	if auth.Collection().Name == "game_masters" {
		return true
	}
	if auth.Collection().Name != "player_profiles" {
		return false
	}
	records, err := event.App.FindRecordsByFilter("participants",
		"game = {:game} && profile = {:profile} && status != 'kicked' && status != 'left'",
		"", 1, 0, dbx.Params{"game": game.Id, "profile": auth.Id})
	return err == nil && len(records) == 1
}

func timerCommand(service *Service, command string) func(*core.RequestEvent) error {
	return func(event *core.RequestEvent) error {
		game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
		if err != nil {
			return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
		}
		status := game.GetString("status")
		if status != "running" && status != "paused" {
			return httpx.WriteError(event, result.Conflict("timer.not_allowed", "The timer is only available while a game is running or paused."))
		}
		var request commandRequest
		if command == "start" || command == "adjust" {
			if err := event.BindBody(&request); err != nil {
				return httpx.WriteError(event, result.Invalid("timer.invalid", "The timer command could not be read.", nil))
			}
		}
		now := time.Now().UTC()
		state := stateFromRecord(game)
		var next State
		switch command {
		case "start":
			next, err = Start(state, time.Duration(request.DurationMS)*time.Millisecond, now)
		case "pause":
			next, err = Pause(state, now)
		case "resume":
			next, err = Resume(state, now)
		case "adjust":
			next, err = Adjust(state, time.Duration(request.DeltaMS)*time.Millisecond, now)
		case "stop":
			next = Stop(state)
		default:
			err = result.Invalid("timer.invalid", "Unknown timer command.", nil)
		}
		if err != nil {
			return httpx.WriteError(event, result.Conflict("timer.transition_not_allowed", err.Error()))
		}
		saveState(game, next)
		if err := event.App.Save(game); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		service.Schedule(game.Id, next)
		projection := Project(next, now)
		service.Publish(game, "timer."+command, projection)
		return event.JSON(http.StatusOK, projection)
	}
}

func stateFromRecord(record *core.Record) State {
	state := State{
		Status:    Status(record.GetString("timer_state")),
		Total:     time.Duration(record.GetInt("timer_total_ms")) * time.Millisecond,
		Remaining: time.Duration(record.GetInt("timer_remaining_ms")) * time.Millisecond,
		Revision:  record.GetInt("timer_revision"),
	}
	value := record.GetDateTime("timer_ends_at")
	if !value.IsZero() {
		endsAt := value.Time().UTC()
		state.EndsAt = &endsAt
	}
	return state
}

func saveState(record *core.Record, state State) {
	record.Set("timer_state", state.Status)
	record.Set("timer_total_ms", state.Total.Milliseconds())
	record.Set("timer_remaining_ms", state.Remaining.Milliseconds())
	if state.EndsAt == nil {
		record.Set("timer_ends_at", nil)
	} else {
		record.Set("timer_ends_at", *state.EndsAt)
	}
	record.Set("timer_revision", state.Revision)
}
