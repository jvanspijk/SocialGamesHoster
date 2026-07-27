package migrations

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestInitialMigrationUpAndDown(t *testing.T) {
	baseDir := filepath.Join(".testdata")
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		t.Fatal(err)
	}
	dataDir, err := os.MkdirTemp(baseDir, "migration-*")
	if err != nil {
		t.Fatal(err)
	}

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

	var indexName string
	if err := app.DB().NewQuery("SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'idx_games_single_live'").Row(&indexName); err != nil {
		t.Fatalf("single-live-game index missing: %v", err)
	}

	runner := core.NewMigrationsRunner(app, core.AppMigrations)
	if _, err := runner.Down(3); err != nil {
		t.Fatal(err)
	}
	if _, err := app.FindCollectionByNameOrId("games"); err == nil || !errors.Is(err, os.ErrNotExist) {
		// PocketBase returns sql.ErrNoRows, but the important invariant is that
		// the collection is no longer queryable.
		if err == nil {
			t.Fatal("games collection still exists after down migration")
		}
	}
}
