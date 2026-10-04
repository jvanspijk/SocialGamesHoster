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

func TestArchiveFreezesFinalRoleAudienceWithoutRestoringOldTeamAccess(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.Teams = map[string]rulesets.RoomPermission{
		"red": {Visible: true, Readable: true}, "blue": {Visible: true, Readable: true},
	}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := PrepareRoleRooms(fixture.app, fixture.game.Id, fixture.definition, fixture.participants); err != nil {
		t.Fatal(err)
	}
	red, err := findRoomByKey(fixture.app, fixture.game.Id, "team:red")
	if err != nil {
		t.Fatal(err)
	}
	blue, err := findRoomByKey(fixture.app, fixture.game.Id, "team:blue")
	if err != nil {
		t.Fatal(err)
	}
	fixture.participants[0].Set("role_key", "blue-role")
	if err := fixture.app.Save(fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	// Match the game aggregate: archive status and grants are written together.
	if err := fixture.app.RunInTransaction(func(tx core.App) error {
		fixture.game.Set("status", "archived")
		if err := tx.Save(fixture.game); err != nil {
			return err
		}
		return FreezeHistoricalAccess(tx, fixture.game.Id, time.Now().UTC())
	}); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		room    *core.Record
		allowed bool
	}{
		{room: red}, {room: blue, allowed: true},
	} {
		resolved, err := resolveAccess(&core.RequestEvent{App: fixture.app, Auth: fixture.profiles[0]}, test.room.Id)
		if (err == nil) != test.allowed {
			t.Fatalf("archived room %s: access=%t, want %t", test.room.GetString("room_key"), err == nil, test.allowed)
		}
		authorize := roomEventAuthorization(fixture.app, access{Game: fixture.game, Room: test.room}, true)
		if authorize(fixture.profiles[0]) != test.allowed {
			t.Fatal("archived publication disagrees with read access")
		}
		if test.allowed && (!resolved.Policy.Readable || resolved.Policy.Sendable) {
			t.Fatal("historical grant must be read-only")
		}
	}
	membership, err := findMembership(fixture.app, blue.Id, fixture.participants[0].Id)
	if err != nil || !membership.GetBool("historical_access") {
		t.Fatalf("newly gained team audience not preserved: %v", err)
	}
	if !membership.GetDateTime("joined_at").Time().Equal(fixture.participants[0].GetDateTime("joined_at").Time().Truncate(time.Millisecond)) {
		t.Fatal("new archive membership would hide messages accessible during live play")
	}
}
