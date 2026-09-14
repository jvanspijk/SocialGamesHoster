package migrations

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/fixtures"
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
		"announcement_attachments",
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
	attachments, err := app.FindCollectionByNameOrId("announcement_attachments")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"game", "announcement", "creator", "kind", "file", "mime_type", "checksum", "storage_state"} {
		if attachments.Fields.GetByName(field) == nil {
			t.Fatalf("announcement attachment field %q is missing", field)
		}
	}
	attachmentFile, ok := attachments.Fields.GetByName("file").(*core.FileField)
	if !ok || attachmentFile.MaxSize != 64<<20 {
		t.Fatalf("announcement attachment file limit = %d, want %d", attachmentFile.MaxSize, 64<<20)
	}
	rulesets, err := app.FindCollectionByNameOrId("rulesets")
	if err != nil {
		t.Fatal(err)
	}
	if rulesets.Fields.GetByName("latest_saved_version") == nil {
		t.Fatal("latest saved ruleset relation is missing")
	}
	rulesetVersions, err := app.FindCollectionByNameOrId("ruleset_versions")
	if err != nil {
		t.Fatal(err)
	}
	if rulesetVersions.Fields.GetByName("validation_report") == nil {
		t.Fatal("ruleset validation report field is missing")
	}
	rulesetAssets, err := app.FindCollectionByNameOrId("ruleset_assets")
	if err != nil {
		t.Fatal(err)
	}
	if rulesetAssets.Fields.GetByName("display_name") == nil {
		t.Fatal("ruleset asset display name field is missing")
	}
	assetFile, ok := rulesetAssets.Fields.GetByName("file").(*core.FileField)
	if !ok || assetFile.MaxSize != 64<<20 {
		t.Fatalf("ruleset asset file limit = %d, want %d", assetFile.MaxSize, 64<<20)
	}
	profiles, err := app.FindCollectionByNameOrId("player_profiles")
	if err != nil {
		t.Fatal(err)
	}
	avatarFile, ok := profiles.Fields.GetByName("avatar").(*core.FileField)
	if !ok || avatarFile.MaxSize != 64<<20 {
		t.Fatalf("profile avatar file limit = %d, want %d", avatarFile.MaxSize, 64<<20)
	}
	rulesetAssetChanges, err := app.FindCollectionByNameOrId("ruleset_asset_changes")
	if err != nil {
		t.Fatal(err)
	}
	changeFile, ok := rulesetAssetChanges.Fields.GetByName("file").(*core.FileField)
	if !ok || changeFile.MaxSize != 64<<20 {
		t.Fatalf("ruleset asset change file limit = %d, want %d", changeFile.MaxSize, 64<<20)
	}
	participants, err := app.FindCollectionByNameOrId("participants")
	if err != nil {
		t.Fatal(err)
	}
	if participants.Fields.GetByName("role_revision") == nil {
		t.Fatal("participant role revision is missing")
	}
	if participants.Fields.GetByName("player_number") == nil {
		t.Fatal("participant player number is missing")
	}
	if participants.Fields.GetByName("seat_number") != nil {
		t.Fatal("legacy participant seat number was not removed")
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

func TestPlayerNumberMigrationPreservesExistingValues(t *testing.T) {
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir(), EncryptionEnv: "sgh_test_encryption"})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}

	if err := renameParticipantNumberField(app, "player_number", "seat_number"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.DB().NewQuery(`
		INSERT INTO participants (id, game, profile, display_name_snapshot, seat_number, status, outcome)
		VALUES ('legacyplayer01', '', '', 'Legacy player', 7, 'active', 'unset')
	`).Execute(); err != nil {
		t.Fatal(err)
	}

	if err := renameParticipantNumberField(app, "seat_number", "player_number"); err != nil {
		t.Fatal(err)
	}
	var number int
	if err := app.DB().NewQuery("SELECT player_number FROM participants WHERE id = 'legacyplayer01'").Row(&number); err != nil {
		t.Fatal(err)
	}
	if number != 7 {
		t.Fatalf("player number = %d, want 7", number)
	}
	participants, err := app.FindCollectionByNameOrId("participants")
	if err != nil {
		t.Fatal(err)
	}
	if participants.Fields.GetByName("player_number") == nil || participants.Fields.GetByName("seat_number") != nil {
		t.Fatal("participant field was not renamed to player_number")
	}
	var indexSQL string
	if err := app.DB().NewQuery("SELECT sql FROM sqlite_master WHERE type = 'index' AND name = 'idx_participants_game_seat'").Row(&indexSQL); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(indexSQL, "player_number") {
		t.Fatalf("participant number index was not updated: %s", indexSQL)
	}
}

func TestUpgradeFixtureKeepsExistingOwnerCredentials(t *testing.T) {
	dataDir := t.TempDir()
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: dataDir, EncryptionEnv: "sgh_test_encryption"})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	masters, err := app.FindCollectionByNameOrId("game_masters")
	if err != nil {
		t.Fatal(err)
	}
	owner := core.NewRecord(masters)
	owner.Set("username", "upgrade-owner")
	owner.Set("display_name", "Upgrade Owner")
	owner.Set("is_owner", true)
	owner.Set("active", true)
	owner.SetPassword("original-password")
	if err := app.Save(owner); err != nil {
		t.Fatal(err)
	}
	if err := fixtures.Seed(app, owner.Id); err != nil {
		t.Fatal(err)
	}
	if err := app.ResetBootstrapState(); err != nil {
		t.Fatal(err)
	}

	upgraded := core.NewBaseApp(core.BaseAppConfig{DataDir: dataDir, EncryptionEnv: "sgh_test_encryption"})
	if err := upgraded.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = upgraded.ResetBootstrapState() })
	if err := upgraded.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	restored, err := upgraded.FindFirstRecordByData("game_masters", "username", "upgrade-owner")
	if err != nil {
		t.Fatal(err)
	}
	if !restored.ValidatePassword("original-password") {
		t.Fatal("original owner password no longer authenticates after upgrade")
	}
	if count, err := upgraded.CountRecords("rulesets"); err != nil || count == 0 {
		t.Fatalf("ruleset fixture was not preserved: %d, %v", count, err)
	}
}
