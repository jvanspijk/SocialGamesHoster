package rulesets

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"

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

func TestReplacementAuditFailureLeavesOldReadyAssetAndCleansStaging(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)
	masters, _ := app.FindCollectionByNameOrId("game_masters")
	master := core.NewRecord(masters)
	master.Set("username", "replace-host")
	master.Set("display_name", "Replace host")
	master.Set("active", true)
	master.SetPassword("secret-password")
	if err := app.Save(master); err != nil {
		t.Fatal(err)
	}
	rulesets, _ := app.FindCollectionByNameOrId("rulesets")
	logical := core.NewRecord(rulesets)
	logical.Set("slug", "replace-test")
	logical.Set("name", "Replace")
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
	old := core.NewRecord(assets)
	old.Set("ruleset_version", version.Id)
	old.Set("asset_key", "cover")
	old.Set("kind", "image")
	old.Set("mime_type", "image/png")
	old.Set("checksum", "old-checksum")
	old.Set("metadata", map[string]any{"old": true})
	old.Set("storage_state", "ready")
	if err := app.Save(old); err != nil {
		t.Fatal(err)
	}
	staged := core.NewRecord(assets)
	staged.Set("ruleset_version", version.Id)
	staged.Set("asset_key", "staging-new")
	staged.Set("kind", "image")
	staged.Set("mime_type", "image/png")
	staged.Set("checksum", "new-checksum")
	staged.Set("metadata", map[string]any{"new": true})
	staged.Set("storage_state", "staging")
	if err := app.Save(staged); err != nil {
		t.Fatal(err)
	}
	previous := auditRecord
	auditRecord = func(core.App, *core.Record, string, string, string, string, any, any) error {
		return errors.New("audit unavailable")
	}
	t.Cleanup(func() { auditRecord = previous })
	err := app.RunInTransaction(func(tx core.App) error {
		current, err := tx.FindRecordById("ruleset_assets", staged.Id)
		if err != nil {
			return err
		}
		prior, err := tx.FindRecordById("ruleset_assets", old.Id)
		if err != nil {
			return err
		}
		prior.Set("asset_key", "replaced-"+prior.Id)
		prior.Set("storage_state", "staging")
		if err := tx.Save(prior); err != nil {
			return err
		}
		current.Set("asset_key", "cover")
		current.Set("storage_state", "ready")
		if err := tx.Save(current); err != nil {
			return err
		}
		return auditRecord(tx, nil, "", "ruleset.asset_uploaded", "ruleset_asset", current.Id, nil, nil)
	})
	if err == nil {
		t.Fatal("expected audit failure")
	}
	// This mirrors uploadAsset compensation: discard only the staged replacement.
	if err := app.Delete(staged); err != nil {
		t.Fatal(err)
	}
	ready, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && storage_state = 'ready'", "", 10, 0, map[string]any{"version": version.Id})
	if err != nil || len(ready) != 1 || ready[0].Id != old.Id || ready[0].GetString("asset_key") != "cover" || ready[0].GetString("checksum") != "old-checksum" {
		t.Fatalf("old asset was not preserved: %#v %v", ready, err)
	}
	stagedRecords, err := app.FindRecordsByFilter("ruleset_assets", "id = {:id}", "", 1, 0, map[string]any{"id": staged.Id})
	if err != nil || len(stagedRecords) != 0 {
		t.Fatalf("staging replacement remained: %d %v", len(stagedRecords), err)
	}
}

