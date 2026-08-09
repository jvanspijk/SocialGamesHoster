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

func TestMessageListingProjectsMessagesAndRejectsInvalidCursor(t *testing.T) {
	app, auth := messageListingFixture()
	message := core.NewRecord(&core.Collection{})
	message.Id = "message"
	message.Set("room", "room")
	message.Set("message_kind", "message")
	message.Set("sender_type", "player")
	message.Set("sender_label_snapshot", "Player")
	message.Set("sender_participant", "participant")
	message.Set("content", "Hello")
	app.messages = []*core.Record{message}

	t.Run("projection", func(t *testing.T) {
		recorder := callMessageListing(t, app, auth, "")
		var body struct {
			Items []struct {
				Content             string `json:"content"`
				SenderParticipantID string `json:"senderParticipantId"`
			} `json:"items"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if recorder.Code != http.StatusOK || len(body.Items) != 1 || body.Items[0].Content != "Hello" || body.Items[0].SenderParticipantID != "participant" {
			t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("invalid cursor", func(t *testing.T) {
		recorder := callMessageListing(t, app, auth, "not-a-cursor")
		var body struct{ Code string }
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if recorder.Code != http.StatusUnprocessableEntity || body.Code != "chat.invalid_cursor" {
			t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

type chatRecordLookupApp struct {
	core.App
	records  map[string]*core.Record
	messages []*core.Record
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

func (app chatRecordLookupApp) FindRecordsByFilter(
	collection any,
	filter, sort string,
	limit, offset int,
	params ...dbx.Params,
) ([]*core.Record, error) {
	if collection == "chat_messages" {
		return app.messages, nil
	}
	return nil, nil
}

func messageListingFixture() (chatRecordLookupApp, *core.Record) {
	game := core.NewRecord(&core.Collection{})
	game.Id = "game"
	room := core.NewRecord(&core.Collection{})
	room.Id = "room"
	room.Set("game", game.Id)
	auth := core.NewRecord(core.NewAuthCollection("game_masters"))
	auth.Set("active", true)
	return chatRecordLookupApp{records: map[string]*core.Record{game.Id: game, room.Id: room}}, auth
}

func callMessageListing(t *testing.T, app core.App, auth *core.Record, cursor string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/rooms/room/messages?cursor="+cursor, nil)
	request.SetPathValue("roomId", "room")
	event := &core.RequestEvent{App: app, Auth: auth}
	event.Request = request
	event.Response = recorder
	if err := listMessages(event); err != nil {
		t.Fatal(err)
	}
	return recorder
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
