package chat

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	platformaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func VisibleRoomsForPlayer(app core.App, game, participant *core.Record, definition rulesets.DefinitionV1) ([]map[string]any, error) {
	memberships, err := app.FindRecordsByFilter("chat_memberships", "participant = {:participant}", "", 200, 0,
		dbx.Params{"participant": participant.Id})
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(memberships))
	for _, membership := range memberships {
		room, err := app.FindRecordById("chat_rooms", membership.GetString("room"))
		if err != nil || room.GetString("game") != game.Id || room.GetString("kind") == "announcements" {
			continue
		}
		base, override := resolveRoomPolicy(definition, game.GetString("phase_key"), room)
		if room.GetString("kind") == "announcements" || room.GetString("kind") == "gm_dm" {
			base = rulesets.RoomPermission{Visible: true, Readable: true, Sendable: room.GetString("kind") == "gm_dm", SenderDisplay: rulesets.SenderProfileName}
			override = nil
		}
		state := ParticipantState{
			IsMember:       membership.GetDateTime("left_at").IsZero(),
			IsActive:       gamepolicy.IsActivePlayer(gamepolicy.ParticipantStatus(participant.GetString("status"))),
			HistoricalRead: membership.GetBool("historical_access"),
		}
		policy := EffectivePolicy(base, override, state, RoomState{
			ManuallyLocked: !room.GetBool("players_can_post"), ManualVisibilityOverride: room.GetString("manual_visibility_override"),
		})
		if room.GetString("kind") == "custom" && !customChannelSenderAllowed(definition, room, participant) {
			policy.Sendable = false
		}
		if game.GetString("status") == string(gamepolicy.GameReview) || gamepolicy.IsArchived(gamepolicy.GameStatus(game.GetString("status"))) {
			policy.Sendable = false
		}
		if game.GetString("status") == string(gamepolicy.GamePaused) && room.GetString("kind") != "gm_dm" {
			explicitlyAllowed := override != nil && override.Sendable != nil && *override.Sendable
			if !explicitlyAllowed {
				policy.Sendable = false
			}
		}
		if !policy.Visible && !policy.Readable {
			continue
		}
		result = append(result, map[string]any{
			"id": room.Id, "key": room.GetString("room_key"), "kind": room.GetString("kind"),
			"label": room.GetString("label"), "visible": policy.Visible, "readable": policy.Readable,
			"sendable": policy.Sendable, "senderDisplay": policy.SenderDisplay,
			"messageRestriction": roomMessageRestriction(definition, room),
			"playersCanPost":     room.GetBool("players_can_post"),
			"latestMessage":      latestMessageSummary(app, room.Id),
		})
	}
	return result, nil
}

type announcementRequest struct {
	Content          string `json:"content"`
	CueKey           string `json:"cueKey"`
	Audience         string `json:"audience"`
	TargetID         string `json:"targetId"`
	ImageAssetKey    string `json:"imageAssetKey"`
	ImageDescription string `json:"imageDescription"`
	AudioAssetKey    string `json:"audioAssetKey"`
	AudioAlternative string `json:"audioAlternative"`
}

