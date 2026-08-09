package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func TestAnnouncementListingProjectsSummaries(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	record := announcementRecords(1, created, "announcement-")[0]
	record.Set("content", "Important update")
	record.Set("audience", "team")
	record.Set("target_id", "red")

	response := callAnnouncementListing(t, &announcementListingApp{records: []*core.Record{record}}, "")
	if len(response.Items) != 1 {
		t.Fatalf("items = %#v", response.Items)
	}
	item := response.Items[0]
	if item.ID != record.Id || item.Content != "Important update" || item.Audience != "team" || item.TargetID != "red" ||
		item.RecipientTotal != 0 || item.AcknowledgementCount != 0 {
		t.Fatalf("item = %#v", item)
	}
}

func TestAnnouncementListingRejectsInvalidCursor(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/game/announcements?cursor=not-a-cursor", nil)
	request.SetPathValue("id", "game")
	event := &core.RequestEvent{}
	event.App = &announcementListingApp{}
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

type announcementListingApp struct {
	core.App
	records []*core.Record
}

func (app *announcementListingApp) FindRecordById(collection any, id string, _ ...func(*dbx.SelectQuery) error) (*core.Record, error) {
	game := core.NewRecord(&core.Collection{})
	game.Id = id
	return game, nil
}

func (app *announcementListingApp) FindRecordsByFilter(collection any, filter, sort string, limit, offset int, params ...dbx.Params) ([]*core.Record, error) {
	if collection == "attention_items" {
		return app.records, nil
	}
	return nil, nil
}

type announcementListingResponse struct {
	Items []struct {
		ID                   string `json:"id"`
		Content              string `json:"content"`
		Audience             string `json:"audience"`
		TargetID             string `json:"targetId"`
		RecipientTotal       int    `json:"recipientTotal"`
		AcknowledgementCount int    `json:"acknowledgementCount"`
	} `json:"items"`
}

func callAnnouncementListing(t *testing.T, app core.App, cursor string) announcementListingResponse {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/game/announcements?cursor="+cursor, nil)
	request.SetPathValue("id", "game")
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
