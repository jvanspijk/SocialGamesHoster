package realtimeauth

import (
	"errors"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
)

func TestGameMasterTopicRemainsDistinctFromPublicGameTopic(t *testing.T) {
	kind, id, ok := splitTopic("game:game123:game-masters")
	if !ok || kind != "game-master" || id != "game123" {
		t.Fatalf("unexpected game-master topic: %q %q %t", kind, id, ok)
	}
	kind, _, ok = splitTopic("game:game123:public")
	if !ok || kind != "game" {
		t.Fatalf("unexpected public topic: %q %t", kind, ok)
	}
}

func TestRejectsUnknownRealtimeTopics(t *testing.T) {
	for _, topic := range []string{"game:game123:private", "game:game123:game-masters:extra", "room:", "collections/games"} {
		if _, _, ok := splitTopic(topic); ok {
			t.Fatalf("expected %q to be rejected", topic)
		}
	}
}

func TestProfileRequestCapabilityTopicParsing(t *testing.T) {
	capability := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	requestID, actual, ok := splitProfileRequestTopic("profile-request:request123:" + capability)
	if !ok || requestID != "request123" || actual != capability {
		t.Fatalf("unexpected capability topic: %q %q %t", requestID, actual, ok)
	}
	if _, _, ok := splitProfileRequestTopic("profile-request:request123:short"); ok {
		t.Fatal("expected malformed capability to be rejected")
	}
}

func TestRealtimeAuthorizationUsesCurrentMembershipAndHistoricalRoomAccess(t *testing.T) {
	app, game, gameMaster, profiles, participants, room := authorizationFixture(t)

	for _, topic := range []string{
		"game:" + game.Id + ":public",
		"game:" + game.Id + ":game-masters",
		"participant:" + participants[0].Id + ":private",
		"room:" + room.Id,
		"profile:" + profiles[0].Id,
		"profile-requests:game-masters",
	} {
		if !canSubscribe(app, gameMaster, topic) {
			t.Errorf("active game master was denied %q", topic)
		}
	}

	tests := []struct {
		status          gamepolicy.ParticipantStatus
		gameAllowed     bool
		participantOpen bool
		roomAllowed     bool
	}{
		{status: gamepolicy.ParticipantActive, gameAllowed: true, participantOpen: true, roomAllowed: true},
		{status: gamepolicy.ParticipantEliminated, gameAllowed: true, participantOpen: true, roomAllowed: true},
		{status: gamepolicy.ParticipantKicked},
		{status: gamepolicy.ParticipantLeft, roomAllowed: true},
	}
	for index, test := range tests {
		t.Run(string(test.status), func(t *testing.T) {
			profile := profiles[index]
			participant := participants[index]
			if got := canSubscribe(app, profile, "profile:"+profile.Id); !got {
				t.Error("player was denied their own profile topic")
			}
			if got := canSubscribe(app, profile, "game:"+game.Id+":public"); got != test.gameAllowed {
				t.Errorf("game topic allowed = %t, want %t", got, test.gameAllowed)
			}
			if got := canSubscribe(app, profile, "participant:"+participant.Id+":private"); got != test.participantOpen {
				t.Errorf("participant topic allowed = %t, want %t", got, test.participantOpen)
			}
			if got := canSubscribe(app, profile, "room:"+room.Id); got != test.roomAllowed {
				t.Errorf("room topic allowed = %t, want %t", got, test.roomAllowed)
			}
		})
	}
}

func TestRealtimeAuthorizationFailsClosedForMissingRecords(t *testing.T) {
	app, _, _, profiles, _, _ := authorizationFixture(t)
	for _, topic := range []string{
		"game:missing:public",
		"participant:missing:private",
		"room:missing",
		"profile:someone-else",
	} {
		if canSubscribe(app, profiles[0], topic) {
			t.Errorf("missing record unexpectedly authorized for %q", topic)
		}
	}
}

func TestRealtimeAuthorizationFailsClosedWhenDatabaseIsUnavailable(t *testing.T) {
	app, game, _, profiles, _, _ := authorizationFixture(t)
	if canSubscribe(failingQueryApp{App: app}, profiles[0], "game:"+game.Id+":public") {
		t.Fatal("database failure unexpectedly authorized a game topic")
	}
}

type failingQueryApp struct {
	core.App
}

func (app failingQueryApp) FindRecordsByFilter(
	collection any,
	filter string,
	sort string,
	limit int,
	offset int,
	params ...dbx.Params,
) ([]*core.Record, error) {
	return nil, errors.New("query unavailable")
}

