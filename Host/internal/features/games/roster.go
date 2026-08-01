package games

import (
	"crypto/rand"
	"encoding/binary"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	chatfeature "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/chat"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type participantUpdateRequest struct {
	GameAlias string `json:"gameAlias"`
}

func updateParticipant(event *core.RequestEvent) error {
	game, participant, err := rosterTarget(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if appError := gamepolicyapp.GameMutationError(game); appError != nil {
		return httpx.WriteError(event, *appError)
	}
	var request participantUpdateRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("participant.invalid", "The roster update could not be read.", nil))
	}
	request.GameAlias = strings.TrimSpace(request.GameAlias)
	if len([]rune(request.GameAlias)) > 32 {
		return httpx.WriteError(event, result.Invalid("participant.invalid_alias", "An alias may contain at most 32 characters.", nil))
	}
	participant.Set("game_alias", request.GameAlias)
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Save(participant); err != nil {
			return err
		}
		incrementRevision(game)
		return tx.Save(game)
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return participantChanged(event, game, participant, "participant.updated")
}

func setParticipantStatus(status gamepolicy.ParticipantStatus) func(*core.RequestEvent) error {
	return func(event *core.RequestEvent) error {
		game, participant, err := rosterTarget(event)
		if err != nil {
			return httpx.WriteErrorFrom(event, err)
		}
		if appError := gamepolicyapp.GameMutationError(game); appError != nil {
			return httpx.WriteError(event, *appError)
		}
		switch status {
		case gamepolicy.ParticipantKicked:
			if gamepolicy.ParticipantStatus(participant.GetString("status")) == gamepolicy.ParticipantKicked {
				return event.JSON(http.StatusOK, projectParticipant(participant, true))
			}
		case gamepolicy.ParticipantEliminated:
			if game.GetString("status") != string(StatusRunning) && game.GetString("status") != string(StatusPaused) {
				return httpx.WriteError(event, result.Conflict("participant.elimination_not_allowed", "Players can only be eliminated during play."))
			}
		case gamepolicy.ParticipantActive:
			if gamepolicy.ParticipantStatus(participant.GetString("status")) != gamepolicy.ParticipantEliminated {
				return httpx.WriteError(event, result.Conflict("participant.reinstate_not_allowed", "Only an eliminated player can be reinstated."))
			}
		}
		err = event.App.RunInTransaction(func(tx core.App) error {
			participant.Set("status", status)
			if status == gamepolicy.ParticipantEliminated {
				participant.Set("eliminated_at", time.Now().UTC())
			} else if status == gamepolicy.ParticipantActive {
				participant.Set("eliminated_at", nil)
			}
			if err := tx.Save(participant); err != nil {
				return err
			}
			if status == gamepolicy.ParticipantKicked {
				if err := chatfeature.CloseParticipantMemberships(tx, participant.Id, time.Now().UTC()); err != nil {
					return err
				}
			}
			incrementRevision(game)
			return tx.Save(game)
		})
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		return participantChanged(event, game, participant, "participant."+string(status))
	}
}

func rosterTarget(event *core.RequestEvent) (*core.Record, *core.Record, error) {
	game, err := findGame(event)
	if err != nil {
		return nil, nil, err
	}
	participant, err := event.App.FindRecordById("participants", event.Request.PathValue("participantId"))
	if err != nil || participant.GetString("game") != game.Id {
		return nil, nil, result.AppError{Code: "participant.not_found", Message: "Participant not found.", Status: http.StatusNotFound}
	}
	return game, participant, nil
}

func participantChanged(event *core.RequestEvent, game, participant *core.Record, kind string) error {
	_ = applicationaudit.Record(event.App, event.Auth, game.Id, kind, "participant", participant.Id,
		map[string]any{"status": participant.GetString("status")}, event.Get(httpx.TraceIDKey))
	public := projectParticipant(participant, false)
	publishGame(event.App, game, kind, public)
	private := projectParticipant(participant, true)
	publishGameMasters(event.App, game, kind, private)
	_ = realtime.Publish(event.App, "participant:"+participant.Id+":private", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: kind, Payload: projectParticipantForPlayer(game, participant),
	}, func(auth *core.Record) bool {
		return auth != nil && auth.GetBool("active") &&
			(actorauth.IsGameMaster(auth) ||
				(actorauth.IsPlayer(auth) && auth.Id == participant.GetString("profile")))
	})
	return event.JSON(http.StatusOK, private)
}

type assignmentRequest struct {
	Assignments []rulesets.Assignment `json:"assignments"`
}

func putAssignments(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if game.GetString("status") != string(StatusLobby) {
		return httpx.WriteError(event, result.Conflict("game.assignments_not_allowed", "Assignments can only change in the lobby."))
	}
	var request assignmentRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game.assignments_invalid", "The assignments could not be read.", nil))
	}
	participants, err := currentParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if appError := validateAssignmentParticipants(participants, request.Assignments); appError != nil {
		return httpx.WriteError(event, *appError)
	}
	report := rulesets.ValidateAssignments(definition, len(participants), request.Assignments)
	if !report.Valid() {
		return httpx.WriteError(event, result.Invalid("game.assignments_invalid", "The assignments do not satisfy the ruleset composition.",
			result.FieldErrors{"assignments": report.Errors}))
	}
	if err := saveAssignments(event.App, game, request.Assignments, event.Auth.Id); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return assignmentsChanged(event, game)
}

