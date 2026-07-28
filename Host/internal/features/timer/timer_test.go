package timer

import (
	"testing"
	"time"
)

func TestTimerTransitionsAndReconciliation(t *testing.T) {
	now := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	state, err := Start(State{Status: Inactive}, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	state, err = Pause(state, now.Add(15*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != Paused || state.Remaining != 45*time.Second {
		t.Fatalf("unexpected pause: %#v", state)
	}
	state, err = Resume(state, now.Add(20*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	state = Reconcile(state, now.Add(70*time.Second))
	if state.Status != Completed || state.Remaining != 0 {
		t.Fatalf("expected completion after delayed scheduler: %#v", state)
	}
}

func TestAddingTimeToCompletedTimerCreatesFreshPausedTimer(t *testing.T) {
	now := time.Now()
	state := State{Status: Completed, Total: time.Minute, Revision: 4}
	state, err := Adjust(state, 30*time.Second, now)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != Paused || state.Total != 30*time.Second || state.Remaining != 30*time.Second {
		t.Fatalf("unexpected adjusted state: %#v", state)
	}
}

func TestAdjustmentCannotGoBelowZero(t *testing.T) {
	now := time.Now()
	state := State{Status: Paused, Total: 10 * time.Second, Remaining: 5 * time.Second}
	if _, err := Adjust(state, -6*time.Second, now); err == nil {
		t.Fatal("expected negative adjustment to fail")
	}
}

func TestInactiveTimerIsProjected(t *testing.T) {
	projected := Project(State{Status: Inactive}, time.Now())
	if projected == nil || projected.Status != Inactive || projected.RemainingMS != 0 {
		t.Fatal("inactive timer must be projected explicitly")
	}
}
