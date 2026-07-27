package games

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	chatfeature "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/chat"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func visibleRoomsForPlayer(app core.App, game, participant *core.Record, definition rulesets.DefinitionV1) ([]map[string]any, error) {
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
		base, override := roomPolicy(definition, game.GetString("phase_key"), room)
		if room.GetString("kind") == "announcements" || room.GetString("kind") == "gm_dm" {
			base = rulesets.RoomPermission{Visible: true, Readable: true, Sendable: room.GetString("kind") == "gm_dm", SenderDisplay: rulesets.SenderProfileName}
			override = nil
		}
		state := chatfeature.ParticipantState{
			IsMember:       membership.GetDateTime("left_at").IsZero(),
			IsActive:       participant.GetString("status") == "active",
			HistoricalRead: membership.GetBool("historical_access"),
		}
		policy := chatfeature.EffectivePolicy(base, override, state, chatfeature.RoomState{
			ManuallyLocked: room.GetBool("manually_locked"), ManualVisibilityOverride: room.GetString("manual_visibility_override"),
		})
		if game.GetString("status") == string(StatusReview) || game.GetString("status") == string(StatusArchived) {
			policy.Sendable = false
		}
		if game.GetString("status") == string(StatusPaused) && room.GetString("kind") != "gm_dm" {
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
			"latestMessage": latestMessageCursor(app, room.Id),
		})
	}
	return result, nil
}

func roomPolicy(definition rulesets.DefinitionV1, phaseKey string, room *core.Record) (rulesets.RoomPermission, *rulesets.PartialRoomPermission) {
	var base rulesets.RoomPermission
	var override *rulesets.PartialRoomPermission
	switch room.GetString("kind") {
	case "general":
		if definition.Chat.DefaultPolicy.General != nil {
			base = *definition.Chat.DefaultPolicy.General
		}
		if phase, ok := definition.Chat.PhaseOverrides[phaseKey]; ok {
			override = phase.General
		}
	case "player_dm":
		if definition.Chat.DefaultPolicy.PlayerDM != nil {
			base = *definition.Chat.DefaultPolicy.PlayerDM
		}
		if phase, ok := definition.Chat.PhaseOverrides[phaseKey]; ok {
			override = phase.PlayerDM
		}
	case "team":
		base = definition.Chat.DefaultPolicy.Teams[room.GetString("team_key")]
		if phase, ok := definition.Chat.PhaseOverrides[phaseKey]; ok {
			value, exists := phase.Teams[room.GetString("team_key")]
			if exists {
				override = &value
			}
		}
	}
	return base, override
}

type announcementRequest struct {
	Content  string `json:"content"`
	CueKey   string `json:"cueKey"`
	Audience string `json:"audience"`
	TargetID string `json:"targetId"`
}

func createAnnouncement(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	if game.GetString("status") != string(StatusLobby) && game.GetString("status") != string(StatusRunning) && game.GetString("status") != string(StatusPaused) {
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
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
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
		return writeGameError(event, err)
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
	_ = audit(event.App, event.Auth, game.Id, "attention.announcement_sent", "attention_item", item.Id,
		map[string]any{"cueKey": request.CueKey, "audience": request.Audience, "targetId": request.TargetID, "recipientTotal": len(recipients)}, event.Get(httpx.TraceIDKey))
	summary, err := projectAdminAttentionSummary(event.App, item)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, summary)
}

func acknowledgeAnnouncement(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	item, err := event.App.FindRecordById("attention_items", event.Request.PathValue("announcementId"))
	if err != nil || item.GetString("game") != game.Id || item.GetString("kind") != "announcement" {
		return httpx.WriteError(event, result.AppError{Code: "attention.not_found", Message: "Announcement not found.", Status: http.StatusNotFound})
	}
	participants, err := event.App.FindRecordsByFilter(
		"participants",
		"game = {:game} && profile = {:profile} && status != 'kicked' && status != 'left'",
		"",
		1,
		0,
		dbx.Params{"game": game.Id, "profile": event.Auth.Id},
	)
	if err != nil || len(participants) != 1 {
		return httpx.WriteError(event, result.Forbidden("attention.forbidden", "This announcement is not available."))
	}
	participant := participants[0]
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
		publishGameMasters(event.App, game, "attention.announcement_acknowledged", summary)
	}
	return event.NoContent(http.StatusNoContent)
}