func randomizeAssignments(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if game.GetString("status") != string(StatusLobby) {
		return httpx.WriteError(event, result.Conflict("game.assignments_not_allowed", "Assignments can only change in the lobby."))
	}
	var request assignmentRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game.assignments_invalid", "The locked assignments could not be read.", nil))
	}
	participants, err := currentParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	ids := make([]string, len(participants))
	for index, participant := range participants {
		ids[index] = participant.Id
	}
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	seedBytes := [8]byte{}
	if _, err := rand.Read(seedBytes[:]); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	assignments, err := rulesets.RandomizeAssignments(definition, ids, request.Assignments, binary.LittleEndian.Uint64(seedBytes[:]))
	if err != nil {
		return httpx.WriteError(event, result.Conflict("game.assignment_unsatisfiable", err.Error()))
	}
	if err := saveAssignments(event.App, game, assignments, event.Auth.Id); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return assignmentsChanged(event, game)
}

func validateAssignmentParticipants(participants []*core.Record, assignments []rulesets.Assignment) *result.AppError {
	allowed := make(map[string]bool, len(participants))
	for _, participant := range participants {
		allowed[participant.Id] = true
	}
	seen := map[string]bool{}
	for _, assignment := range assignments {
		if !allowed[assignment.ParticipantID] || seen[assignment.ParticipantID] {
			value := result.Invalid("game.assignments_invalid", "Every active participant must be assigned exactly once.", nil)
			return &value
		}
		seen[assignment.ParticipantID] = true
	}
	if len(seen) != len(participants) {
		value := result.Invalid("game.assignments_invalid", "Every active participant must be assigned exactly once.", nil)
		return &value
	}
	return nil
}

func saveAssignments(app core.App, game *core.Record, assignments []rulesets.Assignment, gameMasterID string) error {
	return app.RunInTransaction(func(tx core.App) error {
		for _, assignment := range assignments {
			participant, err := tx.FindRecordById("participants", assignment.ParticipantID)
			if err != nil || participant.GetString("game") != game.Id {
				return result.AppError{Code: "participant.not_found", Message: "Participant not found.", Status: http.StatusNotFound}
			}
			participant.Set("role_key", assignment.RoleID)
			participant.Set("role_revision", participant.GetInt("role_revision")+1)
			participant.Set("assigned_by", gameMasterID)
			if err := tx.Save(participant); err != nil {
				return err
			}
		}
		incrementRevision(game)
		return tx.Save(game)
	})
}

func assignmentsChanged(event *core.RequestEvent, game *core.Record) error {
	participants, err := currentParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	admin := make([]map[string]any, len(participants))
	for index, participant := range participants {
		admin[index] = projectParticipant(participant, true)
		_ = realtime.Publish(event.App, "participant:"+participant.Id+":private", realtime.Event[any]{
			EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
			Kind: "assignments.changed", Payload: projectParticipantForPlayer(game, participant),
		}, func(auth *core.Record) bool {
			return auth != nil && auth.GetBool("active") &&
				(actorauth.IsGameMaster(auth) ||
					(actorauth.IsPlayer(auth) && auth.Id == participant.GetString("profile")))
		})
	}
	_ = applicationaudit.Record(event.App, event.Auth, game.Id, "assignments.changed", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	publishGameMasters(event.App, game, "assignments.changed", admin)
	publishGame(event.App, game, "game.revision_changed", map[string]any{"revision": game.GetInt("revision")})
	return event.JSON(http.StatusOK, map[string]any{"revision": game.GetInt("revision"), "assignments": admin})
}

func projectParticipantForPlayer(game, participant *core.Record) map[string]any {
	projected := projectParticipant(participant, false)
	projected["roleAvailable"] = game.GetBool("roles_visible") && participant.GetString("role_key") != ""
	projected["roleRevision"] = participant.GetInt("role_revision")
	return projected
}

type outcomeItem struct {
	ParticipantID string `json:"participantId"`
	Outcome       string `json:"outcome"`
}

type outcomesRequest struct {
	Outcomes []outcomeItem `json:"outcomes"`
}

func putOutcomes(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if game.GetString("status") != string(StatusRunning) && game.GetString("status") != string(StatusPaused) && game.GetString("status") != string(StatusReview) {
		return httpx.WriteError(event, result.Conflict("game.outcomes_not_allowed", "Outcomes can only change during play or review."))
	}
	var request outcomesRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game.outcomes_invalid", "The outcomes could not be read.", nil))
	}
	err = event.App.RunInTransaction(func(tx core.App) error {
		seen := map[string]bool{}
		for _, item := range request.Outcomes {
			if seen[item.ParticipantID] || (item.Outcome != "unset" && item.Outcome != "win" && item.Outcome != "loss" && item.Outcome != "draw") {
				return result.Invalid("game.outcomes_invalid", "Each outcome must be unset, win, loss, or draw.", nil)
			}
			seen[item.ParticipantID] = true
			participant, err := tx.FindRecordById("participants", item.ParticipantID)
			if err != nil || participant.GetString("game") != game.Id {
				return result.AppError{Code: "participant.not_found", Message: "Participant not found.", Status: http.StatusNotFound}
			}
			participant.Set("outcome", item.Outcome)
			if err := tx.Save(participant); err != nil {
				return err
			}
		}
		incrementRevision(game)
		return tx.Save(game)
	})
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	_ = applicationaudit.Record(event.App, event.Auth, game.Id, "outcomes.changed", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "outcomes.changed", map[string]any{"revision": game.GetInt("revision")})
	return event.JSON(http.StatusOK, map[string]any{"revision": game.GetInt("revision")})
}
