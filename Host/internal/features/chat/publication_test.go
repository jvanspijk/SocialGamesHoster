package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/domain/gamecapacity"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
)

func TestRoomPublicationChecksEachRecipientsCurrentAccess(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.General = &rulesets.RoomPermission{Visible: true, Readable: true, Sendable: true}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLobbyRoom(fixture.app, fixture.game.Id, fixture.definition); err != nil {
		t.Fatal(err)
	}
	for _, participant := range fixture.participants[:2] {
		if err := AddParticipant(fixture.app, fixture.game.Id, participant); err != nil {
			t.Fatal(err)
		}
	}
	room, err := findRoomByKey(fixture.app, fixture.game.Id, "general")
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveAccess(&core.RequestEvent{App: fixture.app, Auth: fixture.profiles[0]}, room.Id)
	if err != nil {
		t.Fatal(err)
	}
	authorize := roomEventAuthorization(fixture.app, resolved, true)
	for _, recipient := range []*core.Record{fixture.profiles[0], fixture.profiles[1], fixture.gameMaster} {
		if !authorize(recipient) {
			t.Fatalf("eligible recipient %s denied", recipient.Id)
		}
	}
	if authorize(nil) || authorize(fixture.profiles[2]) {
		t.Fatal("anonymous or nonmember recipient allowed")
	}
	fixture.profiles[1].Set("active", false)
	if authorize(fixture.profiles[1]) {
		t.Fatal("inactive recipient allowed")
	}
	fixture.profiles[1].Set("active", true)
	for _, test := range []struct {
		status  string
		allowed bool
	}{
		{status: "active", allowed: true},
		{status: "eliminated", allowed: true},
		{status: "kicked"},
		{status: "left"},
	} {
		fixture.participants[1].Set("status", test.status)
		if err := fixture.app.Save(fixture.participants[1]); err != nil {
			t.Fatal(err)
		}
		authorize = roomEventAuthorization(fixture.app, resolved, true)
		if got := authorize(fixture.profiles[1]); got != test.allowed {
			t.Fatalf("recipient status %s: allowed = %t, want %t", test.status, got, test.allowed)
		}
	}
	if err := CloseParticipantMemberships(fixture.app, fixture.participants[1].Id, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	authorize = roomEventAuthorization(fixture.app, resolved, true)
	if authorize(fixture.profiles[1]) {
		t.Fatal("removed recipient without historical access allowed")
	}
	memberships, err := fixture.app.FindRecordsByFilter("chat_memberships", "room = {:room} && participant = {:participant}", "", 1, 0,
		dbx.Params{"room": room.Id, "participant": fixture.participants[1].Id})
	if err != nil || len(memberships) != 1 {
		t.Fatalf("membership lookup: %v, count %d", err, len(memberships))
	}
	memberships[0].Set("historical_access", true)
	if err := fixture.app.Save(memberships[0]); err != nil {
		t.Fatal(err)
	}
	authorize = roomEventAuthorization(fixture.app, resolved, true)
	if !authorize(fixture.profiles[1]) {
		t.Fatal("historical reader denied")
	}
	// A room membership must not grant access through a participant in another game.
	gameCollection, err := fixture.app.FindCollectionByNameOrId("games")
	if err != nil {
		t.Fatal(err)
	}
	otherGame := core.NewRecord(gameCollection)
	otherGame.Set("name", "Other historical game")
	otherGame.Set("status", "archived")
	otherGame.Set("ruleset_version", fixture.game.GetString("ruleset_version"))
	otherGame.Set("ruleset_snapshot", fixture.definition)
	otherGame.Set("timer_state", "inactive")
	otherGame.Set("created_by", fixture.gameMaster.Id)
	if err := fixture.app.Save(otherGame); err != nil {
		t.Fatal(err)
	}
	fixture.participants[2].Set("game", otherGame.Id)
	if err := fixture.app.Save(fixture.participants[2]); err != nil {
		t.Fatal(err)
	}
	if err := ensureChatMembership(fixture.app, room.Id, fixture.participants[2].Id); err != nil {
		t.Fatal(err)
	}
	authorize = roomEventAuthorization(fixture.app, resolved, true)
	if authorize(fixture.profiles[2]) {
		t.Fatal("participant from another game received room access")
	}
	// Each publication resolves the room's policy for the current phase.
	disabled := false
	fixture.definition.Chat.PhaseOverrides = map[string]rulesets.ChatPolicyOverride{
		"hidden": {General: &rulesets.PartialRoomPermission{Visible: &disabled, Readable: &disabled}},
	}
	fixture.game.Set("phase_key", "hidden")
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	authorize = roomEventAuthorization(fixture.app, resolved, true)
	if authorize(fixture.profiles[0]) || !authorize(fixture.gameMaster) {
		t.Fatal("phase visibility policy or game-master access not preserved")
	}
}

// Publication must stay bounded at capacity, including recipients with historical
// access; checking the resulting permissions must not execute more SQL.
func TestRoomPublicationDatabaseReadsStayBoundedAtCapacity(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.General = &rulesets.RoomPermission{Visible: true, Readable: true}
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
	for _, participant := range fixture.participants {
		if err := ensureChatMembership(fixture.app, room.Id, participant.Id); err != nil {
			t.Fatal(err)
		}
	}
	resolved := access{Game: fixture.game, Room: room}
	db := fixture.app.ConcurrentDB().(*dbx.DB)
	previous := db.QueryLogFunc
	defer func() { db.QueryLogFunc = previous }()
	queries := 0
	db.QueryLogFunc = func(_ context.Context, _ time.Duration, _ string, _ *sql.Rows, _ error) { queries++ }
	authorize := roomEventAuthorization(fixture.app, resolved, true)
	initialQueries := queries
	for _, profile := range fixture.profiles {
		if !authorize(profile) {
			t.Fatal("eligible player denied")
		}
	}
	if queries != initialQueries {
		t.Fatal("recipient checks executed SQL")
	}
	db.QueryLogFunc = previous
	profileCollection, _ := fixture.app.FindCollectionByNameOrId("player_profiles")
	participantCollection, _ := fixture.app.FindCollectionByNameOrId("participants")
	for index := len(fixture.profiles); index < gamecapacity.MaxPlayers; index++ {
		profile := core.NewRecord(profileCollection)
		profile.Set("display_name", fmt.Sprintf("Player %d", index))
		profile.Set("normalized_name", fmt.Sprintf("player %d", index))
		profile.Set("active", true)
		// These disposable records only exercise delivery authorization, not login.
		if err := fixture.app.SaveNoValidate(profile); err != nil {
			t.Fatal(err)
		}
		participant := core.NewRecord(participantCollection)
		participant.Set("game", fixture.game.Id)
		participant.Set("profile", profile.Id)
		participant.Set("player_number", index+1)
		participant.Set("status", "active")
		if err := fixture.app.SaveNoValidate(participant); err != nil {
			t.Fatal(err)
		}
		if err := ensureChatMembership(fixture.app, room.Id, participant.Id); err != nil {
			t.Fatal(err)
		}
		fixture.profiles = append(fixture.profiles, profile)
	}
	queries = 0
	db.QueryLogFunc = func(_ context.Context, _ time.Duration, _ string, _ *sql.Rows, _ error) { queries++ }
	authorize = roomEventAuthorization(fixture.app, resolved, true)
	capacityQueries := queries
	for _, profile := range fixture.profiles {
		if !authorize(profile) {
			t.Fatalf("eligible player %s denied at capacity", profile.Id)
		}
	}
	if !authorize(fixture.gameMaster) {
		t.Fatal("game master denied at capacity")
	}
	if queries != capacityQueries {
		t.Fatal("recipient checks executed SQL at capacity")
	}
	if capacityQueries != initialQueries || capacityQueries > 6 {
		t.Fatalf("authorization queries grew with recipients: initial=%d, capacity=%d", initialQueries, capacityQueries)
	}
	t.Logf("authorization SELECTs: 3 players=%d; %d players + GM=%d; recipient checks=0", initialQueries, gamecapacity.MaxPlayers, capacityQueries)
}

type failingPublicationApp struct {
	core.App
	collection string
}

func (app failingPublicationApp) FindRecordsByFilter(collection any, filter, sort string, limit, offset int, params ...dbx.Params) ([]*core.Record, error) {
	if collection == app.collection {
		return nil, errors.New("bulk authorization read failed")
	}
	return app.App.FindRecordsByFilter(collection, filter, sort, limit, offset, params...)
}

func TestRoomPublicationFailsClosedOnBulkReadFailure(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.General = &rulesets.RoomPermission{Visible: true, Readable: true}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLobbyRoom(fixture.app, fixture.game.Id, fixture.definition); err != nil {
		t.Fatal(err)
	}
	if err := AddParticipant(fixture.app, fixture.game.Id, fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	room, err := findRoomByKey(fixture.app, fixture.game.Id, "general")
	if err != nil {
		t.Fatal(err)
	}
	resolved := access{Game: fixture.game, Room: room}
	for _, collection := range []string{"participants", "chat_memberships"} {
		authorize := roomEventAuthorization(failingPublicationApp{App: fixture.app, collection: collection}, resolved, true)
		if authorize(fixture.profiles[0]) {
			t.Fatalf("player allowed after failed %s read", collection)
		}
		if !authorize(fixture.gameMaster) {
			t.Fatal("game master access depends on player membership")
		}
	}
}

// Capture the actual publisher output without a network or blocking channel.
type recordingRoomClient struct {
	*subscriptions.DefaultClient
	messages []subscriptions.Message
}

func (client *recordingRoomClient) Send(message subscriptions.Message) {
	client.messages = append(client.messages, message)
}

func TestRoomAccessAndPublicationFollowCurrentRole(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.Teams = map[string]rulesets.RoomPermission{
		"red": {Visible: true, Readable: true}, "blue": {Visible: true, Readable: true},
	}
	fixture.definition.Chat.Channels = []rulesets.ChatChannel{
		{ID: "role", Name: "Role", ReaderRoleIDs: []string{"red-one"}, Visible: true},
		{ID: "team", Name: "Team", ReaderTeamIDs: []string{"red"}, Visible: true},
		{ID: "open", Name: "Open", Visible: true},
	}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := PrepareRoleRooms(fixture.app, fixture.game.Id, fixture.definition, fixture.participants); err != nil {
		t.Fatal(err)
	}
	keys := []string{"team:red", "team:blue", "custom:role", "custom:team", "custom:open"}
	rooms := make([]*core.Record, len(keys))
	player := &recordingRoomClient{DefaultClient: subscriptions.NewDefaultClient()}
	player.Set(apis.RealtimeClientAuthKey, fixture.profiles[0])
	gm := &recordingRoomClient{DefaultClient: subscriptions.NewDefaultClient()}
	gm.Set(apis.RealtimeClientAuthKey, fixture.gameMaster)
	for index, key := range keys {
		room, err := findRoomByKey(fixture.app, fixture.game.Id, key)
		if err != nil {
			t.Fatal(err)
		}
		rooms[index] = room
		// Keep subscriptions even when access is revoked, as a connected client may.
		player.Subscribe("room:" + room.Id)
		gm.Subscribe("room:" + room.Id)
	}
	fixture.app.SubscriptionsBroker().Register(player)
	fixture.app.SubscriptionsBroker().Register(gm)
	defer fixture.app.SubscriptionsBroker().Unregister(player.Id())
	defer fixture.app.SubscriptionsBroker().Unregister(gm.Id())
	messagesCollection, _ := fixture.app.FindCollectionByNameOrId("chat_messages")
	message := core.NewRecord(messagesCollection)
	message.Id = core.GenerateDefaultRandomId()
	message.Set("content", "Private message")
	message.Set("sender_id", fixture.profiles[0].Id)
	for _, test := range []struct {
		role    string
		allowed []bool
	}{
		{role: "red-one", allowed: []bool{true, false, true, true, true}},
		{role: "blue-role", allowed: []bool{false, true, false, false, true}},
		{role: "", allowed: []bool{false, false, false, false, true}},
	} {
		fixture.participants[0].Set("role_key", test.role)
		if err := fixture.app.Save(fixture.participants[0]); err != nil {
			t.Fatal(err)
		}
		listed, err := VisibleRoomsForPlayer(fixture.app, fixture.game, fixture.participants[0], fixture.definition)
		if err != nil {
			t.Fatal(err)
		}
		visible := map[string]bool{}
		for _, room := range listed {
			visible[room["id"].(string)] = true
		}
		for index, room := range rooms {
			want := test.allowed[index]
			_, err := resolveAccess(&core.RequestEvent{App: fixture.app, Auth: fixture.profiles[0]}, room.Id)
			if (err == nil) != want {
				t.Fatalf("role %q room %s: read access=%t, want %t", test.role, keys[index], err == nil, want)
			}
			if visible[room.Id] != want {
				t.Fatalf("role %q room %s: listing=%t, want %t", test.role, keys[index], visible[room.Id], want)
			}
			player.messages = nil
			gm.messages = nil
			message.Set("room", room.Id)
			publishRoomMessage(fixture.app, access{Game: fixture.game, Room: room}, "chat.message_created", message)
			if (len(player.messages) == 1) != want {
				t.Fatalf("role %q room %s: delivered=%d, want allowed=%t", test.role, keys[index], len(player.messages), want)
			}
			if len(gm.messages) != 1 {
				t.Fatal("game master missed message")
			}
			if want {
				var event struct {
					Payload map[string]any `json:"payload"`
				}
				if err := json.Unmarshal(player.messages[0].Data, &event); err != nil {
					t.Fatal(err)
				}
				if event.Payload["content"] != "Private message" || event.Payload["isOwn"] != true || event.Payload["senderParticipantId"] != nil {
					t.Fatalf("unsafe or incorrect player projection: %#v", event.Payload)
				}
			}
		}
	}
}

func TestVisibleButUnreadableRoomDoesNotPublishMessageContent(t *testing.T) {
	fixture := newAttentionFixture(t)
	fixture.definition.Chat.DefaultPolicy.General = &rulesets.RoomPermission{Visible: true, Readable: false}
	fixture.game.Set("ruleset_snapshot", fixture.definition)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLobbyRoom(fixture.app, fixture.game.Id, fixture.definition); err != nil {
		t.Fatal(err)
	}
	if err := AddParticipant(fixture.app, fixture.game.Id, fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	room, err := findRoomByKey(fixture.app, fixture.game.Id, "general")
	if err != nil {
		t.Fatal(err)
	}
	resolved := access{Game: fixture.game, Room: room}
	if !roomEventAuthorization(fixture.app, resolved, false)(fixture.profiles[0]) {
		t.Fatal("visible room metadata denied")
	}
	if roomEventAuthorization(fixture.app, resolved, true)(fixture.profiles[0]) {
		t.Fatal("visible but unreadable room allowed message content")
	}
}
