package rulesets

import (
	"errors"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
)

func TestReadyFinalizationRollsBackWhenAuditWriteFails(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)
	masters, _ := app.FindCollectionByNameOrId("game_masters")
	master := core.NewRecord(masters)
	master.Set("username", "staging-host")
	master.Set("display_name", "Staging host")
	master.Set("active", true)
	master.SetPassword("secret-password")
	if err := app.Save(master); err != nil {
		t.Fatal(err)
	}
	rulesets, _ := app.FindCollectionByNameOrId("rulesets")
	logical := core.NewRecord(rulesets)
	logical.Set("slug", "staging-test")
	logical.Set("name", "Staging test")
	logical.Set("created_by", master.Id)
	if err := app.Save(logical); err != nil {
		t.Fatal(err)
	}
	versions, _ := app.FindCollectionByNameOrId("ruleset_versions")
	version := core.NewRecord(versions)
	version.Set("ruleset", logical.Id)
	version.Set("version_number", 1)
	version.Set("state", "draft")
	version.Set("schema_version", 1)
	version.Set("definition", map[string]any{"schemaVersion": 1})
	version.Set("created_by", master.Id)
	if err := app.Save(version); err != nil {
		t.Fatal(err)
	}
	assets, _ := app.FindCollectionByNameOrId("ruleset_assets")
	asset := core.NewRecord(assets)
	asset.Set("ruleset_version", version.Id)
	asset.Set("asset_key", "staged")
	asset.Set("kind", "image")
	asset.Set("mime_type", "image/png")
	asset.Set("checksum", "abc")
	asset.Set("storage_state", "staging")
	if err := app.Save(asset); err != nil {
		t.Fatal(err)
	}
	previous := auditRecord
	auditRecord = func(core.App, *core.Record, string, string, string, string, any, any) error {
		return errors.New("audit unavailable")
	}
	t.Cleanup(func() { auditRecord = previous })
	err := app.RunInTransaction(func(tx core.App) error {
		if err := markVersionAssetsReady(tx, version.Id); err != nil {
			return err
		}
		return auditRecord(tx, nil, "", "ruleset.draft_created", "ruleset_version", version.Id, nil, nil)
	})
	if err == nil {
		t.Fatal("expected audit failure")
	}
	current, err := app.FindRecordById("ruleset_assets", asset.Id)
	if err != nil || current.GetString("storage_state") != "staging" {
		t.Fatalf("ready state persisted: %#v %v", current, err)
	}
	audits, err := app.FindAllRecords("game_audit")
	if err != nil || len(audits) != 0 {
		t.Fatalf("audit persisted: %d %v", len(audits), err)
	}
	keys, err := versionAssetKeys(app, version.Id)
	if err != nil || len(keys) != 0 {
		t.Fatalf("staged asset leaked into public validation keys: %#v %v", keys, err)
	}
}
