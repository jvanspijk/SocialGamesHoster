package games

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	_ "github.com/jvanspijk/SocialGamesHoster/Host/migrations"
)

func TestActivityListingKeepsBoundedEnvelopeAndKeysetCursor(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	for _, count := range []int{0, 2, 50, 51} {
		t.Run(fmt.Sprintf("%d records", count), func(t *testing.T) {
			records := activityRecords(count, created)
			app := &activityListingApp{records: records}
			response := callActivityListing(t, app, "")
			if got := len(response.Items); got != min(count, 50) {
				t.Fatalf("items = %d, want %d", got, min(count, 50))
			}
			if app.limit != 51 || app.offset != 0 || app.sort != "-created,-id" {
				t.Fatalf("query = limit %d, offset %d, sort %q", app.limit, app.offset, app.sort)
			}
			if count <= 50 && response.NextCursor != "" {
				t.Fatalf("nextCursor = %q, want empty", response.NextCursor)
			}
			if count == 51 {
				cursorCreated, cursorID, err := decodeActivityCursor(response.NextCursor)
				if err != nil || !cursorCreated.Equal(created) || cursorID != records[49].Id {
					t.Fatalf("next cursor = (%v, %q, %v), want (%v, %q)", cursorCreated, cursorID, err, created, records[49].Id)
				}
			}
		})
	}
}

