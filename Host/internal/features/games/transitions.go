package games

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"

	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/abilities"
	chatfeature "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/chat"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func stateFromRecord(record *core.Record) State {
	return State{
		Status:    Status(record.GetString("status")),
		Revision:  record.GetInt("revision"),
		StartedAt: timePointer(record, "started_at"),
		EndedAt:   timePointer(record, "ended_at"),
	}
}

func applyState(record *core.Record, state State) {
	record.Set("status", state.Status)
	record.Set("revision", state.Revision)
	if state.StartedAt == nil {
		record.Set("started_at", nil)
	} else {
		record.Set("started_at", *state.StartedAt)
	}
	if state.EndedAt == nil {
		record.Set("ended_at", nil)
	} else {
		record.Set("ended_at", *state.EndedAt)
	}
}

func transition(command Transition) func(*core.RequestEvent) error {
	return func(event *core.RequestEvent) error {
		game, err := findGame(event)
		if err != nil {
			return httpx.WriteErrorFrom(event, err)
		}
		if command == Start {
			if err := validateStart(event.App, game); err != nil {
				return httpx.WriteErrorFrom(event, err)
			}
		}
		now := time.Now().UTC()
		next, err := ApplyTransition(stateFromRecord(game), command, now)
		if err != nil {
			return httpx.WriteError(event, result.Conflict("game.transition_not_allowed", err.Error()))
		}
		if err := event.App.RunInTransaction(func(tx core.App) error {
			game, err = tx.FindRecordById("games", game.Id)
			if err != nil {
				return err
			}
			if command == BeginReview {
				if _, err := abilities.FinalizePhase(tx, game, now); err != nil {
					return err
				}
			}
			next, err = ApplyTransition(stateFromRecord(game), command, now)
			if err != nil {
				return err
			}
			applyState(game, next)
			if command == Pause && game.GetString("timer_state") == "running" {
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
			if command == BeginReview {
				game.Set("timer_state", "inactive")
				game.Set("timer_total_ms", 0)
				game.Set("timer_remaining_ms", 0)
				game.Set("timer_ends_at", nil)
				game.Set("timer_revision", game.GetInt("timer_revision")+1)
			}
			if command == Pause && game.GetString("timer_state") == "completed" {
				if _, err := abilities.FinalizePhase(tx, game, now); err != nil {
					return err
				}
			}
			if command == Start {
				definition, err := snapshot(game)
				if err != nil {
					return err
				}
				participants, err := currentParticipants(tx, game.Id)
				if err != nil {
					return err
				}
				if err := chatfeature.PrepareRoleRooms(tx, game.Id, definition, participants); err != nil {
					return err
				}
			}
			// A lifecycle command has one externally-visible game revision, even
			// when it also finalizes ability choices or changes timer state.
			if err := tx.Save(game); err != nil {
				return err
			}
			action := "game." + string(command)
			return applicationaudit.Record(tx, event.Auth, game.Id, action, "game", game.Id, nil, event.Get(httpx.TraceIDKey))
		}); err != nil {
			return httpx.WriteErrorFrom(event, err)
		}
		action := "game." + string(command)
		publishGame(event.App, game, action, projectGame(game))
		return event.JSON(http.StatusOK, projectGame(game))
	}
}

func validateStart(app core.App, game *core.Record) error {
	participants, err := currentParticipants(app, game.Id)
	if err != nil {
		return err
	}
	definition, err := snapshot(game)
	if err != nil {
		return err
	}
	if len(participants) < definition.Metadata.MinPlayers || len(participants) > definition.Metadata.MaxPlayers {
		return result.AppError{
			Code: "game.invalid_player_count", Message: "The roster does not meet the ruleset player limits.",
			Status: http.StatusConflict,
		}
	}
	assignments := make([]rulesets.Assignment, len(participants))
	for index, participant := range participants {
		if participant.GetString("role_key") == "" {
			return result.AppError{Code: "game.assignments_incomplete", Message: "Assign a role to every player before starting.", Status: http.StatusConflict}
		}
		assignments[index] = rulesets.Assignment{ParticipantID: participant.Id, RoleID: participant.GetString("role_key")}
	}
	report := rulesets.ValidateAssignments(definition, len(participants), assignments)
	if !report.Valid() {
		return result.AppError{Code: "game.assignments_invalid", Message: "The role assignments do not satisfy the ruleset composition.", Status: http.StatusConflict}
	}
	return nil
}

type archiveRequest struct {
	ConfirmUnsetOutcomes bool `json:"confirmUnsetOutcomes"`
}

func archiveGame(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if game.GetString("status") != string(StatusReview) {
		return httpx.WriteError(event, result.Conflict("game.transition_not_allowed", "Only a game in review can be archived."))
	}
	var request archiveRequest
	_ = event.BindBody(&request)
	participants, err := currentParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	for _, participant := range participants {
		if participant.GetString("outcome") == "unset" && !request.ConfirmUnsetOutcomes {
			return httpx.WriteError(event, result.Conflict("game.outcomes_incomplete", "Set every active player's outcome or explicitly confirm unset outcomes."))
		}
	}
	next, _ := ApplyTransition(stateFromRecord(game), Archive, time.Now().UTC())
	applyState(game, next)
	game.Set("joining_open", false)
	game.Set("join_code", "")
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Save(game); err != nil {
			return err
		}
		if err := chatfeature.FreezeHistoricalAccess(tx, game.Id, time.Now().UTC()); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "game.archived", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	publishGame(event.App, game, "game.archived", projectGame(game))
	return event.JSON(http.StatusOK, projectGame(game))
}

type phaseRequest struct {
	PhaseKey string `json:"phaseKey"`
}

func setPhase(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if game.GetString("status") != string(StatusRunning) && game.GetString("status") != string(StatusPaused) {
		return httpx.WriteError(event, result.Conflict("game.phase_not_allowed", "Phases can only change while a game is running or paused."))
	}
	var request phaseRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game.invalid_phase", "The phase could not be read.", nil))
	}
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	var phase *rulesets.Phase
	for index := range definition.Phases {
		if definition.Phases[index].ID == request.PhaseKey {
			phase = &definition.Phases[index]
			break
		}
	}
	if phase == nil {
		return httpx.WriteError(event, result.Invalid("game.invalid_phase", "That phase is not part of this ruleset.", nil))
	}
	now := time.Now().UTC()
	if err := event.App.RunInTransaction(func(tx core.App) error {
		game, err = tx.FindRecordById("games", game.Id)
		if err != nil {
			return err
		}
		if _, err := abilities.FinalizePhase(tx, game, now); err != nil {
			return err
		}
		if phase.StartsRound {
			game.Set("round_number", game.GetInt("round_number")+1)
		}
		game.Set("phase_key", phase.ID)
		game.Set("phase_started_at", now)
		game.Set("ability_phase_instance", game.GetInt("ability_phase_instance")+1)
		abilities.ResetPhaseLock(game)
		incrementRevision(game)
		if err := tx.Save(game); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "game.phase_changed", "phase", phase.ID, nil, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	payload := map[string]any{"game": projectGame(game), "phase": phase}
	publishGame(event.App, game, "game.phase_changed", payload)
	if phase.AudioCueID != "" {
		for _, cue := range definition.AudioCues {
			if cue.ID == phase.AudioCueID &&
				(cue.DefaultAudience == "all" || cue.DefaultAudience == "game_masters") {
				chatfeature.PublishAudioCue(event.App, game, definition, cue, cue.DefaultAudience, "")
				break
			}
		}
	}
	return event.JSON(http.StatusOK, payload)
}
