// Package testutil provides project-scoped helpers for host tests.
package testutil

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"

	_ "github.com/jvanspijk/SocialGamesHoster/Host/migrations"
)

// NewPocketBaseApp creates a migrated PocketBase test app and registers its cleanup.
func NewPocketBaseApp(t testing.TB) *tests.TestApp {
	t.Helper()

	app, err := tests.NewTestAppWithConfig(core.BaseAppConfig{
		DataDir:       t.TempDir(),
		EncryptionEnv: "sgh_test_encryption",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)

	return app
}
