package games

import (
	"testing"

	"github.com/pocketbase/dbx"

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

	if err := prepareRoleRooms(fixture.app, fixture.game); err != nil {
		t.Fatal(err)
	}
	room, err := findRoom(fixture.app, fixture.game.Id, rulesets.CustomChatRoomPrefix+"red_council")
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
