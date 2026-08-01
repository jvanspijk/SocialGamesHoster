package migrations

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestInitialMigrationUp(t *testing.T) {
	baseDir := filepath.Join(".testdata")
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		t.Fatal(err)
	}
	dataDir, err := os.MkdirTemp(baseDir, "migration-*")
	if err != nil {
		t.Fatal(err)
	}

	// This test deliberately exercises the raw migration lifecycle, so it does
	// not use the migrated PocketBase test-app helper.
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: dataDir, EncryptionEnv: "sgh_test_encryption"})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		app.ResetBootstrapState()
		if err := os.RemoveAll(dataDir); err != nil {
			t.Errorf("remove test data: %v", err)
		}
	})

	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}

	names := []string{
		"game_masters",
		"host_settings",
		"player_profiles",
		"profile_requests",
		"rulesets",
		"ruleset_versions",
		"ruleset_assets",
		"games",
		"participants",
		"chat_rooms",
		"chat_memberships",
		"chat_messages",
		"attention_items",
		"attention_receipts",
		"ability_choices",
		"achievement_awards",
		"game_audit",
	}
	for _, name := range names {
		collection, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			t.Fatalf("missing collection %s: %v", name, err)
		}
		if collection.ListRule != nil || collection.ViewRule != nil || collection.CreateRule != nil || collection.UpdateRule != nil || collection.DeleteRule != nil {
			t.Fatalf("collection %s exposes a generic API rule", name)
		}
	}
	awards, err := app.FindCollectionByNameOrId("achievement_awards")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"points_snapshot", "hidden_until_game_completed"} {
		if awards.Fields.GetByName(field) == nil {
			t.Fatalf("achievement award contract field %q is missing", field)
		}
	}
	games, err := app.FindCollectionByNameOrId("games")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"roles_visible", "role_visibility_revision", "completion_previous_status"} {
		if games.Fields.GetByName(field) == nil {
			t.Fatalf("game contract field %q is missing", field)
		}
	}
	if games.Fields.GetByName("ability_phase_locked_at") == nil {
		t.Fatal("ability phase lock field is missing")
	}
	if games.Fields.GetByName("ability_phase_instance") == nil {
		t.Fatal("ability phase instance field is missing")
	}
	participants, err := app.FindCollectionByNameOrId("participants")
	if err != nil {
		t.Fatal(err)
	}
	if participants.Fields.GetByName("role_revision") == nil {
		t.Fatal("participant role revision is missing")
	}
	rooms, err := app.FindCollectionByNameOrId("chat_rooms")
	if err != nil {
		t.Fatal(err)
	}
	if rooms.Fields.GetByName("players_can_post") == nil {
		t.Fatal("room posting contract is missing")
	}
	kind, ok := rooms.Fields.GetByName("kind").(*core.SelectField)
	if !ok || !slices.Contains(kind.Values, "custom") {
		t.Fatal("custom ruleset chat room kind is missing")
	}

	var indexName string
	if err := app.DB().NewQuery("SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'idx_games_single_live'").Row(&indexName); err != nil {
		t.Fatalf("single-live-game index missing: %v", err)
	}

}
