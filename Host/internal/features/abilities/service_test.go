package abilities

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
)

func TestAbilityOwnershipPhaseCombinationUndoAndFinalization(t *testing.T) {
	app, game, profile, participant, definition := abilityFixture(t)
	now := time.Now().UTC()

	_, err := Activate(app, game.Id, profile.Id, "foreign", now)
	assertAbilityError(t, err, "ability.not_owned")
	_, err = Activate(app, game.Id, profile.Id, "day_only", now)
	assertAbilityError(t, err, "ability.phase_not_allowed")

	choices, err := Activate(app, game.Id, profile.Id, "solo", now)
	if err != nil || len(choices) != 1 || choices[0].Status != "Activated" {
		t.Fatalf("activate solo: %#v, %v", choices, err)
	}
	_, err = Activate(app, game.Id, profile.Id, "combo_one", now)
	assertAbilityError(t, err, "ability.combination_conflict")
	choices, err = Undo(app, game.Id, profile.Id, "solo")
	if err != nil || len(choices) != 0 {
		t.Fatalf("undo solo: %#v, %v", choices, err)
	}

	for _, id := range []string{"combo_one", "combo_two"} {
		if _, err := Activate(app, game.Id, profile.Id, id, now); err != nil {
			t.Fatalf("activate %s: %v", id, err)
		}
	}
	progress, results, err := ProjectAdmin(app, game, definition, []*core.Record{participant})
	if err != nil || progress["activatedPlayerCount"] != 1 || len(results) != 0 {
		t.Fatalf("pending admin projection leaked or lost progress: %#v %#v %v", progress, results, err)
	}

	locked := false
	err = app.RunInTransaction(func(tx core.App) error {
		var finalizeErr error
		current, err := tx.FindRecordById("games", game.Id)
		if err != nil {
			return err
		}
		locked, finalizeErr = FinalizePhase(tx, current, now.Add(time.Minute))
		if finalizeErr != nil {
			return finalizeErr
		}
		if locked {
			current.Set("revision", current.GetInt("revision")+1)
			return tx.Save(current)
		}
		return nil
	})
	if err != nil || !locked {
		t.Fatalf("finalize: %t %v", locked, err)
	}
	game, _ = app.FindRecordById("games", game.Id)
	choices, err = ProjectPlayer(app, game, participant, definition)
	if err != nil || len(choices) != 2 || choices[0].Status != "Finalized" || choices[1].Status != "Finalized" {
		t.Fatalf("final player choices: %#v, %v", choices, err)
	}
	_, err = Undo(app, game.Id, profile.Id, "combo_one")
	assertAbilityError(t, err, "ability.phase_locked")
	progress, results, err = ProjectAdmin(app, game, definition, []*core.Record{participant})
	if err != nil || progress["locked"] != true || len(results) != 1 {
		t.Fatalf("final admin projection: %#v %#v %v", progress, results, err)
	}
}