func TestSaveRulesetInvalidSuccessorStaysUnselectable(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)
	masters, _ := app.FindCollectionByNameOrId("game_masters")
	master := core.NewRecord(masters)
	master.Set("username", "invalid-host")
	master.Set("display_name", "Invalid host")
	master.Set("active", true)
	master.SetPassword("secret-password")
	if err := app.Save(master); err != nil {
		t.Fatal(err)
	}
	rulesets, _ := app.FindCollectionByNameOrId("rulesets")
	logical := core.NewRecord(rulesets)
	logical.Set("slug", "invalid-successor")
	logical.Set("name", "Original")
	logical.Set("created_by", master.Id)
	if err := app.Save(logical); err != nil {
		t.Fatal(err)
	}
	versions, _ := app.FindCollectionByNameOrId("ruleset_versions")
	definition := testDefinition()
	source := core.NewRecord(versions)
	source.Set("ruleset", logical.Id)
	source.Set("version_number", 1)
	source.Set("state", "published")
	source.Set("schema_version", 1)
	source.Set("definition", definition)
	source.Set("created_by", master.Id)
	if err := app.Save(source); err != nil {
		t.Fatal(err)
	}
	logical.Set("latest_published_version", source.Id)
	if err := app.Save(logical); err != nil {
		t.Fatal(err)
	}
	assets, _ := app.FindCollectionByNameOrId("ruleset_assets")
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScL0WQAAAABJRU5ErkJggg==")
	file, err := filesystem.NewFileFromBytes(png, "cover.png")
	if err != nil {
		t.Fatal(err)
	}
	sourceAsset := core.NewRecord(assets)
	sourceAsset.Set("ruleset_version", source.Id)
	sourceAsset.Set("asset_key", "cover")
	sourceAsset.Set("kind", "image")
	sourceAsset.Set("file", file)
	sourceAsset.Set("mime_type", "image/png")
	sourceAsset.Set("checksum", "source-cover")
	sourceAsset.Set("metadata", map[string]any{})
	sourceAsset.Set("storage_state", "ready")
	if err := app.Save(sourceAsset); err != nil {
		t.Fatal(err)
	}
	invalid := definition
	invalid.Metadata.MinPlayers = 0
	body, _ := json.Marshal(saveRulesetRequest{Definition: invalid})
	req := httptest.NewRequest(http.MethodPost, "/api/app/v1/rulesets/"+logical.Id+"/save", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", logical.Id)
	event := &core.RequestEvent{App: app, Auth: master}
	event.Request = req
	event.Response = httptest.NewRecorder()
	if err := saveRuleset(event); err != nil {
		t.Fatal(err)
	}
	updated, err := app.FindRecordById("rulesets", logical.Id)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetString("latest_published_version") != "" {
		t.Fatalf("invalid successor remained available for new games: %q", updated.GetString("latest_published_version"))
	}
	if updated.GetString("latest_saved_version") == "" {
		t.Fatal("invalid successor was not recorded as the latest saved ruleset")
	}
	latestSaved, err := app.FindRecordById("ruleset_versions", updated.GetString("latest_saved_version"))
	if err != nil {
		t.Fatal(err)
	}
	var savedReport ValidationReport
	if err := latestSaved.UnmarshalJSONField("validation_report", &savedReport); err != nil {
		t.Fatal(err)
	}
	if savedReport.Valid() {
		t.Fatal("invalid successor did not persist its validation report")
	}
	drafts, err := app.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset} && state = 'draft'", "", 10, 0, map[string]any{"ruleset": logical.Id})
	if err != nil || len(drafts) != 1 {
		t.Fatalf("draft successor: %d %v", len(drafts), err)
	}
	publicAssets, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && storage_state = 'ready'", "", 10, 0, map[string]any{"version": drafts[0].Id})
	if err != nil || len(publicAssets) != 0 {
		t.Fatalf("invalid successor exposed assets: %d %v", len(publicAssets), err)
	}
	validBody, _ := json.Marshal(saveRulesetRequest{Definition: definition})
	validRequest := httptest.NewRequest(http.MethodPost, "/api/app/v1/rulesets/"+logical.Id+"/save", bytes.NewReader(validBody))
	validRequest.Header.Set("Content-Type", "application/json")
	validRequest.SetPathValue("id", logical.Id)
	validEvent := &core.RequestEvent{App: app, Auth: master}
	validEvent.Request = validRequest
	validEvent.Response = httptest.NewRecorder()
	if err := saveRuleset(validEvent); err != nil {
		t.Fatal(err)
	}
	updated, err = app.FindRecordById("rulesets", logical.Id)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetString("latest_published_version") != drafts[0].Id {
		t.Fatalf("valid resave did not publish successor")
	}
	published, err := app.FindRecordById("ruleset_versions", drafts[0].Id)
	if err != nil || published.GetString("state") != "published" {
		t.Fatalf("successor not published: %#v %v", published, err)
	}
	readyAssets, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && storage_state = 'ready'", "", 10, 0, map[string]any{"version": drafts[0].Id})
	if err != nil || len(readyAssets) != 1 || readyAssets[0].GetString("asset_key") != "cover" || readyAssets[0].GetString("file") == "" {
		t.Fatalf("resaved asset not public: %#v %v", readyAssets, err)
	}
}
