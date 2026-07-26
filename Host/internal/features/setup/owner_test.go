package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	_ "github.com/jvanspijk/SocialGamesHoster/Host/migrations"
)

func TestOwnerRecordCanBeCreated(t *testing.T) {
	if err := os.MkdirAll(".testdata", 0o700); err != nil {
		t.Fatal(err)
	}
	dataDir, err := os.MkdirTemp(".testdata", "owner-*")
	if err != nil {
		t.Fatal(err)
	}
	dataDir, err = filepath.Abs(dataDir)
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
	collection, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("username", "owner")
	record.Set("display_name", "Test Owner")
	record.Set("is_owner", true)
	record.Set("active", true)
	record.SetPassword("correct-horse-battery")
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	if !record.ValidatePassword("correct-horse-battery") {
		t.Fatal("saved password does not validate")
	}
}
