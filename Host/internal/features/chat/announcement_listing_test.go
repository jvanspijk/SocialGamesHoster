package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func TestAnnouncementListingKeepsBoundedEnvelopeAndKeysetCursor(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	for _, count := range []int{0, 2, 50, 51} {
		t.Run(fmt.Sprintf("%d records", count), func(t *testing.T) {
			records := announcementRecords(count, created, "announcement-")
			app := &announcementListingApp{records: records}
			response := callAnnouncementListing(t, app, "")
			if got := len(response.Items); got != listingMin(count, 50) {
				t.Fatalf("items = %d, want %d", got, listingMin(count, 50))
			}
			if app.limit != 51 || app.offset != 0 || app.sort != "-created,-id" {
				t.Fatalf("query = limit %d, offset %d, sort %q", app.limit, app.offset, app.sort)
			}
			if count <= 50 && response.NextCursor != "" {
				t.Fatalf("nextCursor = %q, want empty", response.NextCursor)
			}
			if count == 51 {
				cursorCreated, cursorID, err := decodeAnnouncementCursor(response.NextCursor)
				if err != nil || !cursorCreated.Equal(created) || cursorID != records[49].Id {
					t.Fatalf("next cursor = (%v, %q, %v), want (%v, %q)", cursorCreated, cursorID, err, created, records[49].Id)
				}
			}
		})
	}
}

func TestAnnouncementListingCursorUsesCreatedIDTupleAndRejectsMalformedValue(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	app := &announcementListingApp{records: announcementRecords(1, created, "announcement-")}
	callAnnouncementListing(t, app, encodeAnnouncementCursor(created, "same-timestamp-id"))
	if app.filter != "game = {:game} && (created < {:created} || (created = {:created} && id < {:id}))" {
		t.Fatalf("filter = %q", app.filter)
	}
	if got := app.params["created"]; got != created || app.params["id"] != "same-timestamp-id" {
		t.Fatalf("cursor params = %#v", app.params)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/game/announcements?cursor=not-a-cursor", nil)
	request.SetPathValue("id", "game")
	event := &core.RequestEvent{}
	event.App = app
	event.Request = request
	event.Response = recorder
	if err := listAnnouncements(event); err != nil {
		t.Fatal(err)
	}
	var body struct{ Code string }
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusUnprocessableEntity || body.Code != "announcement.invalid_cursor" {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestAnnouncementCursorContinuesPastNewerInsertWithoutRepeatingOlderRecords(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	firstPage := announcementRecords(51, created, "first-")
	olderPage := announcementRecords(5, created.Add(-time.Second), "older-")
	app := &announcementListingApp{records: firstPage, nextRecords: olderPage}
	first := callAnnouncementListing(t, app, "")
	second := callAnnouncementListing(t, app, first.NextCursor)
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

func TestAnnouncementCursorContinuationExcludesNewerInsertInDatabase(t *testing.T) {
	fixture := newAttentionFixture(t)
	oldCreated := time.Date(2026, time.August, 1, 10, 0, 0, 0, time.UTC)
	originalIDs := make(map[string]bool, 51)
	for index := 0; index < 51; index++ {
		originalIDs[saveAnnouncementRecord(t, fixture.app, fixture.game, fixture.gameMaster, oldCreated.Add(time.Duration(index)*time.Second)).Id] = true
	}

	first := callAnnouncementListingForGame(t, fixture.app, fixture.game.Id, "")
	if len(first.Items) != 50 || first.NextCursor == "" {
		t.Fatalf("first page = %#v", first)
	}
	newer := saveAnnouncementRecord(t, fixture.app, fixture.game, fixture.gameMaster, oldCreated.Add(2*time.Hour))
	second := callAnnouncementListingForGame(t, fixture.app, fixture.game.Id, first.NextCursor)

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

type announcementListingApp struct {
	core.App
	records, nextRecords []*core.Record
	filter, sort         string
	limit, offset        int
	params               dbx.Params
}

func (app *announcementListingApp) FindRecordById(collection any, id string, _ ...func(*dbx.SelectQuery) error) (*core.Record, error) {
	game := core.NewRecord(&core.Collection{})
	game.Id = id
	return game, nil
}

func (app *announcementListingApp) FindRecordsByFilter(collection any, filter, sort string, limit, offset int, params ...dbx.Params) ([]*core.Record, error) {
	if collection != "attention_items" {
		return nil, nil
	}
	app.filter, app.sort, app.limit, app.offset = filter, sort, limit, offset
	app.params = params[0]
	if _, ok := app.params["created"]; ok && app.nextRecords != nil {
		return app.nextRecords, nil
	}
	return app.records, nil
}

type announcementListingResponse struct {
	Items      []struct{ ID string } `json:"items"`
	NextCursor string                `json:"nextCursor"`
}

func callAnnouncementListing(t *testing.T, app core.App, cursor string) announcementListingResponse {
	return callAnnouncementListingForGame(t, app, "game", cursor)
}

func callAnnouncementListingForGame(t *testing.T, app core.App, gameID, cursor string) announcementListingResponse {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/"+gameID+"/announcements?cursor="+cursor, nil)
	request.SetPathValue("id", gameID)
	event := &core.RequestEvent{}
	event.App = app
	event.Request = request
	event.Response = recorder
	if err := listAnnouncements(event); err != nil {
		t.Fatal(err)
	}
	var response announcementListingResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	return response
}

func announcementRecords(count int, created time.Time, prefix string) []*core.Record {
	records := make([]*core.Record, count)
	for index := range records {
		record := core.NewRecord(&core.Collection{})
		record.Id = prefix + string(rune('a'+index))
		record.Set("created", created)
		record.Set("sender_label_snapshot", "Host")
		record.Set("content", "Attention")
		record.Set("audience", "all")
		records[index] = record
	}
	return records
}

func listingMin(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func saveAnnouncementRecord(t *testing.T, app core.App, game, gameMaster *core.Record, created time.Time) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("attention_items")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("game", game.Id)
	record.Set("kind", "announcement")
	record.Set("sender", gameMaster.Id)
	record.Set("sender_label_snapshot", "Host")
	record.Set("content", "Attention")
	record.Set("audience", "all")
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	record.Set("created", created)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}
