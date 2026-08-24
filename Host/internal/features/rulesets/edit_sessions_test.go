package rulesets

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
)

func TestScanAssetUsagesIncludesDirectAndIndirectConsumers(t *testing.T) {
	definition := testDefinition()
	definition.Metadata.CoverAssetKey = "shared-image"
	definition.Teams[0].ImageAssetKey = "shared-image"
	definition.Roles[0].ImageAssetKey = "shared-image"
	definition.AudioCues = []AudioCue{{ID: "cue_1", Name: "Bell", AssetKey: "bell", DefaultAudience: "all"}}
	definition.Phases = []Phase{{ID: "phase_1", Name: "Night", AudioCueID: "cue_1"}}

	images := scanAssetUsages(definition, "shared-image")
	if len(images) != 3 || images[0].Label != "Ruleset cover" || images[1].Label != "Team · Town" || images[2].Label != "Role · Villager" {
		t.Fatalf("unexpected direct usages: %#v", images)
	}
	audio := scanAssetUsages(definition, "bell")
	if len(audio) != 1 || audio[0].Label != "Audio cue · Bell → Phase · Night" || audio[0].Section != "phases" {
		t.Fatalf("unexpected indirect usage: %#v", audio)
	}
}

func TestOwnedEditSessionRejectsAnotherCreatorAndExpiry(t *testing.T) {
	app, logical, _, owner, other := editSessionFixture(t)
	session := newTestEditSession(t, app, logical, owner, time.Now().UTC().Add(time.Hour))
	event := editSessionEvent(app, other, logical.Id, session.Id)
	if _, err := ownedEditSession(event); err == nil || err.(result.AppError).Status != http.StatusForbidden {
		t.Fatalf("expected ownership rejection, got %v", err)
	}

	session.Set("expires_at", time.Now().UTC().Add(-time.Minute))
	if err := app.Save(session); err != nil {
		t.Fatal(err)
	}
	event = editSessionEvent(app, owner, logical.Id, session.Id)
	if _, err := ownedEditSession(event); err == nil || err.(result.AppError).Status != http.StatusGone {
		t.Fatalf("expected expiry rejection, got %v", err)
	}
	if _, err := app.FindRecordById("ruleset_edit_sessions", session.Id); err == nil {
		t.Fatal("expired session was not cleaned")
	}
}

