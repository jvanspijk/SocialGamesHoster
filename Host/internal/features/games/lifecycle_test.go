package games

import (
	"testing"
	"time"
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

func TestLifecycleCanCancelLobbyToDraft(t *testing.T) {
	now := time.Now()
	state, err := ApplyTransition(State{Status: StatusLobby, Revision: 4}, CancelLobby, now)
	if err != nil {
		t.Fatalf("cancel lobby: %v", err)
	}
	if state.Status != StatusDraft || state.Revision != 5 {
		t.Fatalf("unexpected cancelled lobby state: %#v", state)
	}
}
