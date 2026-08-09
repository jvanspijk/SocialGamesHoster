package games

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func TestActivityListingProjectsRecognizedActions(t *testing.T) {
	created := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	records := activityRecords(2, created)
	records[0].Set("actor_label", "")
	records[1].Set("action", "unprojected.action")

	response := callActivityListing(t, &activityListingApp{records: records}, "")
	if len(response.Items) != 1 || response.Items[0].Text != "A game master paused the game" {
		t.Fatalf("items = %#v", response.Items)
	}
}

func TestActivityListingRejectsInvalidCursor(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/game/activity?cursor=not-a-cursor", nil)
	request.SetPathValue("id", "game")
	event := &core.RequestEvent{}
	event.App = &activityListingApp{}
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

type activityListingApp struct {
	core.App
	records []*core.Record
}

func (app *activityListingApp) FindRecordById(collection any, id string, _ ...func(*dbx.SelectQuery) error) (*core.Record, error) {
	game := core.NewRecord(&core.Collection{})
	game.Id = id
	return game, nil
}

func (app *activityListingApp) FindRecordsByFilter(collection any, filter, sort string, limit, offset int, params ...dbx.Params) ([]*core.Record, error) {
	return app.records, nil
}

type activityListingResponse struct {
	Items []struct {
		Text string `json:"text"`
	} `json:"items"`
}

func callActivityListing(t *testing.T, app core.App, cursor string) activityListingResponse {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/games/game/activity?cursor="+cursor, nil)
	request.SetPathValue("id", "game")
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
	records := make([]*core.Record, count)
	for index := range records {
		record := core.NewRecord(&core.Collection{})
		record.Id = "activity-" + string(rune('a'+index))
		record.Set("created", created)
		record.Set("actor_label", "Host")
		record.Set("action", "game.pause")
		records[index] = record
	}
	return records
}