func createAnnouncement(event *core.RequestEvent) error {
	game, err := findAnnouncementGame(event)
	if err != nil {
		return writeError(event, err)
	}
	if game.GetString("status") != string(gamepolicy.GameLobby) && game.GetString("status") != string(gamepolicy.GameRunning) && game.GetString("status") != string(gamepolicy.GamePaused) {
		return httpx.WriteError(event, result.Conflict("chat.read_only", "Announcements are only available during a live game."))
	}
	var request announcementRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("attention.invalid_announcement", "Enter an announcement of at most 1000 characters.", nil))
	}
	request.Content = strings.TrimSpace(request.Content)
	if len([]rune(request.Content)) > 1000 || request.Content == "" || containsControlCharacter(request.Content) {
		return httpx.WriteError(event, result.Invalid("attention.invalid_announcement", "Enter an announcement of at most 1000 characters.", nil))
	}
	definition, err := definitionFromGame(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	request.ImageDescription, err = validateAnnouncementAsset(
		event.App, game, definition, request.ImageAssetKey, "image", request.ImageDescription,
	)
	if err != nil {
		return writeError(event, err)
	}
	request.AudioAlternative, err = validateAnnouncementAsset(
		event.App, game, definition, request.AudioAssetKey, "audio", request.AudioAlternative,
	)
	if err != nil {
		return writeError(event, err)
	}
	var cue *rulesets.AudioCue
	if request.CueKey != "" {
		for index := range definition.AudioCues {
			if definition.AudioCues[index].ID == request.CueKey {
				cue = &definition.AudioCues[index]
				break
			}
		}
		if cue == nil {
			return httpx.WriteError(event, result.Invalid("audio.invalid_cue", "That sound cue is not part of this game.", nil))
		}
	}

	var item *core.Record
	var recipients []*core.Record
	err = event.App.RunInTransaction(func(tx core.App) error {
		var resolveErr error
		recipients, resolveErr = resolveAnnouncementRecipients(tx, game, definition, request.Audience, request.TargetID)
		if resolveErr != nil {
			return resolveErr
		}
		itemCollection, collectionErr := tx.FindCollectionByNameOrId("attention_items")
		if collectionErr != nil {
			return collectionErr
		}
		item = core.NewRecord(itemCollection)
		item.Set("game", game.Id)
		item.Set("kind", "announcement")
		item.Set("sender", event.Auth.Id)
		item.Set("sender_label_snapshot", event.Auth.GetString("display_name"))
		item.Set("content", request.Content)
		item.Set("audience", request.Audience)
		item.Set("target_id", request.TargetID)
		item.Set("cue_key", request.CueKey)
		item.Set("image_asset_key", request.ImageAssetKey)
		item.Set("image_description", request.ImageDescription)
		item.Set("audio_asset_key", request.AudioAssetKey)
		item.Set("audio_alternative", request.AudioAlternative)
		if saveErr := tx.Save(item); saveErr != nil {
			return saveErr
		}
		receiptCollection, collectionErr := tx.FindCollectionByNameOrId("attention_receipts")
		if collectionErr != nil {
			return collectionErr
		}
		for _, participant := range recipients {
			receipt := core.NewRecord(receiptCollection)
			receipt.Set("attention_item", item.Id)
			receipt.Set("participant", participant.Id)
			if saveErr := tx.Save(receipt); saveErr != nil {
				return saveErr
			}
		}
		return nil
	})
	if err != nil {
		return writeError(event, err)
	}

	playerProjection := projectPlayerAttention(item)
	_ = realtime.Publish(event.App, "game:"+game.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: "attention.announcement_created", Payload: playerProjection,
	}, func(auth *core.Record) bool {
		return actorHasAttentionReceipt(event.App, item.Id, auth)
	})
	if cue != nil {
		publishAttentionCue(event.App, game, *cue, item.Id)
	}
	_ = platformaudit.Record(event.App, event.Auth, game.Id, "attention.announcement_sent", "attention_item", item.Id,
		map[string]any{"cueKey": request.CueKey, "audience": request.Audience, "targetId": request.TargetID, "recipientTotal": len(recipients)}, event.Get(httpx.TraceIDKey))
	summary, err := projectAdminAttentionSummary(event.App, item)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, summary)
}

