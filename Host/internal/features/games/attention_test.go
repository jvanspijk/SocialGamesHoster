package games

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
	_ "github.com/jvanspijk/SocialGamesHoster/Host/migrations"
)

type attentionFixture struct {
	app          core.App
	game         *core.Record
	gameMaster   *core.Record
	profiles     []*core.Record
	participants []*core.Record
	definition   rulesets.DefinitionV1
}

func TestAnnouncementRecipientsValidateAndFreezeAudience(t *testing.T) {
	fixture := newAttentionFixture(t)

	all, err := resolveAnnouncementRecipients(fixture.app, fixture.game, fixture.definition, "all", "")
	if err != nil || len(all) != 3 {
		t.Fatalf("all recipients = %d, %v", len(all), err)
	}
	team, err := resolveAnnouncementRecipients(fixture.app, fixture.game, fixture.definition, "team", "red")
	if err != nil || len(team) != 2 {
		t.Fatalf("red recipients = %d, %v", len(team), err)
	}
	player, err := resolveAnnouncementRecipients(
		fixture.app,
		fixture.game,
		fixture.definition,
		"player",
		fixture.participants[2].Id,
	)
	if err != nil || len(player) != 1 || player[0].Id != fixture.participants[2].Id {
		t.Fatalf("player recipients = %#v, %v", player, err)
	}

	fixture.participants[0].Set("role_key", "")
	if err := fixture.app.Save(fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	_, err = resolveAnnouncementRecipients(fixture.app, fixture.game, fixture.definition, "team", "red")
	var appError result.AppError
	if err == nil || !errorAs(err, &appError) || appError.Code != "attention.assignments_required" {
		t.Fatalf("expected assignment error, got %v", err)
	}
}

func TestAttentionProjectionRequiresFrozenReceiptAndHidesRecipientIdentities(t *testing.T) {
	fixture := newAttentionFixture(t)
	itemCollection, err := fixture.app.FindCollectionByNameOrId("attention_items")
	if err != nil {
		t.Fatal(err)
	}
	item := core.NewRecord(itemCollection)
	item.Set("game", fixture.game.Id)
	item.Set("kind", "announcement")
	item.Set("sender", fixture.gameMaster.Id)
	item.Set("sender_label_snapshot", "Host")
	item.Set("content", "Secret for the red team")
	item.Set("audience", "team")
	item.Set("target_id", "red")
	item.Set("image_asset_key", "secret_map")
	item.Set("image_description", "A map marking the red team's meeting point.")
	item.Set("audio_asset_key", "secret_signal")
	item.Set("audio_alternative", "Three short bells.")
	if err := fixture.app.Save(item); err != nil {
		t.Fatal(err)
	}

	receiptCollection, err := fixture.app.FindCollectionByNameOrId("attention_receipts")
	if err != nil {
		t.Fatal(err)
	}
	receipts := make([]*core.Record, 2)
	for index, participant := range fixture.participants[:2] {
		receipt := core.NewRecord(receiptCollection)
		receipt.Set("attention_item", item.Id)
		receipt.Set("participant", participant.Id)
		if err := fixture.app.Save(receipt); err != nil {
			t.Fatal(err)
		}
		receipts[index] = receipt
	}

	// A later role change does not rewrite the frozen recipient set.
	fixture.participants[0].Set("role_key", "blue-role")
	if err := fixture.app.Save(fixture.participants[0]); err != nil {
		t.Fatal(err)
	}
	if !actorHasAttentionReceipt(fixture.app, item.Id, fixture.profiles[0]) {
		t.Fatal("frozen recipient lost access after a role change")
	}
	if actorHasAttentionReceipt(fixture.app, item.Id, fixture.profiles[2]) {
		t.Fatal("unrelated player received the attention item")
	}
	if actorHasAttentionReceipt(fixture.app, item.Id, fixture.gameMaster) {
		t.Fatal("game master must not receive the player attention event")
	}

	projected, err := unacknowledgedAttentionForParticipant(
		fixture.app,
		fixture.game.Id,
		fixture.participants[0].Id,
	)
	if err != nil || len(projected) != 1 || projected[0]["content"] != "Secret for the red team" {
		t.Fatalf("unexpected player projection: %#v, %v", projected, err)
	}
	receipts[0].Set("acknowledged_at", time.Now().UTC())
	if err := fixture.app.Save(receipts[0]); err != nil {
		t.Fatal(err)
	}
	projected, err = unacknowledgedAttentionForParticipant(
		fixture.app,
		fixture.game.Id,
		fixture.participants[0].Id,
	)
	if err != nil || len(projected) != 0 {
		t.Fatalf("acknowledged item remained visible: %#v, %v", projected, err)
	}

	summary, err := projectAdminAttentionSummary(fixture.app, item)
	if err != nil {
		t.Fatal(err)
	}
	if summary["recipientTotal"] != 2 || summary["acknowledgementCount"] != 1 {
		t.Fatalf("unexpected aggregate summary: %#v", summary)
	}
	if _, exposesIdentities := summary["recipients"]; exposesIdentities {
		t.Fatal("admin aggregate exposed receipt identities")
	}
	playerImage := projectedAnnouncementMedia(t, projectPlayerAttention(item), "image")
	if playerImage["description"] != "A map marking the red team's meeting point." {
		t.Fatalf("player media description missing: %#v", playerImage)
	}
	adminAudio := projectedAnnouncementMedia(t, summary, "audio")
	if adminAudio["alternative"] != "Three short bells." {
		t.Fatalf("admin audio alternative missing: %#v", adminAudio)
	}
}

func projectedAnnouncementMedia(t *testing.T, projection map[string]any, key string) map[string]any {
	t.Helper()
	media, ok := projection[key].(map[string]any)
	if !ok {
		t.Fatalf("%s media missing from projection: %#v", key, projection)
	}
	return media
}

func newAttentionFixture(t *testing.T) attentionFixture {
	t.Helper()
	if err := os.MkdirAll(".testdata", 0o700); err != nil {
		t.Fatal(err)
	}
	dataDir, err := os.MkdirTemp(".testdata", "attention-*")
	if err != nil {
		t.Fatal(err)
	}
	dataDir, err = filepath.Abs(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	app := core.NewBaseApp(core.BaseAppConfig{
		DataDir: dataDir, EncryptionEnv: "sgh_test_encryption",
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		app.ResetBootstrapState()
		if err := os.RemoveAll(dataDir); err != nil {
			t.Errorf("remove test data: %v", err)
		}
	})
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}

	gameMasterCollection, _ := app.FindCollectionByNameOrId("game_masters")
	gameMaster := core.NewRecord(gameMasterCollection)
	gameMaster.Set("username", "host")
	gameMaster.Set("display_name", "Host")
	gameMaster.Set("is_owner", true)
	gameMaster.Set("active", true)
	gameMaster.SetPassword("secret-password")
	if err := app.Save(gameMaster); err != nil {
		t.Fatal(err)
	}

	profileCollection, _ := app.FindCollectionByNameOrId("player_profiles")
	profiles := make([]*core.Record, 3)
	for index, name := range []string{"Alice", "Boris", "Cleo"} {
		profile := core.NewRecord(profileCollection)
		profile.Set("display_name", name)
		profile.Set("normalized_name", name)
		profile.Set("active", true)
		profile.Set("approved_at", time.Now().UTC())
		profile.SetPassword("device-secret-" + name)
		if err := app.Save(profile); err != nil {
			t.Fatal(err)
		}
		profiles[index] = profile
	}

	definition := rulesets.DefinitionV1{
		SchemaVersion: 1,
		Metadata:      rulesets.Metadata{Name: "Test", MinPlayers: 1, MaxPlayers: 30},
		Teams: []rulesets.Team{
			{ID: "red", Name: "Red"},
			{ID: "blue", Name: "Blue"},
		},
		Roles: []rulesets.Role{
			{ID: "red-one", Name: "Red one", TeamID: "red"},
			{ID: "red-two", Name: "Red two", TeamID: "red"},
			{ID: "blue-role", Name: "Blue", TeamID: "blue"},
		},
	}
	rulesetCollection, _ := app.FindCollectionByNameOrId("rulesets")
	ruleset := core.NewRecord(rulesetCollection)
	ruleset.Set("slug", "attention-test")
	ruleset.Set("name", "Attention test")
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
	version.Set("definition", definition)
	version.Set("created_by", gameMaster.Id)
	version.Set("published_by", gameMaster.Id)
	version.Set("published_at", time.Now().UTC())
	if err := app.Save(version); err != nil {
		t.Fatal(err)
	}
	gameCollection, _ := app.FindCollectionByNameOrId("games")
	game := core.NewRecord(gameCollection)
	game.Set("name", "Attention test")
	game.Set("status", "running")
	game.Set("ruleset_version", version.Id)
	game.Set("ruleset_snapshot", definition)
	game.Set("timer_state", "inactive")
	game.Set("created_by", gameMaster.Id)
	if err := app.Save(game); err != nil {
		t.Fatal(err)
	}

	participantCollection, _ := app.FindCollectionByNameOrId("participants")
	participants := make([]*core.Record, 3)
	for index, profile := range profiles {
		participant := core.NewRecord(participantCollection)
		participant.Set("game", game.Id)
		participant.Set("profile", profile.Id)
		participant.Set("display_name_snapshot", profile.GetString("display_name"))
		participant.Set("seat_number", index+1)
		participant.Set("status", "active")
		participant.Set("role_key", definition.Roles[index].ID)
		participant.Set("outcome", "unset")
		participant.Set("joined_at", time.Now().UTC())
		if err := app.Save(participant); err != nil {
			t.Fatal(err)
		}
		participants[index] = participant
	}
	return attentionFixture{
		app: app, game: game, gameMaster: gameMaster,
		profiles: profiles, participants: participants, definition: definition,
	}
}

func errorAs(err error, target *result.AppError) bool {
	value, ok := err.(result.AppError)
	if ok {
		*target = value
	}
	return ok
}
