package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func TestMessagePaginationRetainsTimestampTies(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.General = &rulesets.RoomPermission{Visible: true, Readable: true, Sendable: true}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLobbyRoom(fixture.app, fixture.game.Id, fixture.definition); err != nil {
		t.Fatal(err)
	}
	room, err := findRoomByKey(fixture.app, fixture.game.Id, "general")
	if err != nil {
		t.Fatal(err)
	}
	collection, _ := fixture.app.FindCollectionByNameOrId("chat_messages")
	created := time.Date(2026, 10, 5, 12, 0, 0, 598000000, time.UTC)
	for index := 0; index < 116; index++ {
		message := core.NewRecord(collection)
		message.Id = fmt.Sprintf("%015d", index)
		message.Set("room", room.Id)
		message.Set("sender_type", "game_master")
		message.Set("sender_id", fixture.gameMaster.Id)
		message.Set("sender_label_snapshot", "Host")
		message.Set("message_kind", "message")
		message.Set("content", "Hello")
		if err := fixture.app.Save(message); err != nil {
			t.Fatal(err)
		}
		if _, err := fixture.app.DB().NewQuery("UPDATE chat_messages SET created={:created} WHERE id={:id}").Bind(dbx.Params{"created": created.Format(types.DefaultDateLayout), "id": message.Id}).Execute(); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	cursor := ""
	for page := 0; page < 4; page++ {
		request := httptest.NewRequest(http.MethodGet, "/messages?cursor="+cursor, nil)
		request.SetPathValue("roomId", room.Id)
		recorder := httptest.NewRecorder()
		event := &core.RequestEvent{Auth: fixture.gameMaster}
		event.App, event.Request, event.Response = fixture.app, request, recorder
		if err := listMessages(event); err != nil {
			t.Fatal(err)
		}
		var response struct {
			Items      []struct{ ID string }
			NextCursor string
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		for _, item := range response.Items {
			if seen[item.ID] {
				t.Fatalf("message %s appeared twice", item.ID)
			}
			seen[item.ID] = true
		}
		cursor = response.NextCursor
		if cursor == "" {
			break
		}
	}
	if len(seen) != 116 || cursor != "" {
		t.Fatalf("pagination returned %d of 116 messages, remaining cursor %q", len(seen), cursor)
	}
}

func callUnread(t *testing.T, fixture attentionFixture, actor *core.Record, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/unread-counts", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("id", fixture.game.Id)
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{Auth: actor}
	event.App, event.Request, event.Response = fixture.app, request, recorder
	if err := unreadCounts(event); err != nil {
		t.Fatal(err)
	}
	return recorder
}

func TestUnreadCountsRespectMarkersAndRoomAccess(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.General = &rulesets.RoomPermission{Visible: true, Readable: true, Sendable: true}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLobbyRoom(fixture.app, fixture.game.Id, fixture.definition); err != nil {
		t.Fatal(err)
	}
	for _, participant := range fixture.participants {
		if err := AddParticipant(fixture.app, fixture.game.Id, participant); err != nil {
			t.Fatal(err)
		}
	}
	general, _ := findRoomByKey(fixture.app, fixture.game.Id, "general")
	private, _ := findRoomByKey(fixture.app, fixture.game.Id, "gm:"+fixture.participants[1].Id)
	collection, _ := fixture.app.FindCollectionByNameOrId("chat_messages")
	created := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for _, row := range []struct{ id, room string }{{"aaaaaaaaaaaaaaa", general.Id}, {"bbbbbbbbbbbbbbb", general.Id}, {"ccccccccccccccc", private.Id}} {
		message := core.NewRecord(collection)
		message.Id = row.id
		message.Set("room", row.room)
		message.Set("sender_type", "game_master")
		message.Set("sender_id", fixture.gameMaster.Id)
		message.Set("sender_label_snapshot", "Host")
		message.Set("message_kind", "message")
		message.Set("content", "Hello")
		if row.id == "bbbbbbbbbbbbbbb" {
			message.Set("deleted_at", created.Add(time.Hour))
		}
		message.Set("created", created)
		if err := fixture.app.Save(message); err != nil {
			t.Fatal(err)
		}
		if _, err := fixture.app.DB().NewQuery("UPDATE chat_messages SET created={:created} WHERE id={:id}").Bind(dbx.Params{"created": created.Format(types.DefaultDateLayout), "id": message.Id}).Execute(); err != nil {
			t.Fatal(err)
		}
	}
	for _, actor := range []*core.Record{fixture.gameMaster, fixture.profiles[0]} {
		body, _ := json.Marshal(map[string]any{"markers": map[string]any{general.Id: readMarker{ID: "aaaaaaaaaaaaaaa", CreatedAt: created}}})
		response := callUnread(t, fixture, actor, string(body))
		var result struct {
			Counts map[string]int
			Total  int
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		wantTotal := 1
		if actor == fixture.gameMaster {
			wantTotal = 2
		}
		if response.Code != 200 || result.Total != wantTotal || result.Counts[general.Id] != 1 {
			t.Fatalf("response = %d %s", response.Code, response.Body.String())
		}
		if actor != fixture.gameMaster {
			if _, found := result.Counts[private.Id]; found {
				t.Fatal("private room leaked")
			}
		}
	}
	t.Run("unauthenticated", func(t *testing.T) {
		if response := callUnread(t, fixture, nil, `{}`); response.Code != 401 {
			t.Fatalf("status = %d", response.Code)
		}
	})
	t.Run("invalid marker", func(t *testing.T) {
		if response := callUnread(t, fixture, fixture.gameMaster, `{"markers":{"room":{"id":"x","createdAt":"bad"}}}`); response.Code != 422 {
			t.Fatalf("status = %d", response.Code)
		}
	})
	t.Run("empty rooms retain zero", func(t *testing.T) {
		response := callUnread(t, fixture, fixture.profiles[2], `{}`)
		var result struct {
			Counts map[string]int
			Total  int
		}
		json.Unmarshal(response.Body.Bytes(), &result)
		if response.Code != 200 || result.Total != 2 || len(result.Counts) != 2 {
			t.Fatalf("response = %d %s", response.Code, response.Body.String())
		}
	})
}

func TestUnreadCountsRespectArchivedMembershipWindow(t *testing.T) {
	fixture := newAttentionFixture(t)
	if err := AddParticipant(fixture.app, fixture.game.Id, fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	room, _ := findRoomByKey(fixture.app, fixture.game.Id, "gm:"+fixture.participants[0].Id)
	joined := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	left := joined.Add(time.Hour)
	membership, _ := findMembership(fixture.app, room.Id, fixture.participants[0].Id)
	membership.Set("joined_at", joined)
	membership.Set("left_at", left)
	membership.Set("historical_access", true)
	if err := fixture.app.Save(membership); err != nil {
		t.Fatal(err)
	}
	fixture.game.Set("status", "archived")
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	collection, _ := fixture.app.FindCollectionByNameOrId("chat_messages")
	for _, created := range []time.Time{joined.Add(-time.Second), joined, left, left.Add(time.Second)} {
		message := core.NewRecord(collection)
		message.Set("room", room.Id)
		message.Set("message_kind", "message")
		message.Set("sender_type", "game_master")
		message.Set("sender_id", fixture.gameMaster.Id)
		message.Set("sender_label_snapshot", "Host")
		message.Set("content", "Hello")
		if err := fixture.app.Save(message); err != nil {
			t.Fatal(err)
		}
		if _, err := fixture.app.DB().NewQuery("UPDATE chat_messages SET created={:created} WHERE id={:id}").Bind(dbx.Params{"created": created.Format(types.DefaultDateLayout), "id": message.Id}).Execute(); err != nil {
			t.Fatal(err)
		}
	}
	for _, actor := range []*core.Record{fixture.profiles[0], fixture.gameMaster} {
		response := callUnread(t, fixture, actor, `{}`)
		var result struct {
			Counts map[string]int
			Total  int
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		want := 2
		if actor == fixture.gameMaster {
			want = 4
		}
		if response.Code != 200 || result.Counts[room.Id] != want {
			t.Fatalf("response = %d %s", response.Code, response.Body.String())
		}
	}
}

func TestUnreadCountsAreCappedAndIncludeRemovedMessages(t *testing.T) {
	fixture := newAttentionFixture(t)
	if err := AddParticipant(fixture.app, fixture.game.Id, fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	room, _ := findRoomByKey(fixture.app, fixture.game.Id, "gm:"+fixture.participants[0].Id)
	collection, _ := fixture.app.FindCollectionByNameOrId("chat_messages")
	for i := 0; i < 105; i++ {
		message := core.NewRecord(collection)
		message.Set("room", room.Id)
		message.Set("message_kind", "message")
		message.Set("sender_type", "game_master")
		message.Set("sender_id", fixture.gameMaster.Id)
		message.Set("sender_label_snapshot", "Host")
		message.Set("content", "Hello")
		if i == 0 {
			message.Set("deleted_at", time.Now().UTC())
		}
		if err := fixture.app.Save(message); err != nil {
			t.Fatal(err)
		}
	}
	response := callUnread(t, fixture, fixture.profiles[0], `{}`)
	var result struct {
		Counts map[string]int
		Total  int
	}
	json.Unmarshal(response.Body.Bytes(), &result)
	if response.Code != 200 || result.Total != 99 || result.Counts[room.Id] != 99 {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestUnreadCountsUseCurrentRoleAccessOnEveryRequest(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.Teams = map[string]rulesets.RoomPermission{
		"red":  {Visible: true, Readable: true},
		"blue": {Visible: true, Readable: true},
	}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := PrepareRoleRooms(fixture.app, fixture.game.Id, fixture.definition, fixture.participants); err != nil {
		t.Fatal(err)
	}
	red, _ := findRoomByKey(fixture.app, fixture.game.Id, "team:red")
	blue, _ := findRoomByKey(fixture.app, fixture.game.Id, "team:blue")
	collection, _ := fixture.app.FindCollectionByNameOrId("chat_messages")
	for _, room := range []*core.Record{red, blue} {
		message := core.NewRecord(collection)
		message.Set("room", room.Id)
		message.Set("message_kind", "message")
		message.Set("sender_type", "game_master")
		message.Set("sender_id", fixture.gameMaster.Id)
		message.Set("sender_label_snapshot", "Host")
		message.Set("content", "Team message")
		if err := fixture.app.Save(message); err != nil {
			t.Fatal(err)
		}
	}
	for _, role := range []string{"red-one", "blue-role"} {
		fixture.participants[0].Set("role_key", role)
		if err := fixture.app.Save(fixture.participants[0]); err != nil {
			t.Fatal(err)
		}
		response := callUnread(t, fixture, fixture.profiles[0], `{}`)
		var result struct {
			Counts map[string]int
			Total  int
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		want, forbidden := red.Id, blue.Id
		if role == "blue-role" {
			want, forbidden = blue.Id, red.Id
		}
		if response.Code != 200 || result.Total != 1 || result.Counts[want] != 1 {
			t.Fatalf("response = %d %s", response.Code, response.Body.String())
		}
		if _, found := result.Counts[forbidden]; found {
			t.Fatal("previous team count leaked")
		}
	}
}