func acknowledgeAnnouncement(event *core.RequestEvent) error {
	game, err := findAnnouncementGame(event)
	if err != nil {
		return writeError(event, err)
	}
	// Acknowledgement remains valid after archival because it only updates the
	// authenticated participant's receipt, not archived game-owned content.
	item, err := event.App.FindRecordById("attention_items", event.Request.PathValue("announcementId"))
	if err != nil || item.GetString("game") != game.Id || item.GetString("kind") != "announcement" {
		return httpx.WriteError(event, result.AppError{Code: "attention.not_found", Message: "Announcement not found.", Status: http.StatusNotFound})
	}
	participant, err := gamepolicyapp.CurrentParticipantByGameAndProfile(event.App, game.Id, event.Auth.Id)
	if err != nil {
		return httpx.WriteError(event, result.Forbidden("attention.forbidden", "This announcement is not available."))
	}
	receipts, err := event.App.FindRecordsByFilter(
		"attention_receipts",
		"attention_item = {:item} && participant = {:participant}",
		"",
		1,
		0,
		dbx.Params{"item": item.Id, "participant": participant.Id},
	)
	if err != nil || len(receipts) != 1 {
		return httpx.WriteError(event, result.Forbidden("attention.forbidden", "This announcement is not available."))
	}
	receipt := receipts[0]
	if receipt.GetDateTime("acknowledged_at").IsZero() {
		if err := event.App.RunInTransaction(func(tx core.App) error {
			current, findErr := tx.FindRecordById("attention_receipts", receipt.Id)
			if findErr != nil {
				return findErr
			}
			if current.GetDateTime("acknowledged_at").IsZero() {
				current.Set("acknowledged_at", time.Now().UTC())
				return tx.Save(current)
			}
			return nil
		}); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	summary, err := projectAdminAttentionSummary(event.App, item)
	if err == nil {
		publishAnnouncementGameMasters(event.App, game, "attention.announcement_acknowledged", summary)
	}
	return event.NoContent(http.StatusNoContent)
}

func resolveAnnouncementRecipients(app core.App, game *core.Record, definition rulesets.DefinitionV1, audience, targetID string) ([]*core.Record, error) {
	participants, err := currentAnnouncementParticipants(app, game.Id)
	if err != nil {
		return nil, err
	}
	switch audience {
	case "all":
		if targetID != "" {
			return nil, result.Invalid("attention.invalid_audience", "All players does not accept a target.", nil)
		}
		if len(participants) == 0 {
			return nil, result.Conflict("attention.no_recipients", "There are no active players to receive this announcement.")
		}
		return participants, nil
	case "player":
		for _, participant := range participants {
			if participant.Id == targetID {
				return []*core.Record{participant}, nil
			}
		}
		return nil, result.Invalid("attention.invalid_target", "Choose an active player in this game.", nil)
	case "team":
		teamExists := false
		for _, team := range definition.Teams {
			teamExists = teamExists || team.ID == targetID
		}
		if !teamExists {
			return nil, result.Invalid("attention.invalid_target", "Choose a team from this ruleset.", nil)
		}
		roleTeams := make(map[string]string, len(definition.Roles))
		for _, role := range definition.Roles {
			roleTeams[role.ID] = role.TeamID
		}
		recipients := make([]*core.Record, 0)
		for _, participant := range participants {
			roleKey := participant.GetString("role_key")
			teamID, assigned := roleTeams[roleKey]
			if !assigned || roleKey == "" {
				return nil, result.Conflict("attention.assignments_required", "Assign valid roles before targeting a team.")
			}
			if teamID == targetID {
				recipients = append(recipients, participant)
			}
		}
		if len(recipients) == 0 {
			return nil, result.Conflict("attention.no_recipients", "No active players belong to that team.")
		}
		return recipients, nil
	default:
		return nil, result.Invalid("attention.invalid_audience", "Choose all players, one team, or one player.", nil)
	}
}

func UnacknowledgedAttentionForParticipant(app core.App, gameID, participantID string) ([]map[string]any, error) {
	receipts, err := app.FindRecordsByFilter(
		"attention_receipts",
		"participant = {:participant} && acknowledged_at = ''",
		"",
		200,
		0,
		dbx.Params{"participant": participantID},
	)
	if err != nil {
		return nil, err
	}
	items := make([]*core.Record, 0, len(receipts))
	for _, receipt := range receipts {
		item, findErr := app.FindRecordById("attention_items", receipt.GetString("attention_item"))
		if findErr == nil && item.GetString("game") == gameID && item.GetString("kind") == "announcement" {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(left, right int) bool {
		leftTime := items[left].GetDateTime("created").Time()
		rightTime := items[right].GetDateTime("created").Time()
		if leftTime.Equal(rightTime) {
			return items[left].Id < items[right].Id
		}
		return leftTime.Before(rightTime)
	})
	projected := make([]map[string]any, len(items))
	for index, item := range items {
		projected[index] = projectPlayerAttention(item)
	}
	return projected, nil
}

func projectPlayerAttention(item *core.Record) map[string]any {
	projected := map[string]any{
		"id": item.Id, "kind": "announcement",
		"senderLabel": item.GetString("sender_label_snapshot"),
		"content":     item.GetString("content"), "cueKey": item.GetString("cue_key"),
		"createdAt": recordDateValue(item, "created"),
	}
	projectAnnouncementMedia(item, projected)
	return projected
}

func projectAdminAttentionSummary(app core.App, item *core.Record) (map[string]any, error) {
	receipts, err := app.FindRecordsByFilter(
		"attention_receipts",
		"attention_item = {:item}",
		"",
		200,
		0,
		dbx.Params{"item": item.Id},
	)
	if err != nil {
		return nil, err
	}
	acknowledged := 0
	for _, receipt := range receipts {
		if !receipt.GetDateTime("acknowledged_at").IsZero() {
			acknowledged++
		}
	}
	projected := map[string]any{
		"id": item.Id, "kind": "announcement",
		"senderLabel": item.GetString("sender_label_snapshot"),
		"content":     item.GetString("content"), "audience": item.GetString("audience"),
		"targetId": item.GetString("target_id"), "cueKey": item.GetString("cue_key"),
		"createdAt": recordDateValue(item, "created"), "recipientTotal": len(receipts),
		"acknowledgementCount": acknowledged,
	}
	projectAnnouncementMedia(item, projected)
	return projected, nil
}

func projectAnnouncementMedia(item *core.Record, projected map[string]any) {
	base := "/api/app/v1/games/" + item.GetString("game") + "/announcements/" + item.Id + "/media/"
	if item.GetString("image_asset_key") != "" {
		projected["image"] = map[string]any{"url": base + "image", "description": item.GetString("image_description")}
	}
	if item.GetString("audio_asset_key") != "" {
		projected["audio"] = map[string]any{"url": base + "audio", "alternative": item.GetString("audio_alternative")}
	}
}

func validateAnnouncementAsset(app core.App, game *core.Record, definition rulesets.DefinitionV1, assetKey, kind, suppliedDescription string) (string, error) {
	if assetKey == "" {
		if strings.TrimSpace(suppliedDescription) != "" {
			return "", result.Invalid("attention.invalid_media", "An accessibility description requires an attached asset.", nil)
		}
		return "", nil
	}
	assets, err := app.FindRecordsByFilter(
		"ruleset_assets",
		"ruleset_version = {:version} && asset_key = {:key} && kind = {:kind}",
		"",
		1,
		0,
		dbx.Params{"version": game.GetString("ruleset_version"), "key": assetKey, "kind": kind},
	)
	if err != nil {
		return "", err
	}
	if len(assets) != 1 {
		return "", result.Invalid("attention.invalid_media", "Choose a "+kind+" from this game's ruleset.", nil)
	}
	description := strings.TrimSpace(suppliedDescription)
	if description == "" {
		description = strings.TrimSpace(definition.AssetAccessibility[assetKey].Description)
	}
	limit := 500
	label := "image description"
	if kind == "audio" {
		limit = 1000
		label = "audio alternative"
	}
	if description == "" || len([]rune(description)) > limit || containsControlCharacter(description) {
		return "", result.Invalid("attention.accessibility_required", "Provide an accessible "+label+" for the attachment.", nil)
	}
	return description, nil
}

func announcementMedia(event *core.RequestEvent) error {
	if event.Auth == nil || !event.Auth.GetBool("active") {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "Sign in to view this announcement.", Status: http.StatusUnauthorized})
	}
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
	}
	item, err := event.App.FindRecordById("attention_items", event.Request.PathValue("announcementId"))
	if err != nil || item.GetString("game") != game.Id {
		return httpx.WriteError(event, result.AppError{Code: "attention.not_found", Message: "Announcement not found.", Status: http.StatusNotFound})
	}
	if !actorauth.IsGameMaster(event.Auth) && !actorHasAttentionReceipt(event.App, item.Id, event.Auth) {
		return httpx.WriteError(event, result.Forbidden("attention.forbidden", "This announcement attachment is not available."))
	}
	kind := event.Request.PathValue("kind")
	field := ""
	if kind == "image" {
		field = "image_asset_key"
	} else if kind == "audio" {
		field = "audio_asset_key"
	} else {
		return httpx.WriteError(event, result.AppError{Code: "attention.media_not_found", Message: "Announcement attachment not found.", Status: http.StatusNotFound})
	}
	assetKey := item.GetString(field)
	if assetKey == "" {
		return httpx.WriteError(event, result.AppError{Code: "attention.media_not_found", Message: "Announcement attachment not found.", Status: http.StatusNotFound})
	}
	assets, err := event.App.FindRecordsByFilter(
		"ruleset_assets",
		"ruleset_version = {:version} && asset_key = {:key} && kind = {:kind}",
		"",
		1,
		0,
		dbx.Params{"version": game.GetString("ruleset_version"), "key": assetKey, "kind": kind},
	)
	if err != nil || len(assets) != 1 {
		return httpx.WriteError(event, result.AppError{Code: "attention.media_not_found", Message: "Announcement attachment not found.", Status: http.StatusNotFound})
	}
	asset := assets[0]
	fsys, err := event.App.NewFilesystem()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer fsys.Close()
	reader, err := fsys.GetReader(asset.BaseFilesPath() + "/" + asset.GetString("file"))
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer reader.Close()
	content, err := io.ReadAll(io.LimitReader(reader, (5<<20)+1))
	if err != nil || len(content) > 5<<20 {
		return httpx.WriteError(event, result.Internal(errors.New("announcement media exceeds limit")))
	}
	event.Response.Header().Set("Cache-Control", "private, no-store")
	event.Response.Header().Set("Content-Disposition", `inline; filename="`+filepath.Base(asset.GetString("file"))+`"`)
	return event.Blob(http.StatusOK, asset.GetString("mime_type"), content)
}

