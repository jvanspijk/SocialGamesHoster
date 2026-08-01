package testutil_test

import (
	"testing"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
)

func TestNewPocketBaseAppAppliesProjectMigrations(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)

	if _, err := app.FindCollectionByNameOrId("games"); err != nil {
		t.Fatalf("project migrations were not applied: %v", err)
	}
}
