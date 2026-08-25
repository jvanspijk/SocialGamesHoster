package chat

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
)

type attentionFixture struct {
	app          core.App
	game         *core.Record
	gameMaster   *core.Record
	profiles     []*core.Record
	participants []*core.Record
	definition   rulesets.DefinitionV1
}

func TestResolveAudioCuePayload(t *testing.T) {
	game := core.NewRecord(&core.Collection{})
	game.Set("ruleset_version", "version-1")
	cue := rulesets.AudioCue{ID: "bell", Name: "Bell", AssetKey: "bell-audio"}

	t.Run("success", func(t *testing.T) {
		asset := core.NewRecord(&core.Collection{})
		asset.Id = "asset-1"
		payload, err := resolveAudioCuePayload(
			audioCueLookupApp{records: []*core.Record{asset}},
			game,
			cue,
		)
		if err != nil {
			t.Fatal(err)
		}
		assertAudioCuePayload(t, payload, map[string]any{
			"cueKey":  "bell",
			"name":    "Bell",
			"assetId": "asset-1",
			"preview": "/api/app/v1/ruleset-assets/asset-1",
		})
	})

	t.Run("absent", func(t *testing.T) {
		payload, err := resolveAudioCuePayload(audioCueLookupApp{}, game, cue)
		if err != nil {
			t.Fatal(err)
		}
		if payload != nil {
			t.Fatalf("payload = %#v, want nil", payload)
		}
	})

	t.Run("query failure", func(t *testing.T) {
		queryErr := errors.New("query failed")
		payload, err := resolveAudioCuePayload(audioCueLookupApp{err: queryErr}, game, cue)
		if !errors.Is(err, queryErr) {
			t.Fatalf("error = %v, want %v", err, queryErr)
		}
		if payload != nil {
			t.Fatalf("payload = %#v, want nil", payload)
		}
	})
}

type audioCueLookupApp struct {
	core.App
	records []*core.Record
	err     error
}

func (app audioCueLookupApp) FindRecordsByFilter(
	collection any,
	filter string,
	sort string,
	limit int,
	offset int,
	params ...dbx.Params,
) ([]*core.Record, error) {
	return app.records, app.err
}