func TestFinalizePhaseRollsBackWithCallerTransaction(t *testing.T) {
	app, game, profile, participant, definition := abilityFixture(t)
	now := time.Now().UTC()
	if _, err := Activate(app, game.Id, profile.Id, "solo", now); err != nil {
		t.Fatalf("activate: %v", err)
	}
	err := app.RunInTransaction(func(tx core.App) error {
		current, err := tx.FindRecordById("games", game.Id)
		if err != nil {
			return err
		}
		locked, err := FinalizePhase(tx, current, now.Add(time.Minute))
		if err != nil || !locked {
			t.Fatalf("finalize in transaction: %t %v", locked, err)
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}
	game, _ = app.FindRecordById("games", game.Id)
	if !game.GetDateTime("ability_phase_locked_at").IsZero() {
		t.Fatal("ability phase lock persisted after rollback")
	}
	choices, err := ProjectPlayer(app, game, participant, definition)
	if err != nil || len(choices) != 1 || choices[0].Status != "Activated" {
		t.Fatalf("ability choice persisted as finalized after rollback: %#v, %v", choices, err)
	}
}

func TestConcurrentCombinableActivationIsAtomic(t *testing.T) {
	app, game, profile, _, _ := abilityFixture(t)
	var wait sync.WaitGroup
	errs := make(chan error, 2)
	for _, abilityID := range []string{"combo_one", "combo_two"} {
		wait.Add(1)
		go func(id string) {
			defer wait.Done()
			_, err := Activate(app, game.Id, profile.Id, id, time.Now().UTC())
			errs <- err
		}(abilityID)
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent combinable activation failed: %v", err)
		}
	}
	records, err := app.FindRecordsByFilter("ability_choices", "game = {:game}", "", 10, 0, dbx.Params{"game": game.Id})
	if err != nil || len(records) != 2 {
		t.Fatalf("expected two durable choices, got %d, %v", len(records), err)
	}
}

func abilityFixture(t *testing.T) (core.App, *core.Record, *core.Record, *core.Record, rulesets.DefinitionV1) {
	t.Helper()
	app := testutil.NewPocketBaseApp(t)
	gmCollection, _ := app.FindCollectionByNameOrId("game_masters")
	gm := core.NewRecord(gmCollection)
	gm.Set("username", "host")
	gm.Set("display_name", "Host")
	gm.Set("active", true)
	gm.SetPassword("secret-password")
	if err := app.Save(gm); err != nil {
		t.Fatal(err)
	}
	profileCollection, _ := app.FindCollectionByNameOrId("player_profiles")
	profile := core.NewRecord(profileCollection)
	profile.Set("display_name", "Alice")
	profile.Set("normalized_name", "alice")
	profile.Set("active", true)
	profile.Set("approved_at", time.Now().UTC())
	profile.SetPassword("device-secret")
	if err := app.Save(profile); err != nil {
		t.Fatal(err)
	}
	definition := rulesets.DefinitionV1{
		SchemaVersion: 1,
		Metadata:      rulesets.Metadata{Name: "Ability test", MinPlayers: 1, MaxPlayers: 2},
		Teams:         []rulesets.Team{{ID: "town", Name: "Town"}},
		Phases:        []rulesets.Phase{{ID: "night", Name: "Night", Order: 1}, {ID: "day", Name: "Day", Order: 2}},
		Abilities: []rulesets.Ability{
			{ID: "solo", Name: "Solo", ActivationPhaseIDs: []string{"night"}},
			{ID: "combo_one", Name: "Combo one", ActivationPhaseIDs: []string{"night"}, CanCombineWithOtherAbilities: true},
			{ID: "combo_two", Name: "Combo two", ActivationPhaseIDs: []string{"night"}, CanCombineWithOtherAbilities: true},
			{ID: "day_only", Name: "Day only", ActivationPhaseIDs: []string{"day"}},
			{ID: "foreign", Name: "Foreign", ActivationPhaseIDs: []string{"night"}},
		},
		Roles: []rulesets.Role{{
			ID: "seer", Name: "Seer", TeamID: "town", MaxCopies: 1,
			AbilityIDs: []string{"solo", "combo_one", "combo_two", "day_only"},
		}},
	}
	rulesetCollection, _ := app.FindCollectionByNameOrId("rulesets")
	logical := core.NewRecord(rulesetCollection)
	logical.Set("slug", "ability-test")
	logical.Set("name", "Ability test")
	logical.Set("created_by", gm.Id)
	if err := app.Save(logical); err != nil {
		t.Fatal(err)
	}
	versionCollection, _ := app.FindCollectionByNameOrId("ruleset_versions")
	version := core.NewRecord(versionCollection)
	version.Set("ruleset", logical.Id)
	version.Set("version_number", 1)
	version.Set("state", "published")
	version.Set("schema_version", 1)
	version.Set("definition", definition)
	version.Set("created_by", gm.Id)
	if err := app.Save(version); err != nil {
		t.Fatal(err)
	}
	gameCollection, _ := app.FindCollectionByNameOrId("games")
	game := core.NewRecord(gameCollection)
	game.Set("name", "Ability test")
	game.Set("status", "running")
	game.Set("ruleset_version", version.Id)
	game.Set("ruleset_snapshot", definition)
	game.Set("phase_key", "night")
	game.Set("phase_started_at", time.Now().UTC())
	game.Set("round_number", 1)
	game.Set("timer_state", "inactive")
	game.Set("created_by", gm.Id)
	if err := app.Save(game); err != nil {
		t.Fatal(err)
	}
	participantCollection, _ := app.FindCollectionByNameOrId("participants")
	participant := core.NewRecord(participantCollection)
	participant.Set("game", game.Id)
	participant.Set("profile", profile.Id)
	participant.Set("display_name_snapshot", "Alice")
	participant.Set("seat_number", 1)
	participant.Set("status", "active")
	participant.Set("role_key", "seer")
	participant.Set("outcome", "unset")
	participant.Set("joined_at", time.Now().UTC())
	if err := app.Save(participant); err != nil {
		t.Fatal(err)
	}
	return app, game, profile, participant, definition
}

func assertAbilityError(t *testing.T, err error, code string) {
	t.Helper()
	appError, ok := err.(result.AppError)
	if !ok || appError.Code != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}
