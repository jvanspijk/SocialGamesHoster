package timer

import (
	"errors"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
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

func TestCompletionTransactionIsIdempotentAgainstStaleSchedulerWork(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)
	masters, _ := app.FindCollectionByNameOrId("game_masters")
	master := core.NewRecord(masters)
	master.Set("username", "timer-host")
	master.Set("display_name", "Timer host")
	master.Set("active", true)
	master.SetPassword("secret-password")
	if err := app.Save(master); err != nil {
		t.Fatal(err)
	}
	rulesets, _ := app.FindCollectionByNameOrId("rulesets")
	ruleset := core.NewRecord(rulesets)
	ruleset.Set("slug", "timer-test")
	ruleset.Set("name", "Timer test")
	ruleset.Set("created_by", master.Id)
	if err := app.Save(ruleset); err != nil {
		t.Fatal(err)
	}
	versions, _ := app.FindCollectionByNameOrId("ruleset_versions")
	version := core.NewRecord(versions)
	version.Set("ruleset", ruleset.Id)
	version.Set("version_number", 1)
	version.Set("state", "published")
	version.Set("schema_version", 1)
	version.Set("definition", map[string]any{"schemaVersion": 1})
	version.Set("created_by", master.Id)
	if err := app.Save(version); err != nil {
		t.Fatal(err)
	}
	collection, err := app.FindCollectionByNameOrId("games")
	if err != nil {
		t.Fatal(err)
	}
	game := core.NewRecord(collection)
	game.Set("name", "Timer test")
	game.Set("status", "running")
	game.Set("ruleset_version", version.Id)
	game.Set("ruleset_snapshot", map[string]any{"schemaVersion": 1})
	game.Set("created_by", master.Id)
	game.Set("timer_state", "running")
	game.Set("phase_key", "night")
	game.Set("timer_total_ms", 1000)
	game.Set("timer_remaining_ms", 1000)
	game.Set("timer_ends_at", time.Now().UTC().Add(-time.Second))
	game.Set("timer_revision", 7)
	if err := app.Save(game); err != nil {
		t.Fatal(err)
	}
	service := NewService(app)
	previousAudit := auditRecord
	auditRecord = func(core.App, *core.Record, string, string, string, string, any, any) error {
		return errors.New("audit unavailable")
	}
	completed, err := service.completeTransaction(game.Id, 7, time.Now().UTC())
	auditRecord = previousAudit
	if err == nil || completed {
		t.Fatalf("audit failure completion = %t, %v", completed, err)
	}
	current, err := app.FindRecordById("games", game.Id)
	if err != nil || current.GetString("timer_state") != string(Running) || current.GetInt("revision") != 0 || !current.GetDateTime("ability_phase_locked_at").IsZero() {
		t.Fatalf("audit failure persisted timer state: %#v %v", current, err)
	}
	failedAudits, err := app.FindAllRecords("game_audit")
	if err != nil || len(failedAudits) != 0 {
		t.Fatalf("audit failure persisted audit: %d %v", len(failedAudits), err)
	}
	completed, err = service.completeTransaction(game.Id, 7, time.Now().UTC())
	if err != nil || !completed {
		t.Fatalf("first completion: %t %v", completed, err)
	}
	completed, err = service.completeTransaction(game.Id, 7, time.Now().UTC())
	if err != nil || completed {
		t.Fatalf("stale completion repeated work: %t %v", completed, err)
	}
	game, err = app.FindRecordById("games", game.Id)
	if err != nil || game.GetString("timer_state") != string(Completed) || game.GetInt("revision") != 1 {
		t.Fatalf("completion state: %#v %v", game, err)
	}
	if game.GetDateTime("ability_phase_locked_at").IsZero() {
		t.Fatal("completion did not retain ability phase lock")
	}
	audits, err := app.FindRecordsByFilter("game_audit", "game = {:game} && action = 'timer.completed'", "", 10, 0, dbx.Params{"game": game.Id})
	if err != nil || len(audits) != 1 {
		t.Fatalf("completion audit count: %d %v", len(audits), err)
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