func assertAudioCuePayload(t *testing.T, actual, expected map[string]any) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("payload = %#v, want %#v", actual, expected)
	}
	for key, expectedValue := range expected {
		if actual[key] != expectedValue {
			t.Fatalf("payload[%q] = %#v, want %#v", key, actual[key], expectedValue)
		}
	}
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

	projected, err := UnacknowledgedAttentionForParticipant(
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
	projected, err = UnacknowledgedAttentionForParticipant(
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

func TestAnnouncementAcknowledgementRemainsAvailableAfterArchive(t *testing.T) {
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
	item.Set("content", "Archived announcement")
	item.Set("audience", "player")
	item.Set("target_id", fixture.participants[0].Id)
	if err := fixture.app.Save(item); err != nil {
		t.Fatal(err)
	}
	receiptCollection, err := fixture.app.FindCollectionByNameOrId("attention_receipts")
	if err != nil {
		t.Fatal(err)
	}
	receipt := core.NewRecord(receiptCollection)
	receipt.Set("attention_item", item.Id)
	receipt.Set("participant", fixture.participants[0].Id)
	receipt.Set("acknowledged_at", time.Now().UTC())
	if err := fixture.app.Save(receipt); err != nil {
		t.Fatal(err)
	}
	fixture.game.Set("status", gamepolicy.GameArchived)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/games/"+fixture.game.Id+"/announcements/"+item.Id+"/acknowledge", nil)
	request.SetPathValue("id", fixture.game.Id)
	request.SetPathValue("announcementId", item.Id)
	event := &core.RequestEvent{}
	event.App = fixture.app
	event.Auth = fixture.profiles[0]
	event.Request = request
	event.Response = recorder

	if err := acknowledgeAnnouncement(event); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestOneOffAnnouncementUploadsArePrivateAndAnnouncementOwned(t *testing.T) {
	fixture := newAttentionFixture(t)
	fields := map[string]string{
		"content": "Look and listen", "audience": "player", "targetId": fixture.participants[0].Id,
		"imageDescription": "A red marker on the village map.", "audioAlternative": "A short alert tone.",
	}
	recorder := createMultipartAnnouncement(t, fixture, fields, map[string]uploadFixture{
		"imageFile": {name: "map.png", contentType: "image/png", content: testAnnouncementPNG(t)},
		"audioFile": {name: "alert.wav", contentType: "audio/wav", content: testAnnouncementWAV()},
	})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	items, err := fixture.app.FindRecordsByFilter("attention_items", "game = {:game}", "", 10, 0, dbx.Params{"game": fixture.game.Id})
	if err != nil || len(items) != 1 {
		t.Fatalf("attention items = %d, %v", len(items), err)
	}
	item := items[0]
	if item.GetString("image_asset_key") != "" || item.GetString("audio_asset_key") != "" || item.GetString("image_attachment") == "" || item.GetString("audio_attachment") == "" {
		t.Fatalf("announcement did not retain private attachment relations: %#v", item)
	}
	attachments, err := fixture.app.FindRecordsByFilter("announcement_attachments", "announcement = {:item}", "kind", 10, 0, dbx.Params{"item": item.Id})
	if err != nil || len(attachments) != 2 {
		t.Fatalf("attachments = %d, %v", len(attachments), err)
	}
	for _, attachment := range attachments {
		if attachment.GetString("game") != fixture.game.Id || attachment.GetString("creator") != fixture.gameMaster.Id || attachment.GetString("storage_state") != "ready" || attachment.GetString("checksum") == "" {
			t.Fatalf("incomplete attachment record: %#v", attachment)
		}
	}

	allowed := readAnnouncementAttachment(t, fixture, item, fixture.profiles[0], "image")
	if allowed.Code != http.StatusOK || allowed.Header().Get("Cache-Control") != "private, no-store" || allowed.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("recipient media response = %d, %#v", allowed.Code, allowed.Header())
	}
	denied := readAnnouncementAttachment(t, fixture, item, fixture.profiles[1], "image")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("non-recipient status = %d, body = %s", denied.Code, denied.Body.String())
	}

	fixture.game.Set("status", gamepolicy.GameReview)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	completed := readAnnouncementAttachment(t, fixture, item, fixture.profiles[0], "audio")
	if completed.Code != http.StatusOK {
		t.Fatalf("completed-game recipient media status = %d, body = %s", completed.Code, completed.Body.String())
	}
	rejectedAfterCompletion := createMultipartAnnouncement(t, fixture, map[string]string{
		"content": "Too late", "audience": "all", "imageDescription": "A valid description.",
	}, map[string]uploadFixture{"imageFile": {name: "late.png", contentType: "image/png", content: testAnnouncementPNG(t)}})
	if rejectedAfterCompletion.Code != http.StatusConflict {
		t.Fatalf("completed-game upload status = %d, body = %s", rejectedAfterCompletion.Code, rejectedAfterCompletion.Body.String())
	}

	fixture.game.Set("status", gamepolicy.GameArchived)
	if err := fixture.app.Save(fixture.game); err != nil {
		t.Fatal(err)
	}
	archived := readAnnouncementAttachment(t, fixture, item, fixture.profiles[0], "audio")
	if archived.Code != http.StatusOK {
		t.Fatalf("archived recipient media status = %d, body = %s", archived.Code, archived.Body.String())
	}

	receipts, err := fixture.app.FindRecordsByFilter("attention_receipts", "attention_item = {:item}", "", 10, 0, dbx.Params{"item": item.Id})
	if err != nil {
		t.Fatal(err)
	}
	for _, receipt := range receipts {
		if err := fixture.app.Delete(receipt); err != nil {
			t.Fatal(err)
		}
	}
	if err := fixture.app.Delete(item); err != nil {
		t.Fatal(err)
	}
	if count, err := fixture.app.CountRecords("announcement_attachments"); err != nil || count != 0 {
		t.Fatalf("announcement deletion left %d attachments: %v", count, err)
	}
}

func TestAnnouncementUploadRejectsInvalidSourcesAndCleansFailedWork(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]string
		file   uploadFixture
	}{
		{
			name:   "mixed source",
			fields: map[string]string{"content": "Mixed", "audience": "all", "imageAssetKey": "existing", "imageDescription": "Description"},
			file:   uploadFixture{name: "mixed.png", contentType: "image/png", content: testAnnouncementPNG(t)},
		},
		{
			name:   "missing accessibility text",
			fields: map[string]string{"content": "No description", "audience": "all"},
			file:   uploadFixture{name: "missing.png", contentType: "image/png", content: testAnnouncementPNG(t)},
		},
		{
			name:   "invalid signature",
			fields: map[string]string{"content": "Bad image", "audience": "all", "imageDescription": "Description"},
			file:   uploadFixture{name: "bad.png", contentType: "image/png", content: []byte("not an image")},
		},
		{
			name:   "over size limit",
			fields: map[string]string{"content": "Too large", "audience": "all", "imageDescription": "Description"},
			file:   uploadFixture{name: "large.png", contentType: "image/png", content: make([]byte, rulesets.MediaUploadLimit("image")+1)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newAttentionFixture(t)
			recorder := createMultipartAnnouncement(t, fixture, test.fields, map[string]uploadFixture{"imageFile": test.file})
			if recorder.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
			}
			if count, err := fixture.app.CountRecords("announcement_attachments"); err != nil || count != 0 {
				t.Fatalf("rejected upload left %d records: %v", count, err)
			}
		})
	}

	t.Run("recipient transaction failure", func(t *testing.T) {
		fixture := newAttentionFixture(t)
		recorder := createMultipartAnnouncement(t, fixture, map[string]string{
			"content": "No recipient", "audience": "player", "targetId": "missing", "imageDescription": "A valid description.",
		}, map[string]uploadFixture{"imageFile": {name: "map.png", contentType: "image/png", content: testAnnouncementPNG(t)}})
		if recorder.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
		}
		if count, err := fixture.app.CountRecords("announcement_attachments"); err != nil || count != 0 {
			t.Fatalf("failed transaction left %d staged attachments: %v", count, err)
		}
		if count, err := fixture.app.CountRecords("attention_items"); err != nil || count != 0 {
			t.Fatalf("failed transaction left %d announcements: %v", count, err)
		}
	})
}

