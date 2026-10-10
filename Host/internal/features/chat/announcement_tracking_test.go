package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
)

func TestAnnouncementReminderPreservesMessageMediaAndAcknowledgements(t *testing.T) {
	fixture := newAttentionFixture(t)
	response := createMultipartAnnouncement(t, fixture, map[string]string{
		"content": "Gather at the table", "audience": "all", "imageDescription": "A meeting point",
	}, map[string]uploadFixture{"imageFile": {name: "map.png", contentType: "image/png", content: testAnnouncementPNG(t)}})
	if response.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", response.Code, response.Body.String())
	}
	items, _ := fixture.app.FindRecordsByFilter("attention_items", "game = {:game}", "", 0, 0, dbx.Params{"game": fixture.game.Id})
	item := items[0]
	before := projectPlayerAttention(item)
	receipts, _ := fixture.app.FindRecordsByFilter("attention_receipts", "attention_item = {:item} && participant = {:participant}", "", 1, 0,
		dbx.Params{"item": item.Id, "participant": fixture.participants[0].Id})
	receipts[0].Set("acknowledged_at", time.Now().UTC())
	if err := fixture.app.Save(receipts[0]); err != nil {
		t.Fatal(err)
	}
	savedReceipt, err := fixture.app.FindRecordById("attention_receipts", receipts[0].Id)
	if err != nil {
		t.Fatal(err)
	}
	acknowledgedAt := savedReceipt.GetDateTime("acknowledged_at")
	fixture.participants[2].Set("status", "left")
	if err := fixture.app.Save(fixture.participants[2]); err != nil {
		t.Fatal(err)
	}
	fixture.participants[1].Set("game_alias", "Pending player")
	fixture.participants[1].Set("status", "eliminated")
	if err := fixture.app.Save(fixture.participants[1]); err != nil {
		t.Fatal(err)
	}

	clients := make([]*recordingRoomClient, 3)
	for i, profile := range fixture.profiles {
		client := &recordingRoomClient{DefaultClient: subscriptions.NewDefaultClient()}
		client.Set(apis.RealtimeClientAuthKey, profile)
		client.Subscribe("game:" + fixture.game.Id + ":public")
		fixture.app.SubscriptionsBroker().Register(client)
		defer fixture.app.SubscriptionsBroker().Unregister(client.Id())
		clients[i] = client
	}
	response = announcementTrackingCommand(t, fixture, item, remindAnnouncement)
	if response.Code != http.StatusOK {
		t.Fatalf("remind = %d: %s", response.Code, response.Body.String())
	}
	if len(clients[0].messages) != 0 || len(clients[1].messages) != 1 || len(clients[2].messages) != 0 {
		t.Fatalf("reminder delivery counts = %d, %d, %d", len(clients[0].messages), len(clients[1].messages), len(clients[2].messages))
	}
	var delivered struct {
		Payload map[string]any `json:"payload"`
	}
	if err := json.Unmarshal(clients[1].messages[0].Data, &delivered); err != nil {
		t.Fatal(err)
	}
	if delivered.Payload["content"] != before["content"] || delivered.Payload["id"] != item.Id ||
		!reflect.DeepEqual(delivered.Payload["image"], before["image"]) {
		t.Fatalf("reminder changed message or media: %#v", delivered.Payload)
	}
	currentReceipt, _ := fixture.app.FindRecordById("attention_receipts", receipts[0].Id)
	if currentReceipt.GetDateTime("acknowledged_at") != acknowledgedAt {
		t.Fatal("reminder reset acknowledgement")
	}
	count, _ := fixture.app.CountRecords("attention_items", dbx.HashExp{"game": fixture.game.Id})
	if count != 1 {
		t.Fatalf("reminder duplicated announcement: %d", count)
	}
	var summary struct {
		UnacknowledgedPlayers []struct {
			ParticipantID string `json:"participantId"`
			DisplayName   string `json:"displayName"`
		} `json:"unacknowledgedPlayers"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if len(summary.UnacknowledgedPlayers) != 1 || summary.UnacknowledgedPlayers[0].DisplayName != "Pending player" {
		t.Fatalf("pending players = %#v", summary)
	}

	response = announcementTrackingCommand(t, fixture, item, clearAnnouncementTracking)
	if response.Code != http.StatusNoContent {
		t.Fatalf("clear = %d: %s", response.Code, response.Body.String())
	}
	currentItem, _ := fixture.app.FindRecordById("attention_items", item.Id)
	if !currentItem.GetBool("tracking_cleared") || !reflect.DeepEqual(projectPlayerAttention(currentItem), before) {
		t.Fatal("clear removed or changed announcement")
	}
	pending, err := UnacknowledgedAttentionForParticipant(fixture.app, fixture.game.Id, fixture.participants[1].Id)
	if err != nil || len(pending) != 1 {
		t.Fatalf("clear changed pending delivery: %#v, %v", pending, err)
	}
	if response := announcementTrackingCommand(t, fixture, item, remindAnnouncement); response.Code != http.StatusConflict {
		t.Fatalf("cleared reminder = %d", response.Code)
	}
	if response := announcementTrackingCommand(t, fixture, item, clearAnnouncementTracking); response.Code != http.StatusNoContent {
		t.Fatalf("repeat clear = %d", response.Code)
	}
}

func TestClearAnnouncementWithoutMedia(t *testing.T) {
	fixture := newAttentionFixture(t)
	response := createMultipartAnnouncement(t, fixture, map[string]string{
		"content": "Plain announcement", "audience": "all",
	}, nil)
	if response.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", response.Code, response.Body.String())
	}
	items, err := fixture.app.FindRecordsByFilter("attention_items", "game = {:game}", "", 1, 0, dbx.Params{"game": fixture.game.Id})
	if err != nil || len(items) != 1 {
		t.Fatalf("items = %#v, %v", items, err)
	}
	response = announcementTrackingCommand(t, fixture, items[0], clearAnnouncementTracking)
	if response.Code != http.StatusNoContent {
		t.Fatalf("clear = %d: %s", response.Code, response.Body.String())
	}
	_, summaries := AdminViewData(fixture.app, fixture.game.Id)
	if len(summaries) != 1 || summaries[0]["trackingCleared"] != true {
		t.Fatalf("clear was not saved: %#v", summaries)
	}
}

func TestAnnouncementTrackingRejectsNonLiveGamesAndForeignAnnouncements(t *testing.T) {
	for _, command := range []func(*core.RequestEvent) error{remindAnnouncement, clearAnnouncementTracking} {
		fixture := newAttentionFixture(t)
		collection, _ := fixture.app.FindCollectionByNameOrId("attention_items")
		item := core.NewRecord(collection)
		item.Set("game", fixture.game.Id)
		item.Set("kind", "announcement")
		item.Set("sender", fixture.gameMaster.Id)
		item.Set("sender_label_snapshot", "Host")
		item.Set("content", "Message")
		item.Set("audience", "all")
		if err := fixture.app.Save(item); err != nil {
			t.Fatal(err)
		}
		if response := announcementTrackingCommand(t, fixture, item, remindAnnouncement); response.Code != http.StatusConflict {
			t.Fatalf("empty audience = %d", response.Code)
		}
		for _, status := range []string{"draft", "review", "archived"} {
			fixture.game.Set("status", status)
			if err := fixture.app.Save(fixture.game); err != nil {
				t.Fatal(err)
			}
			if response := announcementTrackingCommand(t, fixture, item, command); response.Code != http.StatusConflict {
				t.Fatalf("%s = %d", status, response.Code)
			}
		}
		otherGame := fixture.game.Fresh()
		otherGame.Id = ""
		otherGame.MarkAsNew()
		otherGame.Set("status", "running")
		if err := fixture.app.Save(otherGame); err != nil {
			t.Fatal(err)
		}
		fixture.game = otherGame
		if response := announcementTrackingCommand(t, fixture, item, command); response.Code != http.StatusNotFound {
			t.Fatalf("foreign announcement = %d", response.Code)
		}
	}
}

func announcementTrackingCommand(t *testing.T, fixture attentionFixture, item *core.Record, handler func(*core.RequestEvent) error) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/games/"+fixture.game.Id+"/announcements/"+item.Id, nil)
	request.SetPathValue("id", fixture.game.Id)
	request.SetPathValue("announcementId", item.Id)
	response := httptest.NewRecorder()
	event := &core.RequestEvent{App: fixture.app}
	event.Auth = fixture.gameMaster
	event.Request = request
	event.Response = response
	if err := handler(event); err != nil {
		t.Fatal(err)
	}
	return response
}