func authorizationFixture(t *testing.T) (
	core.App,
	*core.Record,
	*core.Record,
	[]*core.Record,
	[]*core.Record,
	*core.Record,
) {
	t.Helper()
	app := testutil.NewPocketBaseApp(t)

	gameMasters, err := app.FindCollectionByNameOrId("game_masters")
	if err != nil {
		t.Fatal(err)
	}
	gameMaster := core.NewRecord(gameMasters)
	gameMaster.Set("username", "host")
	gameMaster.Set("display_name", "Host")
	gameMaster.Set("is_owner", true)
	gameMaster.Set("active", true)
	gameMaster.SetPassword("secret-password")
	if err := app.Save(gameMaster); err != nil {
		t.Fatal(err)
	}

	profileCollection, err := app.FindCollectionByNameOrId("player_profiles")
	if err != nil {
		t.Fatal(err)
	}
	statuses := []gamepolicy.ParticipantStatus{
		gamepolicy.ParticipantActive,
		gamepolicy.ParticipantEliminated,
		gamepolicy.ParticipantKicked,
		gamepolicy.ParticipantLeft,
	}
	profiles := make([]*core.Record, len(statuses))
	for index, status := range statuses {
		profile := core.NewRecord(profileCollection)
		profile.Set("display_name", "Player "+string(rune('A'+index)))
		profile.Set("normalized_name", "player "+string(rune('a'+index)))
		profile.Set("active", true)
		profile.SetPassword("device-secret-" + string(status))
		if err := app.Save(profile); err != nil {
			t.Fatal(err)
		}
		profiles[index] = profile
	}

	rulesetCollection, _ := app.FindCollectionByNameOrId("rulesets")
	ruleset := core.NewRecord(rulesetCollection)
	ruleset.Set("slug", "authorization-test")
	ruleset.Set("name", "Authorization test")
	ruleset.Set("created_by", gameMaster.Id)
	if err := app.Save(ruleset); err != nil {
		t.Fatal(err)
	}
	versionCollection, _ := app.FindCollectionByNameOrId("ruleset_versions")
	version := core.NewRecord(versionCollection)
	version.Set("ruleset", ruleset.Id)
	version.Set("version_number", 1)
	version.Set("state", "published")
	version.Set("schema_version", 1)
	version.Set("definition", map[string]any{"schemaVersion": 1})
	version.Set("created_by", gameMaster.Id)
	if err := app.Save(version); err != nil {
		t.Fatal(err)
	}
	gameCollection, _ := app.FindCollectionByNameOrId("games")
	game := core.NewRecord(gameCollection)
	game.Set("name", "Authorization test")
	game.Set("status", gamepolicy.GameRunning)
	game.Set("ruleset_version", version.Id)
	game.Set("ruleset_snapshot", map[string]any{"schemaVersion": 1})
	game.Set("timer_state", "inactive")
	game.Set("created_by", gameMaster.Id)
	if err := app.Save(game); err != nil {
		t.Fatal(err)
	}

	participantCollection, _ := app.FindCollectionByNameOrId("participants")
	participants := make([]*core.Record, len(statuses))
	for index, status := range statuses {
		participant := core.NewRecord(participantCollection)
		participant.Set("game", game.Id)
		participant.Set("profile", profiles[index].Id)
		participant.Set("display_name_snapshot", profiles[index].GetString("display_name"))
		participant.Set("seat_number", index+1)
		participant.Set("status", status)
		participant.Set("outcome", "unset")
		participant.Set("joined_at", time.Now().UTC())
		if err := app.Save(participant); err != nil {
			t.Fatal(err)
		}
		participants[index] = participant
	}

	roomCollection, _ := app.FindCollectionByNameOrId("chat_rooms")
	room := core.NewRecord(roomCollection)
	room.Set("game", game.Id)
	room.Set("room_key", "general")
	room.Set("kind", "general")
	room.Set("label", "General")
	room.Set("manual_visibility_override", "default")
	room.Set("sender_display", "profile_name")
	if err := app.Save(room); err != nil {
		t.Fatal(err)
	}
	membershipCollection, _ := app.FindCollectionByNameOrId("chat_memberships")
	for index, participant := range participants {
		membership := core.NewRecord(membershipCollection)
		membership.Set("room", room.Id)
		membership.Set("participant", participant.Id)
		membership.Set("joined_at", time.Now().UTC())
		if statuses[index] == gamepolicy.ParticipantLeft {
			membership.Set("left_at", time.Now().UTC())
			membership.Set("historical_access", true)
		}
		if err := app.Save(membership); err != nil {
			t.Fatal(err)
		}
	}
	return app, game, gameMaster, profiles, participants, room
}
