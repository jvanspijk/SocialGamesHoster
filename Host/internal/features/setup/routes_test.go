package setup

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jvanspijk/SocialGamesHoster/Host/migrations"
	"github.com/pocketbase/pocketbase/core"
)

func TestReplaceOwnersPreservesNonOwnersAndReassignsRelations(t *testing.T) {
	app := recoveryTestApp(t)
	masters, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		t.Fatal(err)
	}
	owner := newMaster(t, app, masters, "former-owner", true)
	nonOwner := newMaster(t, app, masters, "game-master", false)
	profiles, err := app.FindCollectionByNameOrId("player_profiles")
	if err != nil {
		t.Fatal(err)
	}
	profile := core.NewRecord(profiles)
	profile.Set("display_name", "Player One")
	profile.Set("normalized_name", "player one")
	profile.Set("active", true)
	profile.Set("approved_by", owner.Id)
	profile.SetPassword("profile-secret")
	if err := app.Save(profile); err != nil {
		t.Fatal(err)
	}

	replacement, err := replaceOwners(app, "replacement", "Replacement", "new-secret", "test")
	if err != nil {
		t.Fatal(err)
	}
	if !replacement.GetBool("is_owner") || !replacement.ValidatePassword("new-secret") {
		t.Fatal("replacement owner was not created")
	}
	if _, err := app.FindRecordById(collectionName, owner.Id); err == nil {
		t.Fatal("former owner still exists")
	}
	if _, err := app.FindRecordById(collectionName, nonOwner.Id); err != nil {
		t.Fatalf("non-owner was removed: %v", err)
	}
	updated, err := app.FindRecordById("player_profiles", profile.Id)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetString("approved_by") != replacement.Id {
		t.Fatalf("approved_by = %q, want %q", updated.GetString("approved_by"), replacement.Id)
	}
}

func TestReplaceOwnersRollsBackWhenReplacementCannotBeCreated(t *testing.T) {
	app := recoveryTestApp(t)
	masters, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		t.Fatal(err)
	}
	owner := newMaster(t, app, masters, "former-owner", true)
	if _, err := replaceOwners(app, "former-owner", "Replacement", "new-secret", "test"); err == nil {
		t.Fatal("expected duplicate username to fail")
	}
	if _, err := app.FindRecordById(collectionName, owner.Id); err != nil {
		t.Fatalf("owner changed after failed recovery: %v", err)
	}
	count, err := app.CountRecords(collectionName)
	if err != nil || count != 1 {
		t.Fatalf("master count after failed recovery = %d, %v", count, err)
	}
}

func recoveryTestApp(t *testing.T) core.App {
	t.Helper()
	dir := t.TempDir()
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: filepath.Join(dir, "pb_data"), EncryptionEnv: "sgh_test_encryption"})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState(); _ = os.RemoveAll(dir) })
	return app
}

func newMaster(t *testing.T, app core.App, collection *core.Collection, username string, owner bool) *core.Record {
	t.Helper()
	record := core.NewRecord(collection)
	record.Set("username", username)
	record.Set("display_name", username)
	record.Set("is_owner", owner)
	record.Set("active", true)
	record.SetPassword("secret")
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}