func TestValidateRulesetRefreshesEditSessionExpiry(t *testing.T) {
	app, logical, _, owner, _ := editSessionFixture(t)
	initialExpiry := time.Now().UTC().Add(time.Hour)
	session := newTestEditSession(t, app, logical, owner, initialExpiry)
	body, _ := json.Marshal(saveRulesetRequest{Definition: testDefinition(), SessionID: session.Id})
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("id", logical.Id)
	event := &core.RequestEvent{App: app, Auth: owner}
	event.Request = request
	event.Response = httptest.NewRecorder()

	if err := validateRuleset(event); err != nil {
		t.Fatal(err)
	}
	refreshed, err := app.FindRecordById("ruleset_edit_sessions", session.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !refreshed.GetDateTime("expires_at").Time().After(initialExpiry) {
		t.Fatalf("validation did not refresh expiry: %s", refreshed.GetDateTime("expires_at"))
	}
}

func TestOpenEditSessionResumesAndDiscardRemovesStagedChanges(t *testing.T) {
	app, logical, _, owner, _ := editSessionFixture(t)
	open := func() string {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.SetPathValue("id", logical.Id)
		event := &core.RequestEvent{App: app, Auth: owner}
		event.Request = request
		event.Response = httptest.NewRecorder()
		if err := openEditSession(event); err != nil {
			t.Fatal(err)
		}
		var projection struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(event.Response.(*httptest.ResponseRecorder).Body.Bytes(), &projection); err != nil {
			t.Fatal(err)
		}
		return projection.ID
	}
	first := open()
	if second := open(); second != first {
		t.Fatalf("session was not resumed: %q != %q", second, first)
	}
	changes, _ := app.FindCollectionByNameOrId("ruleset_asset_changes")
	change := core.NewRecord(changes)
	change.Set("session", first)
	change.Set("asset_key", "asset_added")
	change.Set("operation", "add")
	change.Set("kind", "image")
	change.Set("display_name", "Staged image")
	if err := app.Save(change); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/", nil)
	request.SetPathValue("id", logical.Id)
	request.SetPathValue("sessionId", first)
	event := &core.RequestEvent{App: app, Auth: owner}
	event.Request = request
	event.Response = httptest.NewRecorder()
	if err := discardEditSession(event); err != nil {
		t.Fatal(err)
	}
	if count, _ := app.CountRecords("ruleset_asset_changes", dbx.HashExp{"session": first}); count != 0 {
		t.Fatalf("discard left %d staged changes", count)
	}
	if _, err := app.FindRecordById("ruleset_edit_sessions", first); err == nil {
		t.Fatal("discard left the edit session")
	}
}

func TestSessionSaveReplacesEverywhereWithoutChangingHistoricalAsset(t *testing.T) {
	app, logical, base, owner, _ := editSessionFixture(t)
	oldPNG := tinyPNG(t)
	oldAsset := newVersionAsset(t, app, base.Id, "cover", "Old cover", oldPNG)
	definition := testDefinition()
	definition.Metadata.CoverAssetKey = "cover"
	base.Set("definition", definition)
	if err := app.Save(base); err != nil {
		t.Fatal(err)
	}
	session := newTestEditSession(t, app, logical, owner, time.Now().UTC().Add(time.Hour))
	changes, _ := app.FindCollectionByNameOrId("ruleset_asset_changes")
	change := core.NewRecord(changes)
	change.Set("session", session.Id)
	change.Set("asset_key", "cover")
	change.Set("operation", "replace")
	change.Set("kind", "image")
	change.Set("display_name", "New cover")
	change.Set("accessibility_text", "A new cover")
	change.Set("mime_type", "image/png")
	change.Set("checksum", "new-checksum")
	change.Set("metadata", map[string]any{"width": 1, "height": 1})
	file, _ := filesystem.NewFileFromBytes(oldPNG, "new-cover.png")
	change.Set("file", file)
	if err := app.Save(change); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(saveRulesetRequest{Definition: definition, SessionID: session.Id})
	request := httptest.NewRequest(http.MethodPost, "/api/app/v1/rulesets/"+logical.Id+"/save", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("id", logical.Id)
	event := &core.RequestEvent{App: app, Auth: owner}
	event.Request = request
	event.Response = httptest.NewRecorder()
	if err := saveRuleset(event); err != nil {
		t.Fatal(err)
	}
	updated, _ := app.FindRecordById("rulesets", logical.Id)
	if updated.GetString("latest_saved_version") == base.Id || updated.GetString("latest_published_version") == "" {
		t.Fatalf("save did not create and select a new valid revision: %#v", updated)
	}
	historical, err := app.FindRecordById("ruleset_assets", oldAsset.Id)
	if err != nil || historical.GetString("display_name") != "Old cover" {
		t.Fatalf("historical asset changed: %#v %v", historical, err)
	}
	current, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && asset_key = 'cover'", "", 1, 0, map[string]any{"version": updated.GetString("latest_saved_version")})
	if err != nil || len(current) != 1 || current[0].GetString("display_name") != "New cover" || current[0].GetString("accessibility_text") != "A new cover" {
		t.Fatalf("replacement was not materialized: %#v %v", current, err)
	}
	if _, err := app.FindRecordById("ruleset_edit_sessions", session.Id); err == nil {
		t.Fatal("successful Save did not consume the edit session")
	}
}

func TestSessionSaveFailureKeepsSessionAndCanRetry(t *testing.T) {
	app, logical, base, owner, _ := editSessionFixture(t)
	session := newTestEditSession(t, app, logical, owner, time.Now().UTC().Add(time.Hour))
	body, _ := json.Marshal(saveRulesetRequest{Definition: testDefinition(), SessionID: session.Id})
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("id", logical.Id)
	event := &core.RequestEvent{App: app, Auth: owner}
	event.Request = request
	event.Response = httptest.NewRecorder()
	previous := auditRecord
	auditRecord = func(core.App, *core.Record, string, string, string, string, any, any) error {
		return errors.New("audit unavailable")
	}
	if err := saveRuleset(event); err != nil {
		t.Fatal(err)
	}
	auditRecord = previous
	t.Cleanup(func() { auditRecord = previous })
	current, _ := app.FindRecordById("rulesets", logical.Id)
	if current.GetString("latest_saved_version") != base.Id {
		t.Fatal("failed Save changed the saved revision")
	}
	if _, err := app.FindRecordById("ruleset_edit_sessions", session.Id); err != nil {
		t.Fatal("failed Save discarded the retryable session")
	}

	retryRequest := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	retryRequest.Header.Set("Content-Type", "application/json")
	retryRequest.SetPathValue("id", logical.Id)
	retry := &core.RequestEvent{App: app, Auth: owner}
	retry.Request = retryRequest
	retry.Response = httptest.NewRecorder()
	if err := saveRuleset(retry); err != nil {
		t.Fatal(err)
	}
	current, _ = app.FindRecordById("rulesets", logical.Id)
	if current.GetString("latest_saved_version") == base.Id {
		t.Fatal("retry did not save a successor revision")
	}
}

func TestDeleteRulesetFailureKeepsSessionAndSuccessfulRetryConsumesIt(t *testing.T) {
	app, logical, _, owner, _ := editSessionFixture(t)
	session := newTestEditSession(t, app, logical, owner, time.Now().UTC().Add(time.Hour))
	previous := auditRecord
	auditRecord = func(core.App, *core.Record, string, string, string, string, any, any) error {
		return errors.New("audit unavailable")
	}
	t.Cleanup(func() { auditRecord = previous })

	request := httptest.NewRequest(http.MethodDelete, "/", nil)
	request.SetPathValue("id", logical.Id)
	event := &core.RequestEvent{App: app, Auth: owner}
	event.Request = request
	event.Response = httptest.NewRecorder()
	if err := deleteRuleset(event); err != nil {
		t.Fatal(err)
	}
	if _, err := app.FindRecordById("rulesets", logical.Id); err != nil {
		t.Fatal("failed deletion removed the ruleset")
	}
	if _, err := app.FindRecordById("ruleset_edit_sessions", session.Id); err != nil {
		t.Fatal("failed deletion removed the edit session")
	}

	auditRecord = previous
	retryRequest := httptest.NewRequest(http.MethodDelete, "/", nil)
	retryRequest.SetPathValue("id", logical.Id)
	retry := &core.RequestEvent{App: app, Auth: owner}
	retry.Request = retryRequest
	retry.Response = httptest.NewRecorder()
	if err := deleteRuleset(retry); err != nil {
		t.Fatal(err)
	}
	if _, err := app.FindRecordById("rulesets", logical.Id); err == nil {
		t.Fatal("successful deletion left the ruleset")
	}
	if _, err := app.FindRecordById("ruleset_edit_sessions", session.Id); err == nil {
		t.Fatal("successful deletion left the edit session")
	}
}

func editSessionFixture(t *testing.T) (core.App, *core.Record, *core.Record, *core.Record, *core.Record) {
	t.Helper()
	app := testutil.NewPocketBaseApp(t)
	masters, _ := app.FindCollectionByNameOrId("game_masters")
	newMaster := func(username string) *core.Record {
		record := core.NewRecord(masters)
		record.Set("username", username)
		record.Set("display_name", username)
		record.Set("active", true)
		record.SetPassword("secret-password")
		if err := app.Save(record); err != nil {
			t.Fatal(err)
		}
		return record
	}
	owner := newMaster("media-owner")
	other := newMaster("media-other")
	rulesets, _ := app.FindCollectionByNameOrId("rulesets")
	logical := core.NewRecord(rulesets)
	logical.Set("slug", "media-session")
	logical.Set("name", "Media session")
	logical.Set("created_by", owner.Id)
	if err := app.Save(logical); err != nil {
		t.Fatal(err)
	}
	versions, _ := app.FindCollectionByNameOrId("ruleset_versions")
	base := core.NewRecord(versions)
	base.Set("ruleset", logical.Id)
	base.Set("version_number", 1)
	base.Set("state", "published")
	base.Set("schema_version", 1)
	base.Set("definition", testDefinition())
	base.Set("created_by", owner.Id)
	if err := app.Save(base); err != nil {
		t.Fatal(err)
	}
	logical.Set("latest_saved_version", base.Id)
	logical.Set("latest_published_version", base.Id)
	if err := app.Save(logical); err != nil {
		t.Fatal(err)
	}
	return app, logical, base, owner, other
}

func newTestEditSession(t *testing.T, app core.App, logical, owner *core.Record, expires time.Time) *core.Record {
	t.Helper()
	collection, _ := app.FindCollectionByNameOrId("ruleset_edit_sessions")
	record := core.NewRecord(collection)
	record.Set("ruleset", logical.Id)
	record.Set("base_version", logical.GetString("latest_saved_version"))
	record.Set("creator", owner.Id)
	record.Set("activity_at", time.Now().UTC())
	record.Set("expires_at", expires)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}

func editSessionEvent(app core.App, auth *core.Record, rulesetID, sessionID string) *core.RequestEvent {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("id", rulesetID)
	request.SetPathValue("sessionId", sessionID)
	event := &core.RequestEvent{App: app, Auth: auth}
	event.Request = request
	event.Response = httptest.NewRecorder()
	return event
}

func newVersionAsset(t *testing.T, app core.App, versionID, key, name string, content []byte) *core.Record {
	t.Helper()
	collection, _ := app.FindCollectionByNameOrId("ruleset_assets")
	record := core.NewRecord(collection)
	record.Set("ruleset_version", versionID)
	record.Set("asset_key", key)
	record.Set("kind", "image")
	record.Set("display_name", name)
	record.Set("mime_type", "image/png")
	record.Set("checksum", "old-checksum")
	record.Set("metadata", map[string]any{"width": 1, "height": 1})
	record.Set("storage_state", "ready")
	file, _ := filesystem.NewFileFromBytes(content, "cover.png")
	record.Set("file", file)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	content, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScL0WQAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatal(err)
	}
	return content
}
