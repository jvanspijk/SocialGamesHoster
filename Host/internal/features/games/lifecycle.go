package games

import (
	"fmt"
	"slices"
	"time"
)

type Status string

const (
	StatusDraft    Status = "draft"
	StatusLobby    Status = "lobby"
	StatusRunning  Status = "running"
	StatusPaused   Status = "paused"
	StatusReview   Status = "review"
	StatusArchived Status = "archived"
)

type Transition string

const (
	OpenLobby       Transition = "open_lobby"
	CancelLobby     Transition = "cancel_lobby"
	Start           Transition = "start"
	Pause           Transition = "pause"
	Resume          Transition = "resume"
	BeginReview     Transition = "begin_review"
	ReturnToRunning Transition = "return_to_running"
	Archive         Transition = "archive"
)

type State struct {
	Status    Status
	Revision  int
	StartedAt *time.Time
	EndedAt   *time.Time
}

func ApplyTransition(state State, transition Transition, now time.Time) (State, error) {
	if state.Status == StatusArchived {
		return state, fmt.Errorf("archived games are immutable")
	}

	allowed := map[Transition][]Status{
		OpenLobby:       {StatusDraft},
		CancelLobby:     {StatusLobby},
		Start:           {StatusLobby},
		Pause:           {StatusRunning},
		Resume:          {StatusPaused},
		BeginReview:     {StatusRunning, StatusPaused},
		ReturnToRunning: {StatusReview},
		Archive:         {StatusReview},
	}
	if !slices.Contains(allowed[transition], state.Status) {
		return state, fmt.Errorf("transition %s is not allowed from %s", transition, state.Status)
	}

	next := state
	switch transition {
	case OpenLobby:
		next.Status = StatusLobby
	case CancelLobby:
		next.Status = StatusDraft
	case Start:
		next.Status = StatusRunning
		if next.StartedAt == nil {
			startedAt := now.UTC()
			next.StartedAt = &startedAt
		}
	case Pause:
		next.Status = StatusPaused
	case Resume, ReturnToRunning:
		next.Status = StatusRunning
	case BeginReview:
		next.Status = StatusReview
	case Archive:
		next.Status = StatusArchived
		endedAt := now.UTC()
		next.EndedAt = &endedAt
	default:
		return state, fmt.Errorf("unknown transition %q", transition)
	}
	next.Revision++
	return next, nil
}

func IsLive(status Status) bool {
	return status == StatusLobby || status == StatusRunning || status == StatusPaused
}