type uploadFixture struct {
	name, contentType string
	content           []byte
}

func createMultipartAnnouncement(t *testing.T, fixture attentionFixture, fields map[string]string, files map[string]uploadFixture) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	for field, file := range files {
		part, err := writer.CreateFormFile(field, file.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(file.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/games/"+fixture.game.Id+"/announcements", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.SetPathValue("id", fixture.game.Id)
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{}
	event.App = fixture.app
	event.Auth = fixture.gameMaster
	event.Request = request
	event.Response = recorder
	if err := createAnnouncement(event); err != nil {
		t.Fatal(err)
	}
	return recorder
}

func readAnnouncementAttachment(t *testing.T, fixture attentionFixture, item, actor *core.Record, kind string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/games/"+fixture.game.Id+"/announcements/"+item.Id+"/media/"+kind, nil)
	request.SetPathValue("id", fixture.game.Id)
	request.SetPathValue("announcementId", item.Id)
	request.SetPathValue("kind", kind)
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{}
	event.App = fixture.app
	event.Auth = actor
	event.Request = request
	event.Response = recorder
	if err := announcementMedia(event); err != nil {
		t.Fatal(err)
	}
	return recorder
}

func testAnnouncementPNG(t *testing.T) []byte {
	t.Helper()
	var content bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 1, 1))
	picture.Set(0, 0, color.RGBA{R: 180, A: 255})
	if err := png.Encode(&content, picture); err != nil {
		t.Fatal(err)
	}
	return content.Bytes()
}

func testAnnouncementWAV() []byte {
	data := bytes.Repeat([]byte{128}, 800)
	content := make([]byte, 44+len(data))
	copy(content[0:4], "RIFF")
	binary.LittleEndian.PutUint32(content[4:8], uint32(len(content)-8))
	copy(content[8:12], "WAVE")
	copy(content[12:16], "fmt ")
	binary.LittleEndian.PutUint32(content[16:20], 16)
	binary.LittleEndian.PutUint16(content[20:22], 1)
	binary.LittleEndian.PutUint16(content[22:24], 1)
	binary.LittleEndian.PutUint32(content[24:28], 8000)
	binary.LittleEndian.PutUint32(content[28:32], 8000)
	binary.LittleEndian.PutUint16(content[32:34], 1)
	binary.LittleEndian.PutUint16(content[34:36], 8)
	copy(content[36:40], "data")
	binary.LittleEndian.PutUint32(content[40:44], uint32(len(data)))
	copy(content[44:], data)
	return content
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
	app := testutil.NewPocketBaseApp(t)

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
