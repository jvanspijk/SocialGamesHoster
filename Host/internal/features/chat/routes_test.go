package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func TestArchivedRoomMutationUsesSharedConflict(t *testing.T) {
	game := core.NewRecord(&core.Collection{})
	game.Id = "game"
	game.Set("status", "archived")
	room := core.NewRecord(&core.Collection{})
	room.Id = "room"
	room.Set("game", game.Id)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/rooms/room", nil)
	request.SetPathValue("roomId", room.Id)
	event := &core.RequestEvent{}
	event.App = chatRecordLookupApp{
		App:     nil,
		records: map[string]*core.Record{game.Id: game, room.Id: room},
	}
	event.Request = request
	event.Response = recorder

	if err := updateRoom(event); err != nil {
		t.Fatal(err)
	}
	assertArchivedConflict(t, recorder)
}

type chatRecordLookupApp struct {
	core.App
	records map[string]*core.Record
}

func (app chatRecordLookupApp) FindRecordById(
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

func assertArchivedConflict(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
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
