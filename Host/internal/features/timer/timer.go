package timer

import (
	"fmt"
	"time"
)

type Status string

const (
	Inactive  Status = "inactive"
	Running   Status = "running"
	Paused    Status = "paused"
	Completed Status = "completed"
)

type State struct {
	Status    Status        `json:"status"`
	Total     time.Duration `json:"-"`
	Remaining time.Duration `json:"-"`
	EndsAt    *time.Time    `json:"endsAt,omitempty"`
	Revision  int           `json:"revision"`
}

type Projection struct {
	Status      Status `json:"status"`
	TotalMS     int64  `json:"totalMs"`
	RemainingMS int64  `json:"remainingMs"`
	EndsAt      string `json:"endsAt,omitempty"`
	Revision    int    `json:"revision"`
	ServerTime  string `json:"serverTime"`
}

func Start(state State, duration time.Duration, now time.Time) (State, error) {
	if duration <= 0 {
		return state, fmt.Errorf("timer duration must be greater than zero")
	}
	end := now.UTC().Add(duration)
	return State{
		Status:    Running,
		Total:     duration,
		Remaining: duration,
		EndsAt:    &end,
		Revision:  state.Revision + 1,
	}, nil
}

func Pause(state State, now time.Time) (State, error) {
	state = Reconcile(state, now)
	if state.Status != Running || state.EndsAt == nil {
		return state, fmt.Errorf("only a running timer can be paused")
	}
	remaining := state.EndsAt.Sub(now.UTC())
	if remaining <= 0 {
		return complete(state), nil
	}
	state.Status = Paused
	state.Remaining = remaining
	state.EndsAt = nil
	state.Revision++
	return state, nil
}

func Resume(state State, now time.Time) (State, error) {
	if state.Status != Paused || state.Remaining <= 0 {
		return state, fmt.Errorf("only a paused timer with remaining time can be resumed")
	}
	end := now.UTC().Add(state.Remaining)
	state.Status = Running
	state.EndsAt = &end
	state.Revision++
	return state, nil
}

func Adjust(state State, delta time.Duration, now time.Time) (State, error) {
	state = Reconcile(state, now)
	switch state.Status {
	case Inactive:
		return state, fmt.Errorf("inactive timers cannot be adjusted")
	case Completed:
		if delta < 0 {
			return state, fmt.Errorf("completed timers can only have time added")
		}
		if delta == 0 {
			return state, nil
		}
		state.Status = Paused
		state.Total = delta
		state.Remaining = delta
		state.EndsAt = nil
	case Paused:
		remaining := state.Remaining + delta
		total := state.Total + delta
		if remaining < 0 || total < 0 {
			return state, fmt.Errorf("timer adjustment cannot reduce time below zero")
		}
		state.Total = total
		state.Remaining = remaining
		if remaining == 0 {
			state.Status = Completed
		}
	case Running:
		if state.EndsAt == nil {
			return state, fmt.Errorf("running timer is missing its end timestamp")
		}
		remaining := state.EndsAt.Sub(now.UTC()) + delta
		total := state.Total + delta
		if remaining < 0 || total < 0 {
			return state, fmt.Errorf("timer adjustment cannot reduce time below zero")
		}
		if remaining == 0 {
			state = complete(state)
			state.Total = total
		} else {
			end := now.UTC().Add(remaining)
			state.Total = total
			state.Remaining = remaining
			state.EndsAt = &end
		}
	default:
		return state, fmt.Errorf("unknown timer state %q", state.Status)
	}
	state.Revision++
	return state, nil
}

func Stop(state State) State {
	return State{Status: Inactive, Revision: state.Revision + 1}
}

func Reconcile(state State, now time.Time) State {
	if state.Status == Running && state.EndsAt != nil {
		remaining := state.EndsAt.Sub(now.UTC())
		if remaining <= 0 {
			return complete(state)
		}
		state.Remaining = remaining
	}
	return state
}

func Project(state State, now time.Time) *Projection {
	state = Reconcile(state, now)
	if state.Status == Inactive {
		return nil
	}
	projection := &Projection{
		Status:      state.Status,
		TotalMS:     state.Total.Milliseconds(),
		RemainingMS: state.Remaining.Milliseconds(),
		Revision:    state.Revision,
		ServerTime:  now.UTC().Format(time.RFC3339Nano),
	}
	if state.EndsAt != nil {
		projection.EndsAt = state.EndsAt.UTC().Format(time.RFC3339Nano)
	}
	return projection
}

func complete(state State) State {
	state.Status = Completed
	state.Remaining = 0
	state.EndsAt = nil
	return state
}
