package chat

import (
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

func TestGameLifecycleChatOperationsPreserveOwnedState(t *testing.T) {
	fixture := newAttentionFixture(t)
	generalPolicy := rulesets.RoomPermission{Visible: true, Readable: true, Sendable: true}
	fixture.definition.Chat.DefaultPolicy.General = &generalPolicy

	if err := EnsureLobbyRoom(fixture.app, fixture.game.Id, fixture.definition); err != nil {
		t.Fatal(err)
	}
	if err := AddParticipant(fixture.app, fixture.game.Id, fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	general, err := findRoomByKey(fixture.app, fixture.game.Id, "general")
	if err != nil {
		t.Fatal(err)
	}
	gmRoom, err := findRoomByKey(fixture.app, fixture.game.Id, "gm:"+fixture.participants[0].Id)
	if err != nil {
		t.Fatal(err)
	}
	if countRecords(t, fixture.app, "chat_memberships", "room = {:room}", dbx.Params{"room": general.Id}) != 1 ||
		countRecords(t, fixture.app, "chat_memberships", "room = {:room}", dbx.Params{"room": gmRoom.Id}) != 1 {
		t.Fatal("joining did not establish both chat memberships")
	}

	leftAt := time.Now().UTC()
	if err := CloseParticipantMemberships(fixture.app, fixture.participants[0].Id, leftAt); err != nil {
		t.Fatal(err)
	}
	if err := AddParticipant(fixture.app, fixture.game.Id, fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	membership, err := findMembership(fixture.app, general.Id, fixture.participants[0].Id)
	if err != nil || !membership.GetDateTime("left_at").IsZero() {
		t.Fatalf("rejoining did not reopen membership: %v", err)
	}

	archivedAt := leftAt.Add(time.Second)
	if err := FreezeHistoricalAccess(fixture.app, fixture.game.Id, archivedAt); err != nil {
		t.Fatal(err)
	}
	membership, err = findMembership(fixture.app, general.Id, fixture.participants[0].Id)
	if err != nil || !membership.GetBool("historical_access") || membership.GetDateTime("left_at").IsZero() {
		t.Fatalf("archive did not freeze historical access: %v", err)
	}

	messageCollection, _ := fixture.app.FindCollectionByNameOrId("chat_messages")
	message := core.NewRecord(messageCollection)
	message.Set("room", general.Id)
	message.Set("message_kind", "message")
	message.Set("sender_type", "player")
	message.Set("sender_id", fixture.profiles[0].Id)
	message.Set("sender_participant", fixture.participants[0].Id)
	message.Set("sender_label_snapshot", "Alice")
	message.Set("content", "Hello")
	if err := fixture.app.Save(message); err != nil {
		t.Fatal(err)
	}
	itemCollection, _ := fixture.app.FindCollectionByNameOrId("attention_items")
	item := core.NewRecord(itemCollection)
	item.Set("game", fixture.game.Id)
	item.Set("kind", "announcement")
	item.Set("sender", fixture.gameMaster.Id)
	item.Set("sender_label_snapshot", "Host")
	item.Set("content", "Attention")
	item.Set("audience", "player")
	item.Set("target_id", fixture.participants[0].Id)
	if err := fixture.app.Save(item); err != nil {
		t.Fatal(err)
	}
	receiptCollection, _ := fixture.app.FindCollectionByNameOrId("attention_receipts")
	receipt := core.NewRecord(receiptCollection)
	receipt.Set("attention_item", item.Id)
	receipt.Set("participant", fixture.participants[0].Id)
	if err := fixture.app.Save(receipt); err != nil {
		t.Fatal(err)
	}

	if err := ClearGameSession(fixture.app, fixture.game.Id); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		collection string
		filter     string
		params     dbx.Params
	}{
		{"chat_rooms", "game = {:game}", dbx.Params{"game": fixture.game.Id}},
		{"chat_memberships", "room = {:room}", dbx.Params{"room": general.Id}},
		{"chat_messages", "room = {:room}", dbx.Params{"room": general.Id}},
		{"attention_items", "game = {:game}", dbx.Params{"game": fixture.game.Id}},
		{"attention_receipts", "attention_item = {:item}", dbx.Params{"item": item.Id}},
	} {
		if count := countRecords(t, fixture.app, check.collection, check.filter, check.params); count != 0 {
			t.Fatalf("%s still contains %d game-owned records", check.collection, count)
		}
	}
}

func findMembership(app core.App, roomID, participantID string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"chat_memberships",
		"room = {:room} && participant = {:participant}",
		"",
		1,
		0,
		dbx.Params{"room": roomID, "participant": participantID},
	)
	if err != nil {
		return nil, err
	}
	return records[0], nil
}

func countRecords(t *testing.T, app core.App, collection, filter string, params dbx.Params) int {
	t.Helper()
	records, err := app.FindRecordsByFilter(collection, filter, "", 100, 0, params)
	if err != nil {
		t.Fatal(err)
	}
	return len(records)
}