func actorHasAttentionReceipt(app core.App, itemID string, auth *core.Record) bool {
	if !actorauth.IsActivePlayer(auth) {
		return false
	}
	receipts, err := app.FindRecordsByFilter(
		"attention_receipts",
		"attention_item = {:item} && participant.profile = {:profile} && "+gamepolicyapp.CurrentRelatedParticipantStatusFilter,
		"",
		1,
		0,
		dbx.Params{"item": itemID, "profile": auth.Id},
	)
	return err == nil && len(receipts) == 1
}

func latestMessageCursor(app core.App, roomID string) any {
	messages, err := app.FindRecordsByFilter(
		"chat_messages",
		"room = {:room} && message_kind != 'announcement'",
		"-created,-id",
		1,
		0,
		dbx.Params{"room": roomID},
	)
	if err != nil || len(messages) == 0 {
		return nil
	}
	return map[string]any{"createdAt": recordDateValue(messages[0], "created"), "id": messages[0].Id}
}

func publishAttentionCue(app core.App, game *core.Record, cue rulesets.AudioCue, itemID string) {
	payload, err := resolveAudioCuePayload(app, game, cue)
	if err != nil || payload == nil {
		return
	}
	_ = realtime.Publish(app, "game:"+game.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: "audio.cue", Payload: payload,
	}, func(auth *core.Record) bool {
		return actorHasAttentionReceipt(app, itemID, auth)
	})
}