func resolveAnnouncementRecipients(app core.App, game *core.Record, definition rulesets.DefinitionV1, audience, targetID string) ([]*core.Record, error) {
	participants, err := activeParticipants(app, game.Id)
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

func unacknowledgedAttentionForParticipant(app core.App, gameID, participantID string) ([]map[string]any, error) {
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
	return map[string]any{
		"id": item.Id, "kind": "announcement",
		"senderLabel": item.GetString("sender_label_snapshot"),
		"content":     item.GetString("content"), "cueKey": item.GetString("cue_key"),
		"createdAt": dateValue(item, "created"),
	}
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
	return map[string]any{
		"id": item.Id, "kind": "announcement",
		"senderLabel": item.GetString("sender_label_snapshot"),
		"content":     item.GetString("content"), "audience": item.GetString("audience"),
		"targetId": item.GetString("target_id"), "cueKey": item.GetString("cue_key"),
		"createdAt": dateValue(item, "created"), "recipientTotal": len(receipts),
		"acknowledgementCount": acknowledged,
	}, nil
}

func actorHasAttentionReceipt(app core.App, itemID string, auth *core.Record) bool {
	if auth == nil || !auth.GetBool("active") || auth.Collection().Name != "player_profiles" {
		return false
	}
	receipts, err := app.FindRecordsByFilter(
		"attention_receipts",
		"attention_item = {:item} && participant.profile = {:profile} && participant.status != 'kicked' && participant.status != 'left'",
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
	return map[string]any{"createdAt": dateValue(messages[0], "created"), "id": messages[0].Id}
}

func publishAttentionCue(app core.App, game *core.Record, cue rulesets.AudioCue, itemID string) {
	assets, err := app.FindRecordsByFilter(
		"ruleset_assets",
		"ruleset_version = {:version} && asset_key = {:key} && kind = 'audio'",
		"",
		1,
		0,
		dbx.Params{"version": game.GetString("ruleset_version"), "key": cue.AssetKey},
	)
	if err != nil || len(assets) != 1 {
		return
	}
	asset := assets[0]
	payload := map[string]any{
		"cueKey": cue.ID, "name": cue.Name, "assetId": asset.Id,
		"preview": "/api/app/v1/ruleset-assets/" + asset.Id,
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
			participant.GetString("status") == "kicked" || participant.GetString("status") == "left" {
			value := result.Invalid("audio.invalid_target", "Choose a player in this game.", nil)
			return &value
		}
	default:
		value := result.Invalid("audio.invalid_audience", "Choose who should hear the sound cue.", nil)
		return &value
	}
	return nil
}

func publishAudioCue(app core.App, game *core.Record, definition rulesets.DefinitionV1, cue rulesets.AudioCue, audience, targetID string) {
	assets, err := app.FindRecordsByFilter(
		"ruleset_assets",
		"ruleset_version = {:version} && asset_key = {:key} && kind = 'audio'",
		"",
		1,
		0,
		dbx.Params{"version": game.GetString("ruleset_version"), "key": cue.AssetKey},
	)
	if err != nil || len(assets) != 1 {
		return
	}
	asset := assets[0]
	payload := map[string]any{
		"cueKey": cue.ID, "name": cue.Name, "assetId": asset.Id,
		"preview": "/api/app/v1/ruleset-assets/" + asset.Id,
	}
	_ = realtime.Publish(app, "game:"+game.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: "audio.cue", Payload: payload,
	}, func(auth *core.Record) bool {
		if auth == nil || !auth.GetBool("active") {
			return false
		}
		if audience == "game_masters" {
			return auth.Collection().Name == "game_masters"
		}
		if auth.Collection().Name != "player_profiles" {
			return false
		}
		participants, err := app.FindRecordsByFilter(
			"participants",
			"game = {:game} && profile = {:profile} && status != 'kicked' && status != 'left'",
			"",
			1,
			0,
			dbx.Params{"game": game.Id, "profile": auth.Id},
		)
		if err != nil || len(participants) != 1 {
			return false
		}
		participant := participants[0]
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