func TestActivityListingCursorUsesCreatedIDTupleAndRejectsMalformedValue(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	app := &activityListingApp{records: activityRecords(1, created)}
	callActivityListing(t, app, encodeActivityCursor(created, "same-timestamp-id"))
	if app.filter != "game = {:game} && (created < {:created} || (created = {:created} && id < {:id}))" {
		t.Fatalf("filter = %q", app.filter)
	}
	if got := app.params["created"]; got != created || app.params["id"] != "same-timestamp-id" {
		t.Fatalf("cursor params = %#v", app.params)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/game/activity?cursor=not-a-cursor", nil)
	request.SetPathValue("id", "game")
	event := &core.RequestEvent{}
	event.App = app
	event.Request = request
	event.Response = recorder
	if err := listActivity(event); err != nil {
		t.Fatal(err)
	}
	var body struct{ Code string }
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusUnprocessableEntity || body.Code != "activity.invalid_cursor" {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestActivityCursorContinuesPastNewerInsertWithoutRepeatingOlderRecords(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	firstPage := activityRecordsWithPrefix(51, created, "first-")
	olderPage := activityRecordsWithPrefix(5, created.Add(-time.Second), "older-")
	app := &activityListingApp{records: firstPage, nextRecords: olderPage}
	first := callActivityListing(t, app, "")
	second := callActivityListing(t, app, first.NextCursor)
	seen := map[string]bool{}
	for _, item := range append(first.Items, second.Items...) {
		if seen[item.ID] {
			t.Fatalf("repeated older record %q", item.ID)
		}
		seen[item.ID] = true
	}
	if len(seen) != 55 || app.params["created"] != created || app.params["id"] != firstPage[49].Id {
		t.Fatalf("older records were not continued after newer insert: seen=%d params=%#v", len(seen), app.params)
	}
}

func TestActivityCursorContinuationExcludesNewerInsertInDatabase(t *testing.T) {
	app, game := newActivityListingFixture(t)
	oldCreated := time.Date(2026, time.August, 1, 10, 0, 0, 0, time.UTC)
	originalIDs := make(map[string]bool, 51)
	for index := 0; index < 51; index++ {
		originalIDs[saveActivityRecord(t, app, game.Id, oldCreated.Add(time.Duration(index)*time.Second)).Id] = true
	}

	first := callActivityListingForGame(t, app, game.Id, "")
	if len(first.Items) != 50 || first.NextCursor == "" {
		t.Fatalf("first page = %#v", first)
	}
	newer := saveActivityRecord(t, app, game.Id, oldCreated.Add(2*time.Hour))
	second := callActivityListingForGame(t, app, game.Id, first.NextCursor)

	seen := map[string]bool{}
	for _, item := range append(first.Items, second.Items...) {
		if seen[item.ID] {
			t.Fatalf("duplicate record %q", item.ID)
		}
		seen[item.ID] = true
	}
	if len(second.Items) != 1 || seen[newer.Id] || len(seen) != len(originalIDs) {
		t.Fatalf("continuation = %#v, want all %d original records without newer %q", seen, len(originalIDs), newer.Id)
	}
	for id := range originalIDs {
		if !seen[id] {
			t.Fatalf("skipped original record %q", id)
		}
	}
}

type activityListingApp struct {
	core.App
	records, nextRecords []*core.Record
	filter, sort         string
	limit, offset        int
	params               dbx.Params
}

func (app *activityListingApp) FindRecordById(collection any, id string, _ ...func(*dbx.SelectQuery) error) (*core.Record, error) {
	game := core.NewRecord(&core.Collection{})
	game.Id = id
	return game, nil
}

func (app *activityListingApp) FindRecordsByFilter(collection any, filter, sort string, limit, offset int, params ...dbx.Params) ([]*core.Record, error) {
	app.filter, app.sort, app.limit, app.offset = filter, sort, limit, offset
	app.params = params[0]
	if _, ok := app.params["created"]; ok && app.nextRecords != nil {
		return app.nextRecords, nil
	}
	return app.records, nil
}

type activityListingResponse struct {
	Items      []struct{ ID string } `json:"items"`
	NextCursor string                `json:"nextCursor"`
}

func callActivityListing(t *testing.T, app core.App, cursor string) activityListingResponse {
	return callActivityListingForGame(t, app, "game", cursor)
}

func callActivityListingForGame(t *testing.T, app core.App, gameID, cursor string) activityListingResponse {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/"+gameID+"/activity?cursor="+cursor, nil)
	request.SetPathValue("id", gameID)
	event := &core.RequestEvent{}
	event.App = app
	event.Request = request
	event.Response = recorder
	if err := listActivity(event); err != nil {
		t.Fatal(err)
	}
	var response activityListingResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	return response
}

func activityRecords(count int, created time.Time) []*core.Record {
	return activityRecordsWithPrefix(count, created, "activity-")
}

func activityRecordsWithPrefix(count int, created time.Time, prefix string) []*core.Record {
	records := make([]*core.Record, count)
	for index := range records {
		record := core.NewRecord(&core.Collection{})
		record.Id = prefix + string(rune('a'+index))
		record.Set("created", created)
		record.Set("actor_label", "Host")
		record.Set("action", "game.pause")
		records[index] = record
	}
	return records
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func newActivityListingFixture(t *testing.T) (core.App, *core.Record) {
	t.Helper()
	if err := os.MkdirAll(".testdata", 0o700); err != nil {
		t.Fatal(err)
	}
	dataDir, err := os.MkdirTemp(".testdata", "activity-listing-*")
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

	gameMasters, _ := app.FindCollectionByNameOrId("game_masters")
	gameMaster := core.NewRecord(gameMasters)
	gameMaster.Set("username", "activity-host")
	gameMaster.Set("display_name", "Host")
	gameMaster.Set("is_owner", true)
	gameMaster.Set("active", true)
	gameMaster.SetPassword("secret-password")
	if err := app.Save(gameMaster); err != nil {
		t.Fatal(err)
	}
	rulesetsCollection, _ := app.FindCollectionByNameOrId("rulesets")
	ruleset := core.NewRecord(rulesetsCollection)
	ruleset.Set("slug", "activity-listing")
	ruleset.Set("name", "Activity listing")
	ruleset.Set("created_by", gameMaster.Id)
	if err := app.Save(ruleset); err != nil {
		t.Fatal(err)
	}
	definition := rulesets.DefinitionV1{SchemaVersion: 1}
	versions, _ := app.FindCollectionByNameOrId("ruleset_versions")
	version := core.NewRecord(versions)
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
	games, _ := app.FindCollectionByNameOrId("games")
	game := core.NewRecord(games)
	game.Set("name", "Activity listing")
	game.Set("status", "running")
	game.Set("ruleset_version", version.Id)
	game.Set("ruleset_snapshot", definition)
	game.Set("timer_state", "inactive")
	game.Set("created_by", gameMaster.Id)
	if err := app.Save(game); err != nil {
		t.Fatal(err)
	}
	return app, game
}

func saveActivityRecord(t *testing.T, app core.App, gameID string, created time.Time) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("game_audit")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("game", gameID)
	record.Set("actor_type", "game_master")
	record.Set("actor_label", "Host")
	record.Set("action", "game.pause")
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	record.Set("created", created)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}
