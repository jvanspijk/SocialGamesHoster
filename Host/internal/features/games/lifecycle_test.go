package games

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func TestLifecycleHappyPath(t *testing.T) {
	now := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	state := State{Status: StatusDraft}
	var err error
	for _, transition := range []Transition{OpenLobby, Start, Pause, Resume, BeginReview, Archive} {
		state, err = ApplyTransition(state, transition, now)
		if err != nil {
			t.Fatalf("%s: %v", transition, err)
		}
	}
	if state.Status != StatusArchived || state.Revision != 6 || state.StartedAt == nil || state.EndedAt == nil {
		t.Fatalf("unexpected final state: %#v", state)
	}
}

func TestLifecycleRejectsInvalidAndArchivedTransitions(t *testing.T) {
	now := time.Now()
	if _, err := ApplyTransition(State{Status: StatusDraft}, Pause, now); err == nil {
		t.Fatal("expected invalid transition to fail")
	}
	if _, err := ApplyTransition(State{Status: StatusArchived}, ReturnToRunning, now); err == nil {
		t.Fatal("expected archived transition to fail")
	}
}

func TestValidateAssignmentRolesAllowsPartialAssignments(t *testing.T) {
	definition := rulesets.DefinitionV1{Roles: []rulesets.Role{{ID: "dealer"}}}
	if appError := validateAssignmentRoles(definition, []rulesets.Assignment{{
		ParticipantID: "first-player",
		RoleID:        "dealer",
	}}); appError != nil {
		t.Fatalf("partial assignment was rejected: %#v", appError)
	}
	if appError := validateAssignmentRoles(definition, []rulesets.Assignment{{
		ParticipantID: "first-player",
		RoleID:        "unknown",
	}}); appError == nil {
		t.Fatal("unknown role was accepted")
	}
}

func TestStartingClosesJoining(t *testing.T) {
	state, err := ApplyTransition(State{Status: StatusLobby, JoiningOpen: true}, Start, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if state.JoiningOpen {
		t.Fatal("starting a game must stop new players from joining")
	}
}

func TestArchivedGameDeletionRemainsAnExplicitException(t *testing.T) {
	if !canDeleteGame(StatusArchived) {
		t.Fatal("archived games must remain deletable with exact confirmation")
	}
	for _, status := range []Status{StatusLobby, StatusRunning, StatusPaused} {
		if canDeleteGame(status) {
			t.Fatalf("%q game unexpectedly deletable", status)
		}
	}
}

func TestOnlyAnUnstartedLobbyCanBeCancelled(t *testing.T) {
	for _, status := range []Status{StatusDraft, StatusRunning, StatusPaused, StatusReview, StatusArchived} {
		if canCancelGame(status) {
			t.Fatalf("%q game unexpectedly cancellable", status)
		}
	}
	if !canCancelGame(StatusLobby) {
		t.Fatal("an unstarted lobby must be cancellable")
	}
}

func TestArchivedRosterMutationUsesSharedConflict(t *testing.T) {
	game := core.NewRecord(&core.Collection{})
	game.Id = "game"
	game.Set("status", StatusArchived)
	participant := core.NewRecord(&core.Collection{})
	participant.Id = "participant"
	participant.Set("game", game.Id)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/games/game/participants/participant", nil)
	request.SetPathValue("id", game.Id)
	request.SetPathValue("participantId", participant.Id)
	event := &core.RequestEvent{}
	event.App = gameRecordLookupApp{
		App:     nil,
		records: map[string]*core.Record{game.Id: game, participant.Id: participant},
	}
	event.Request = request
	event.Response = recorder

	if err := updateParticipant(event); err != nil {
		t.Fatal(err)
	}
	assertArchivedConflict(t, recorder)
}

type gameRecordLookupApp struct {
	core.App
	records map[string]*core.Record
}

func (app gameRecordLookupApp) FindRecordById(
	collection any,
	recordID string,
	optFilters ...func(*dbx.SelectQuery) error,
) (*core.Record, error) {
	record := app.records[recordID]
	if record == nil {
		return nil, errors.New("record not found")
	}
	return record, nil
}

func assertArchivedConflict(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	var response struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusConflict || response.Code != "game.archived_immutable" {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
