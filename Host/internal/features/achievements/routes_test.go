package achievements

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func TestSpoilerAchievementVisibilityFollowsGameCompletion(t *testing.T) {
	for _, status := range []string{"running", "paused"} {
		if awardVisibleDuringStatus(true, status) {
			t.Fatalf("hidden achievement became visible while game status was %q", status)
		}
	}
	for _, status := range []string{"review", "archived"} {
		if !awardVisibleDuringStatus(true, status) {
			t.Fatalf("hidden achievement remained concealed after game status became %q", status)
		}
	}
	if !awardVisibleDuringStatus(false, "running") {
		t.Fatal("ordinary achievement must be visible during a running game")
	}
}

func TestArchivedAchievementRevocationUsesSharedConflict(t *testing.T) {
	game := core.NewRecord(&core.Collection{})
	game.Id = "game"
	game.Set("status", "archived")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/games/game/achievement-awards/award", nil)
	request.SetPathValue("id", game.Id)
	request.SetPathValue("awardId", "award")
	event := &core.RequestEvent{}
	event.App = achievementRecordLookupApp{
		App:     nil,
		records: map[string]*core.Record{game.Id: game},
	}
	event.Request = request
	event.Response = recorder

	if err := revoke(event); err != nil {
		t.Fatal(err)
	}
	var response struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusConflict || response.Code != "game.archived_immutable" {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

type achievementRecordLookupApp struct {
	core.App
	records map[string]*core.Record
}

func (app achievementRecordLookupApp) FindRecordById(
	collection any,
	recordID string,
	optFilters ...func(*dbx.SelectQuery) error,
) (*core.Record, error) {
	record := app.records[recordID]
	if record == nil {
		return nil, errors.New("record not found")
	}
	return record, nil
}
