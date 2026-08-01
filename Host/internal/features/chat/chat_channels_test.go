package chat

import (
	"errors"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

func TestPrepareRoleRoomsFreezesCustomChannelReaders(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.Channels = []rulesets.ChatChannel{{
		ID: "red_council", Name: "Red council",
		ReaderTeamIDs: []string{"red"}, SenderRoleIDs: []string{"red-one"},
		MessageRestriction: rulesets.ChatEmojiOnly, Visible: true, Sendable: true,
		GameMasterMaySend: true, SenderDisplay: rulesets.SenderRoleLabel,
		PhaseOverrides: map[string]rulesets.ChatChannelPhaseOverride{},
	}}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}

	if err := PrepareRoleRooms(fixture.app, fixture.game.Id, fixture.definition, fixture.participants); err != nil {
		t.Fatal(err)
	}
	room, err := findRoomByKey(fixture.app, fixture.game.Id, rulesets.CustomChatRoomPrefix+"red_council")
	if err != nil || room.GetString("kind") != "custom" {
		t.Fatalf("custom room not created: %v, %#v", err, room)
	}
	memberships, err := fixture.app.FindRecordsByFilter(
		"chat_memberships",
		"room = {:room}",
		"",
		10,
		0,
		dbx.Params{"room": room.Id},
	)
	if err != nil || len(memberships) != 2 {
		t.Fatalf("reader memberships = %d, %v", len(memberships), err)
	}
	if !customChannelSenderAllowed(fixture.definition, room, fixture.participants[0]) {
		t.Fatal("allowed role could not send")
	}
	if customChannelSenderAllowed(fixture.definition, room, fixture.participants[1]) {
		t.Fatal("reader-only role could send")
	}
}

func TestPrepareRoleRoomsRollsBackWithOwningGameTransaction(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.Teams = map[string]rulesets.RoomPermission{"red": {Visible: true, Readable: true}}
	err := fixture.app.RunInTransaction(func(tx core.App) error {
		if err := PrepareRoleRooms(tx, fixture.game.Id, fixture.definition, fixture.participants); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected rollback")
	}
	rooms, err := fixture.app.FindRecordsByFilter("chat_rooms", "game = {:game} && room_key = 'team:red'", "", 10, 0, dbx.Params{"game": fixture.game.Id})
	if err != nil || len(rooms) != 0 {
		t.Fatalf("role room persisted after rollback: %d %v", len(rooms), err)
	}
}
