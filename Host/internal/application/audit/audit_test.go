package audit_test

import (
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
)

func TestRecordSnapshotsActorAndEntryFields(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)
	gameMaster := saveActor(t, app, actorauth.GameMastersCollection, "Game Master")
	player := saveActor(t, app, actorauth.PlayerProfilesCollection, "Player One")
	game := saveGame(t, app, gameMaster)

	tests := []struct {
		name       string
		actor      *core.Record
		actorType  string
		actorID    string
		label      string
		gameID     string
		action     string
		targetType string
		targetID   string
	}{
		{name: "system host action", actorType: "system", label: "System", action: "host.settings_updated", targetType: "host_settings", targetID: "settings-1"},
		{name: "game master game action", actor: gameMaster, actorType: "game_master", actorID: gameMaster.Id, label: "Game Master", gameID: game.Id, action: "game.created", targetType: "game", targetID: game.Id},
		{name: "player game action", actor: player, actorType: "player", actorID: player.Id, label: "Player One", gameID: game.Id, action: "participant.joined", targetType: "participant", targetID: "participant-1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := applicationaudit.Record(app, test.actor, test.gameID, test.action, test.targetType, test.targetID, map[string]any{"safe": true}, "request-123"); err != nil {
				t.Fatal(err)
			}
			if test.actor != nil {
				test.actor.Set("display_name", "Changed after audit")
				if err := app.Save(test.actor); err != nil {
					t.Fatal(err)
				}
			}
			record := latestAudit(t, app, test.action)
			if got := record.GetString("actor_type"); got != test.actorType {
				t.Errorf("actor_type = %q, want %q", got, test.actorType)
			}
			if got := record.GetString("actor_id"); got != test.actorID {
				t.Errorf("actor_id = %q, want %q", got, test.actorID)
			}
			if got := record.GetString("actor_label"); got != test.label {
				t.Errorf("actor_label = %q, want %q", got, test.label)
			}
			if got := record.GetString("action"); got != test.action {
				t.Errorf("action = %q, want %q", got, test.action)
			}
			if got := record.GetString("target_type"); got != test.targetType {
				t.Errorf("target_type = %q, want %q", got, test.targetType)
			}
			if got := record.GetString("target_id"); got != test.targetID {
				t.Errorf("target_id = %q, want %q", got, test.targetID)
			}
			if got := record.GetString("request_id"); got != "request-123" {
				t.Errorf("request_id = %q", got)
			}
			if got := record.GetString("detail"); got != `{"safe":true}` {
				t.Errorf("detail = %q", got)
			}
		})
	}
}

func TestRecordUsesTransactionApp(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)
	if err := app.RunInTransaction(func(txApp core.App) error {
		return applicationaudit.Record(txApp, nil, "", "game.transaction_written", "game", "game-1", nil, "request-transaction")
	}); err != nil {
		t.Fatal(err)
	}
	record := latestAudit(t, app, "game.transaction_written")
	if got := record.GetString("action"); got != "game.transaction_written" {
		t.Errorf("action = %q", got)
	}
}

func saveActor(t *testing.T, app core.App, collectionName, label string) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("display_name", label)
	record.Set("active", true)
	if collectionName == actorauth.GameMastersCollection {
		record.Set("username", label)
	}
	if collectionName == actorauth.PlayerProfilesCollection {
		record.Set("normalized_name", label)
	}
	record.SetPassword("secret-password")
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}

func saveGame(t *testing.T, app core.App, gameMaster *core.Record) *core.Record {
	t.Helper()
	rulesetCollection, err := app.FindCollectionByNameOrId("rulesets")
	if err != nil {
		t.Fatal(err)
	}
	ruleset := core.NewRecord(rulesetCollection)
	ruleset.Set("slug", "audit-test")
	ruleset.Set("name", "Audit test")
	ruleset.Set("created_by", gameMaster.Id)
	if err := app.Save(ruleset); err != nil {
		t.Fatal(err)
	}
	versionCollection, err := app.FindCollectionByNameOrId("ruleset_versions")
	if err != nil {
		t.Fatal(err)
	}
	definition := rulesets.DefinitionV1{SchemaVersion: 1}
	version := core.NewRecord(versionCollection)
	version.Set("ruleset", ruleset.Id)
	version.Set("version_number", 1)
	version.Set("state", "published")
	version.Set("schema_version", 1)
	version.Set("definition", definition)
	version.Set("created_by", gameMaster.Id)
	version.Set("published_by", gameMaster.Id)
	version.Set("published_at", time.Now().UTC())
	if err := app.Save(version); err != nil {
		t.Fatal(err)
	}
	gameCollection, err := app.FindCollectionByNameOrId("games")
	if err != nil {
		t.Fatal(err)
	}
	game := core.NewRecord(gameCollection)
	game.Set("name", "Audit test")
	game.Set("status", "running")
	game.Set("ruleset_version", version.Id)
	game.Set("ruleset_snapshot", definition)
	game.Set("timer_state", "inactive")
	game.Set("created_by", gameMaster.Id)
	if err := app.Save(game); err != nil {
		t.Fatal(err)
	}
	return game
}

func latestAudit(t *testing.T, app core.App, action string) *core.Record {
	t.Helper()
	records, err := app.FindRecordsByFilter("game_audit", "action = {:action}", "-created,-id", 1, 0, dbx.Params{"action": action})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("audit record count = %d, want 1", len(records))
	}
	return records[0]
}