func containsControlCharacter(value string) bool {
	for _, character := range value {
		if (character < 0x20 && character != '\n' && character != '\t') || character == 0x7f {
			return true
		}
	}
	return false
}

func validateCueAudience(app core.App, game *core.Record, definition rulesets.DefinitionV1, audience, targetID string) *result.AppError {
	switch audience {
	case "all", "game_masters":
		if targetID != "" {
			value := result.Invalid("audio.invalid_audience", "This audience does not accept a target.", nil)
			return &value
		}
	case "team":
		found := false
		for _, team := range definition.Teams {
			found = found || team.ID == targetID
		}
		if !found {
			value := result.Invalid("audio.invalid_target", "Choose a team from this ruleset.", nil)
			return &value
		}
	case "player":
		participant, err := app.FindRecordById("participants", targetID)
		if err != nil || participant.GetString("game") != game.Id ||
			!gamepolicy.IsCurrentMember(gamepolicy.ParticipantStatus(participant.GetString("status"))) {
			value := result.Invalid("audio.invalid_target", "Choose a player in this game.", nil)
			return &value
		}
	default:
		value := result.Invalid("audio.invalid_audience", "Choose who should hear the sound cue.", nil)
		return &value
	}
	return nil
}

func PublishAudioCue(app core.App, game *core.Record, definition rulesets.DefinitionV1, cue rulesets.AudioCue, audience, targetID string) {
	payload, err := resolveAudioCuePayload(app, game, cue)
	if err != nil || payload == nil {
		return
	}
	_ = realtime.Publish(app, "game:"+game.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: "audio.cue", Payload: payload,
	}, func(auth *core.Record) bool {
		if auth == nil || !auth.GetBool("active") {
			return false
		}
		if audience == "game_masters" {
			return actorauth.IsGameMaster(auth)
		}
		if !actorauth.IsPlayer(auth) {
			return false
		}
		participant, err := gamepolicyapp.CurrentParticipantByGameAndProfile(app, game.Id, auth.Id)
		if err != nil {
			return false
		}
		switch audience {
		case "all":
			return true
		case "player":
			return participant.Id == targetID
		case "team":
			for _, role := range definition.Roles {
				if role.ID == participant.GetString("role_key") {
					return role.TeamID == targetID
				}
			}
		}
		return false
	})
}

func resolveAudioCuePayload(app core.App, game *core.Record, cue rulesets.AudioCue) (map[string]any, error) {
	assets, err := app.FindRecordsByFilter(
		"ruleset_assets",
		"ruleset_version = {:version} && asset_key = {:key} && kind = 'audio'",
		"",
		1,
		0,
		dbx.Params{"version": game.GetString("ruleset_version"), "key": cue.AssetKey},
	)
	if err != nil {
		return nil, err
	}
	if len(assets) != 1 {
		return nil, nil
	}

	asset := assets[0]
	return map[string]any{
		"cueKey": cue.ID, "name": cue.Name, "assetId": asset.Id,
		"preview": "/api/app/v1/ruleset-assets/" + asset.Id,
	}, nil
}
