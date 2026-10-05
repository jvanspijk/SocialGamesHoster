package chat

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/application/pagination"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func Register(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1")
	group.GET("/games/{id}/rooms", listRooms)
	group.POST("/games/{id}/unread-counts", unreadCounts).Bind(apis.BodyLimit(64 << 10))
	group.POST("/games/{id}/rooms/player-dm", createPlayerDM)
	group.GET("/rooms/{roomId}/messages", listMessages)
	group.POST("/rooms/{roomId}/messages", createMessage)
	group.DELETE("/rooms/{roomId}/messages/{messageId}", deleteMessage)
	group.PATCH("/rooms/{roomId}", updateRoom).BindFunc(actorauth.RequireGameMaster)
	group.POST("/rooms/{roomId}/lock", setRoomLock(true)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/rooms/{roomId}/unlock", setRoomLock(false)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/announcements", createAnnouncement).Bind(apis.BodyLimit(136 << 20)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/announcements/{announcementId}/acknowledge", acknowledgeAnnouncement).BindFunc(actorauth.RequirePlayer)
	group.GET("/games/{id}/announcements/{announcementId}/media/{kind}", announcementMedia)
	group.GET("/games/{id}/announcements", listAnnouncements).BindFunc(actorauth.RequireGameMaster)
}

type access struct {
	Game        *core.Record
	Room        *core.Record
	Participant *core.Record
	Membership  *core.Record
	Policy      rulesets.RoomPermission
	IsGM        bool
	Definition  *rulesets.DefinitionV1
}

func resolveAccess(event *core.RequestEvent, roomID string) (access, error) {
	if event.Auth == nil || !event.Auth.GetBool("active") {
		return access{}, result.AppError{Code: "auth.required", Message: "Sign in to continue.", Status: http.StatusUnauthorized}
	}
	room, err := event.App.FindRecordById("chat_rooms", roomID)
	if err != nil || room.GetString("kind") == "announcements" {
		return access{}, result.AppError{Code: "chat.room_not_found", Message: "Chat room not found.", Status: http.StatusNotFound}
	}
	game, err := event.App.FindRecordById("games", room.GetString("game"))
	if err != nil {
		return access{}, err
	}
	return resolveRoomAccess(event, game, room)
}

// Room listings and counts already load these records. Keep authorization in one
// owner while avoiding a second lookup of each room and its game.
func resolveRoomAccess(event *core.RequestEvent, game, room *core.Record) (access, error) {
	if actorauth.IsGameMaster(event.Auth) {
		resolved := access{Game: game, Room: room, IsGM: true}
		if definition, err := resolved.definition(); err == nil {
			resolved.Policy = gameMasterRoomPolicy(definition, game, room)
		}
		return resolved, nil
	}
	if !actorauth.IsPlayer(event.Auth) {
		return access{}, result.Forbidden("chat.forbidden", "This room is not available.")
	}
	participants, err := event.App.FindRecordsByFilter("participants", "game = {:game} && profile = {:profile}", "", 1, 0,
		dbx.Params{"game": game.Id, "profile": event.Auth.Id})
	if err != nil || len(participants) == 0 {
		return access{}, result.Forbidden("chat.forbidden", "This room is not available.")
	}
	participant := participants[0]
	memberships, err := event.App.FindRecordsByFilter("chat_memberships", "room = {:room} && participant = {:participant}", "", 1, 0,
		dbx.Params{"room": room.Id, "participant": participant.Id})
	if err != nil {
		return access{}, err
	}
	var membership *core.Record
	if len(memberships) > 0 {
		membership = memberships[0]
	}
	definition, err := definitionFromGame(game)
	if err != nil {
		return access{}, err
	}
	return resolvePlayerRoomAccess(game, room, participant, membership, definition)
}

func resolvePlayerRoomAccess(game, room, participant, membership *core.Record, definition rulesets.DefinitionV1) (access, error) {
	base, override := resolveRoomPolicy(definition, game.GetString("phase_key"), room)
	if room.GetString("kind") == "announcements" || room.GetString("kind") == "gm_dm" {
		base = rulesets.RoomPermission{
			Visible: true, Readable: true, Sendable: room.GetString("kind") == "gm_dm",
			SenderDisplay: rulesets.SenderProfileName,
		}
		override = nil
	}
	policy := EffectivePolicy(base, override, participantRoomState(definition, game, room, participant, membership), RoomState{
		ManuallyLocked: !room.GetBool("players_can_post"), ManualVisibilityOverride: room.GetString("manual_visibility_override"),
	})
	if room.GetString("kind") == "custom" && !customChannelSenderAllowed(definition, room, participant) {
		policy.Sendable = false
	}
	if game.GetString("status") == string(gamepolicy.GameReview) ||
		gamepolicy.IsArchived(gamepolicy.GameStatus(game.GetString("status"))) {
		policy.Sendable = false
	}
	if game.GetString("status") == "paused" && room.GetString("kind") != "gm_dm" {
		explicitlyAllowed := override != nil && override.Sendable != nil && *override.Sendable
		if !explicitlyAllowed {
			policy.Sendable = false
		}
	}
	if !policy.Readable && !policy.Visible {
		return access{}, result.Forbidden("chat.forbidden", "This room is not available.")
	}
	return access{Game: game, Room: room, Participant: participant, Membership: membership, Policy: policy, Definition: &definition}, nil
}

// Bulk reads reuse only this request's current ruleset, participant and
// memberships. They evaluate each room through the same policy as history.
func resolveGameRoomAccess(event *core.RequestEvent, game *core.Record, rooms []*core.Record) ([]access, error) {
	resolved := make([]access, 0, len(rooms))
	if len(rooms) == 0 {
		return resolved, nil
	}
	definition, definitionErr := definitionFromGame(game)
	if actorauth.IsGameMaster(event.Auth) {
		for _, room := range rooms {
			item := access{Game: game, Room: room, IsGM: true}
			if definitionErr == nil {
				item.Definition = &definition
				item.Policy = gameMasterRoomPolicy(definition, game, room)
			}
			resolved = append(resolved, item)
		}
		return resolved, nil
	}
	if !actorauth.IsPlayer(event.Auth) {
		return resolved, nil
	}
	if definitionErr != nil {
		return nil, definitionErr
	}
	participants, err := event.App.FindRecordsByFilter("participants", "game = {:game} && profile = {:profile}", "", 1, 0, dbx.Params{"game": game.Id, "profile": event.Auth.Id})
	if err != nil {
		return nil, err
	}
	if len(participants) == 0 {
		return resolved, nil
	}
	participant := participants[0]
	memberships, err := event.App.FindRecordsByFilter("chat_memberships", "participant = {:participant} && room.game = {:game}", "", 0, 0, dbx.Params{"participant": participant.Id, "game": game.Id})
	if err != nil {
		return nil, err
	}
	byRoom := make(map[string]*core.Record, len(memberships))
	for _, membership := range memberships {
		byRoom[membership.GetString("room")] = membership
	}
	for _, room := range rooms {
		item, err := resolvePlayerRoomAccess(game, room, participant, byRoom[room.Id], definition)
		if err != nil {
			continue
		}
		resolved = append(resolved, item)
	}
	return resolved, nil
}

func listRooms(event *core.RequestEvent) error {
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
	}
	if event.Auth == nil || !event.Auth.GetBool("active") {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "Sign in to continue.", Status: http.StatusUnauthorized})
	}
	rooms, err := event.App.FindRecordsByFilter("chat_rooms", "game = {:game} && kind != 'announcements'", "kind,label", 200, 0, dbx.Params{"game": game.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	response := make([]map[string]any, 0, len(rooms))
	accesses, err := resolveGameRoomAccess(event, game, rooms)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	for _, resolved := range accesses {
		response = append(response, projectRoom(event.App, resolved))
	}
	return event.JSON(http.StatusOK, response)
}

type playerDMRequest struct {
	ParticipantID string `json:"participantId"`
}

func createPlayerDM(event *core.RequestEvent) error {
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
	}
	if !actorauth.IsActivePlayer(event.Auth) {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "A player profile is required.", Status: http.StatusUnauthorized})
	}
	if game.GetString("status") != "running" && game.GetString("status") != "paused" {
		return httpx.WriteError(event, result.Conflict("chat.dm_not_allowed", "Player messages are only available during play."))
	}
	definition, err := definitionFromGame(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if definition.Chat.DefaultPolicy.PlayerDM == nil || !definition.Chat.DefaultPolicy.PlayerDM.Visible {
		return httpx.WriteError(event, result.Forbidden("chat.dm_disabled", "Player-to-player messages are disabled for this game."))
	}
	var request playerDMRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("chat.invalid_dm", "Choose another player.", nil))
	}
	selfRecords, err := event.App.FindRecordsByFilter("participants",
		"game = {:game} && profile = {:profile} && status = 'active'", "", 1, 0,
		dbx.Params{"game": game.Id, "profile": event.Auth.Id})
	if err != nil || len(selfRecords) == 0 {
		return httpx.WriteError(event, result.Forbidden("chat.forbidden", "Join this game before creating a room."))
	}
	self := selfRecords[0]
	other, err := event.App.FindRecordById("participants", request.ParticipantID)
	if err != nil || other.GetString("game") != game.Id ||
		!gamepolicy.IsActivePlayer(gamepolicy.ParticipantStatus(other.GetString("status"))) ||
		other.Id == self.Id {
		return httpx.WriteError(event, result.Invalid("chat.invalid_dm", "Choose another active player.", nil))
	}
	ids := []string{self.Id, other.Id}
	sort.Strings(ids)
	key := "dm:" + ids[0] + ":" + ids[1]
	var room *core.Record
	err = event.App.RunInTransaction(func(tx core.App) error {
		room, err = findRoomByKey(tx, game.Id, key)
		if err != nil {
			collection, collectionErr := tx.FindCollectionByNameOrId("chat_rooms")
			if collectionErr != nil {
				return collectionErr
			}
			room = core.NewRecord(collection)
			room.Set("game", game.Id)
			room.Set("room_key", key)
			room.Set("kind", "player_dm")
			room.Set(
				"label",
				"Private · "+self.GetString("display_name_snapshot")+" & "+other.GetString("display_name_snapshot"),
			)
			room.Set("manual_visibility_override", "default")
			room.Set("sender_display", definition.Chat.DefaultPolicy.PlayerDM.SenderDisplay)
			if saveErr := tx.Save(room); saveErr != nil {
				return saveErr
			}
		}
		if err := ensureChatMembership(tx, room.Id, self.Id); err != nil {
			return err
		}
		return ensureChatMembership(tx, room.Id, other.Id)
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	resolved, err := resolveAccess(event, room.Id)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	return event.JSON(http.StatusCreated, projectRoom(event.App, resolved))
}

type messageRequest struct {
	Content string `json:"content"`
}

func createMessage(event *core.RequestEvent) error {
	resolved, err := resolveAccess(event, event.Request.PathValue("roomId"))
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if gamepolicy.IsArchived(gamepolicy.GameStatus(resolved.Game.GetString("status"))) {
		return httpx.WriteError(event, result.Conflict("chat.read_only", "Archived chat is read-only."))
	}
	if resolved.IsGM {
		if !resolved.Policy.GameMasterMaySend && resolved.Room.GetString("kind") != "gm_dm" && resolved.Room.GetString("kind") != "announcements" {
			return httpx.WriteError(event, result.Forbidden("chat.send_forbidden", "Game masters cannot send to this room."))
		}
	} else if !resolved.Policy.Sendable {
		return httpx.WriteError(event, result.Forbidden("chat.send_forbidden", "Sending is disabled in this room."))
	}
	var request messageRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("chat.invalid_message", "The message could not be read.", nil))
	}
	request.Content = strings.TrimSpace(request.Content)
	if len([]rune(request.Content)) < 1 || len([]rune(request.Content)) > 1000 || hasDisallowedControl(request.Content) {
		return httpx.WriteError(event, result.Invalid("chat.invalid_message", "Enter a message of at most 1000 characters.", nil))
	}
	definition, err := resolved.definition()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if roomMessageRestriction(definition, resolved.Room) == rulesets.ChatEmojiOnly && !isEmojiOnly(request.Content) {
		return httpx.WriteError(event, result.Invalid("chat.emoji_only", "This channel accepts emoji only.", nil))
	}
	collection, err := event.App.FindCollectionByNameOrId("chat_messages")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	message := core.NewRecord(collection)
	message.Set("room", resolved.Room.Id)
	message.Set("message_kind", "message")
	message.Set("sender_id", event.Auth.Id)
	if resolved.IsGM {
		message.Set("sender_type", "game_master")
		message.Set("sender_label_snapshot", event.Auth.GetString("display_name"))
	} else {
		label := playerSenderLabel(definition, resolved, event.Auth)
		message.Set("sender_type", "player")
		message.Set("sender_participant", resolved.Participant.Id)
		message.Set("sender_label_snapshot", label)
	}
	message.Set("content", request.Content)
	if err := event.App.Save(message); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projected := projectMessage(message, event.Auth, resolved.IsGM)
	publishRoomMessage(event.App, resolved, "chat.message_created", message)
	return event.JSON(http.StatusCreated, projected)
}

type updateRoomRequest struct {
	PlayersCanPost *bool `json:"playersCanPost"`
}

func updateRoom(event *core.RequestEvent) error {
	room, err := event.App.FindRecordById("chat_rooms", event.Request.PathValue("roomId"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "chat.room_not_found", Message: "Chat room not found.", Status: http.StatusNotFound})
	}
	game, err := event.App.FindRecordById("games", room.GetString("game"))
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if appError := gamepolicyapp.GameMutationError(game); appError != nil {
		return httpx.WriteError(event, *appError)
	}
	var request updateRoomRequest
	if err := event.BindBody(&request); err != nil || request.PlayersCanPost == nil {
		return httpx.WriteError(event, result.Invalid("chat.room_invalid", "Choose whether players can post.", nil))
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		room, err = tx.FindRecordById("chat_rooms", room.Id)
		if err != nil {
			return err
		}
		room.Set("players_can_post", *request.PlayersCanPost)
		room.Set("manually_locked", !*request.PlayersCanPost)
		if err := tx.Save(room); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "chat.players_can_post_changed", "chat_room", room.Id,
			map[string]any{"playersCanPost": *request.PlayersCanPost}, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	resolved := access{Game: game, Room: room, IsGM: true, Policy: policyForGM(game, room)}
	projected := projectRoom(event.App, resolved)
	publishRoom(event.App, resolved, "chat.room_updated", projected)
	return event.JSON(http.StatusOK, projected)
}

func hasDisallowedControl(value string) bool {
	for _, character := range value {
		if character < 0x20 && character != '\n' && character != '\t' {
			return true
		}
		if character == 0x7f {
			return true
		}
	}
	return false
}

func listMessages(event *core.RequestEvent) error {
	resolved, err := resolveAccess(event, event.Request.PathValue("roomId"))
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if !resolved.IsGM && !resolved.Policy.Readable {
		return httpx.WriteError(event, result.Forbidden("chat.read_forbidden", "Reading is disabled in this room."))
	}
	filter := "room = {:room}"
	params := dbx.Params{"room": resolved.Room.Id}
	if joined, left, bounded := messageHistoryWindow(resolved); bounded {
		filter += " && created >= {:joined} && created <= {:left}"
		params["joined"] = joined.Format(types.DefaultDateLayout)
		params["left"] = left.Format(types.DefaultDateLayout)
	}
	if cursor := event.Request.URL.Query().Get("cursor"); cursor != "" {
		position, err := pagination.Decode(cursor)
		if err != nil {
			return httpx.WriteError(event, result.Invalid("chat.invalid_cursor", "The message cursor is invalid.", nil))
		}
		filter += " && " + pagination.DescendingCreatedIDPredicate
		params["created"] = position.Created.UTC().Format(types.DefaultDateLayout)
		params["id"] = position.ID
	}
	records, err := event.App.FindRecordsByFilter("chat_messages", filter, pagination.DescendingCreatedIDSort, pagination.QueryLimit, 0, params)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	records, nextCursor := pagination.Window(records, func(record *core.Record) pagination.Cursor {
		return pagination.Cursor{Created: record.GetDateTime("created").Time().UTC(), ID: record.Id}
	})
	messages := make([]map[string]any, len(records))
	for index, record := range records {
		messages[index] = projectMessage(record, event.Auth, resolved.IsGM)
	}
	return event.JSON(http.StatusOK, map[string]any{"items": messages, "nextCursor": nextCursor})
}

func deleteMessage(event *core.RequestEvent) error {
	resolved, err := resolveAccess(event, event.Request.PathValue("roomId"))
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if appError := gamepolicyapp.GameMutationError(resolved.Game); appError != nil {
		return httpx.WriteError(event, *appError)
	}
	message, err := event.App.FindRecordById("chat_messages", event.Request.PathValue("messageId"))
	if err != nil || message.GetString("room") != resolved.Room.Id {
		return httpx.WriteError(event, result.AppError{Code: "chat.message_not_found", Message: "Message not found.", Status: http.StatusNotFound})
	}
	own := !resolved.IsGM && message.GetString("sender_type") == "player" &&
		message.GetString("sender_participant") == resolved.Participant.Id
	if !resolved.IsGM && !own {
		return httpx.WriteError(event, result.Forbidden("chat.delete_forbidden", "You cannot delete this message."))
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		message, err = tx.FindRecordById("chat_messages", message.Id)
		if err != nil {
			return err
		}
		message.Set("content", "")
		message.Set("deleted_at", time.Now().UTC())
		if resolved.IsGM {
			message.Set("deleted_by", event.Auth.Id)
		}
		if err := tx.Save(message); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, resolved.Game.Id, "chat.message_deleted", "chat_message", message.Id,
			map[string]any{"roomId": resolved.Room.Id}, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	projected := projectMessage(message, event.Auth, resolved.IsGM)
	publishRoomMessage(event.App, resolved, "chat.message_deleted", message)
	return event.JSON(http.StatusOK, projected)
}

func setRoomLock(locked bool) func(*core.RequestEvent) error {
	return func(event *core.RequestEvent) error {
		room, err := event.App.FindRecordById("chat_rooms", event.Request.PathValue("roomId"))
		if err != nil {
			return httpx.WriteError(event, result.AppError{Code: "chat.room_not_found", Message: "Chat room not found.", Status: http.StatusNotFound})
		}
		game, err := event.App.FindRecordById("games", room.GetString("game"))
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		if appError := gamepolicyapp.GameMutationError(game); appError != nil {
			return httpx.WriteError(event, *appError)
		}
		if err := event.App.RunInTransaction(func(tx core.App) error {
			room, err = tx.FindRecordById("chat_rooms", room.Id)
			if err != nil {
				return err
			}
			room.Set("manually_locked", locked)
			room.Set("players_can_post", !locked)
			if err := tx.Save(room); err != nil {
				return err
			}
			return applicationaudit.Record(tx, event.Auth, game.Id, "chat.room_lock_changed", "chat_room", room.Id,
				map[string]any{"locked": locked}, event.Get(httpx.TraceIDKey))
		}); err != nil {
			return httpx.WriteErrorFrom(event, err)
		}
		resolved := access{Game: game, Room: room, IsGM: true}
		publishRoom(event.App, resolved, "chat.room_updated", map[string]any{"id": room.Id, "playersCanPost": !locked})
		return event.JSON(http.StatusOK, map[string]any{"id": room.Id, "playersCanPost": !locked})
	}
}

func policyForGM(game, room *core.Record) rulesets.RoomPermission {
	definition, err := definitionFromGame(game)
	if err != nil {
		return rulesets.RoomPermission{}
	}
	return gameMasterRoomPolicy(definition, game, room)
}

func gameMasterRoomPolicy(definition rulesets.DefinitionV1, game, room *core.Record) rulesets.RoomPermission {
	base, override := resolveRoomPolicy(definition, game.GetString("phase_key"), room)
	if override != nil && override.GameMasterMaySend != nil {
		base.GameMasterMaySend = *override.GameMasterMaySend
	}
	if room.GetString("kind") == "gm_dm" || room.GetString("kind") == "announcements" {
		base.GameMasterMaySend = true
	}
	return base
}

func resolveRoomPolicy(definition rulesets.DefinitionV1, phaseKey string, room *core.Record) (rulesets.RoomPermission, *rulesets.PartialRoomPermission) {
	var base rulesets.RoomPermission
	var override *rulesets.PartialRoomPermission
	phase := definition.Chat.PhaseOverrides[phaseKey]
	switch room.GetString("kind") {
	case "general":
		if definition.Chat.DefaultPolicy.General != nil {
			base = *definition.Chat.DefaultPolicy.General
		}
		override = phase.General
	case "player_dm":
		if definition.Chat.DefaultPolicy.PlayerDM != nil {
			base = *definition.Chat.DefaultPolicy.PlayerDM
		}
		override = phase.PlayerDM
	case "team":
		base = definition.Chat.DefaultPolicy.Teams[room.GetString("team_key")]
		if value, ok := phase.Teams[room.GetString("team_key")]; ok {
			override = &value
		}
	case "custom":
		channel := rulesets.FindChatChannel(definition, rulesets.ChatChannelIDFromRoomKey(room.GetString("room_key")))
		if channel != nil {
			base, override = rulesets.ChatChannelPolicy(*channel, phaseKey)
		}
	}
	return base, override
}

func customChannelSenderAllowed(definition rulesets.DefinitionV1, room, participant *core.Record) bool {
	channel := rulesets.FindChatChannel(definition, rulesets.ChatChannelIDFromRoomKey(room.GetString("room_key")))
	if channel == nil {
		return false
	}
	for _, role := range definition.Roles {
		if role.ID == participant.GetString("role_key") {
			return rulesets.ChatChannelAudienceMatches(*channel, role, true)
		}
	}
	return false
}

func roomMessageRestriction(definition rulesets.DefinitionV1, room *core.Record) rulesets.ChatMessageRestriction {
	if room.GetString("kind") != "custom" {
		return rulesets.ChatNormalText
	}
	channel := rulesets.FindChatChannel(definition, rulesets.ChatChannelIDFromRoomKey(room.GetString("room_key")))
	if channel == nil {
		return rulesets.ChatNormalText
	}
	return channel.MessageRestriction
}

func definitionFromGame(game *core.Record) (rulesets.DefinitionV1, error) {
	return rulesets.DecodeSnapshot(game.Get("ruleset_snapshot"))
}

// definition reuses the frozen snapshot only within this request or publication.
func (resolved *access) definition() (rulesets.DefinitionV1, error) {
	if resolved.Definition != nil {
		return *resolved.Definition, nil
	}
	definition, err := definitionFromGame(resolved.Game)
	if err != nil {
		return rulesets.DefinitionV1{}, err
	}
	resolved.Definition = &definition
	return definition, nil
}

func playerSenderLabel(definition rulesets.DefinitionV1, resolved access, profile *core.Record) string {
	roleName := ""
	teamName := ""
	for _, role := range definition.Roles {
		if role.ID != resolved.Participant.GetString("role_key") {
			continue
		}
		roleName = role.Name
		for _, team := range definition.Teams {
			if team.ID == role.TeamID {
				teamName = team.Name
			}
		}
	}
	display := resolved.Policy.SenderDisplay
	if display == "" {
		display = rulesets.SenderProfileName
	}
	return SenderLabel(Sender{
		ProfileName: profile.GetString("display_name"), GameAlias: resolved.Participant.GetString("game_alias"),
		PlayerNumber: resolved.Participant.GetInt("player_number"), RoleLabel: roleName, TeamLabel: teamName,
	}, display)
}

func projectRoom(app core.App, resolved access) map[string]any {
	sendable := resolved.Policy.Sendable
	if resolved.IsGM {
		sendable = resolved.Policy.GameMasterMaySend ||
			resolved.Room.GetString("kind") == "gm_dm" ||
			resolved.Room.GetString("kind") == "announcements"
	}
	return map[string]any{
		"id": resolved.Room.Id, "key": resolved.Room.GetString("room_key"), "kind": resolved.Room.GetString("kind"),
		"label": resolved.Room.GetString("label"), "playersCanPost": resolved.Room.GetBool("players_can_post"),
		"readable": resolved.IsGM || resolved.Policy.Readable, "sendable": sendable,
		"gameMasterMaySend":  resolved.Policy.GameMasterMaySend,
		"messageRestriction": roomMessageRestrictionFromGame(resolved.Game, resolved.Room),
		"latestMessage":      latestMessageSummary(app, resolved.Room.Id),
	}
}

func roomMessageRestrictionFromGame(game, room *core.Record) rulesets.ChatMessageRestriction {
	definition, err := definitionFromGame(game)
	if err != nil {
		return rulesets.ChatNormalText
	}
	return roomMessageRestriction(definition, room)
}

func latestMessageSummary(app core.App, roomID string) any {
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
	message := messages[0]
	preview := message.GetString("content")
	if !message.GetDateTime("deleted_at").IsZero() {
		preview = "Message removed"
	}
	runes := []rune(strings.TrimSpace(preview))
	if len(runes) > 120 {
		preview = string(runes[:120]) + "…"
	}
	return map[string]any{
		"createdAt": message.GetDateTime("created").Time().UTC(), "id": message.Id,
		"senderLabel": message.GetString("sender_label_snapshot"), "preview": preview,
	}
}

func projectMessage(message, viewer *core.Record, isGM bool) map[string]any {
	content := message.GetString("content")
	deleted := !message.GetDateTime("deleted_at").IsZero()
	if deleted {
		content = ""
	}
	projected := map[string]any{
		"id": message.Id, "roomId": message.GetString("room"), "kind": message.GetString("message_kind"),
		"senderType": message.GetString("sender_type"), "senderLabel": message.GetString("sender_label_snapshot"),
		"content": content, "cueKey": message.GetString("cue_key"), "deleted": deleted,
		"createdAt": message.GetDateTime("created").Time().UTC(),
	}
	if isGM {
		projected["senderParticipantId"] = message.GetString("sender_participant")
	}
	if viewer != nil && message.GetString("sender_id") == viewer.Id {
		projected["isOwn"] = true
	}
	return projected
}

func publishRoom(app core.App, resolved access, kind string, payload any) {
	_ = realtime.Publish(app, "room:"+resolved.Room.Id, realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: resolved.Game.Id, Revision: resolved.Game.GetInt("revision"),
		Kind: kind, Payload: payload,
	}, roomEventAuthorization(app, resolved, false))
}

func publishRoomMessage(app core.App, resolved access, kind string, message *core.Record) {
	_ = realtime.PublishProjected(app, "room:"+resolved.Room.Id, realtime.Event[map[string]any]{
		EventID: realtime.NewEventID(), GameID: resolved.Game.Id, Revision: resolved.Game.GetInt("revision"),
		Kind: kind,
	}, roomEventAuthorization(app, resolved, true),
		func(auth *core.Record) map[string]any {
			return projectMessage(message, auth, auth != nil && actorauth.IsGameMaster(auth))
		})
}

// roomEventAuthorization takes a fresh, publication-local access snapshot. Recipient
// checks perform no database reads, and no permissions survive into the next event.
func roomEventAuthorization(app core.App, resolved access, requireReadable bool) func(*core.Record) bool {
	deny := func(*core.Record) bool { return false }
	room, err := app.FindRecordById("chat_rooms", resolved.Room.Id)
	if err != nil || room.GetString("game") != resolved.Game.Id {
		return deny
	}
	game, err := app.FindRecordById("games", resolved.Game.Id)
	if err != nil {
		return deny
	}
	definition, definitionErr := definitionFromGame(game)
	base, override := resolveRoomPolicy(definition, game.GetString("phase_key"), room)
	if room.GetString("kind") == "gm_dm" {
		base = rulesets.RoomPermission{
			Visible: true, Readable: true, Sendable: true, SenderDisplay: rulesets.SenderProfileName,
		}
		override = nil
	}
	roomState := RoomState{
		ManuallyLocked:           !room.GetBool("players_can_post"),
		ManualVisibilityOverride: room.GetString("manual_visibility_override"),
	}
	allowed := make(map[string]bool)
	if definitionErr == nil {
		memberships, membershipErr := app.FindRecordsByFilter("chat_memberships",
			gamepolicyapp.RoomReadableMembershipFilter+" && participant.game = {:game}",
			"", 0, 0, dbx.Params{"room": room.Id, "game": game.Id})
		participants, participantErr := app.FindRecordsByFilter("participants", "game = {:game}",
			"", 0, 0, dbx.Params{"game": game.Id})
		// Fail closed for players if either bulk read fails; never use partial data.
		if membershipErr == nil && participantErr == nil {
			byParticipant := make(map[string]*core.Record, len(memberships))
			for _, membership := range memberships {
				byParticipant[membership.GetString("participant")] = membership
			}
			for _, participant := range participants {
				policy := EffectivePolicy(base, override,
					participantRoomState(definition, game, room, participant, byParticipant[participant.Id]), roomState)
				if policy.Readable || (!requireReadable && policy.Visible) {
					allowed[participant.GetString("profile")] = true
				}
			}
		}
	}
	return func(auth *core.Record) bool {
		if auth == nil || !auth.GetBool("active") {
			return false
		}
		if actorauth.IsGameMaster(auth) {
			return true
		}
		return actorauth.IsPlayer(auth) && allowed[auth.Id]
	}
}

// Live role-controlled audiences are derived from current assignments, so role
// changes need neither membership synchronization nor cache invalidation.
// Archived rooms retain their explicit historical membership grants.
func participantRoomState(definition rulesets.DefinitionV1, game, room, participant, membership *core.Record) ParticipantState {
	status := gamepolicy.ParticipantStatus(participant.GetString("status"))
	state := ParticipantState{IsActive: gamepolicy.IsActivePlayer(status)}
	if membership != nil {
		state.IsMember = membership.GetDateTime("left_at").IsZero() && gamepolicy.IsCurrentMember(status)
		state.HistoricalRead = membership.GetBool("historical_access")
	}
	if rulesets.RoleControlsChatRoom(room.GetString("kind")) && !gamepolicy.IsArchived(gamepolicy.GameStatus(game.GetString("status"))) {
		state.IsMember = gamepolicy.IsCurrentMember(status) && rulesets.ChatRoomReaderMatches(definition,
			room.GetString("kind"), room.GetString("room_key"), room.GetString("team_key"), participant.GetString("role_key"))
		state.HistoricalRead = false
	}
	return state
}

func findRoomByKey(app core.App, gameID, key string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter("chat_rooms", "game = {:game} && room_key = {:key}", "", 1, 0,
		dbx.Params{"game": gameID, "key": key})
	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("room not found")
	}
	return records[0], nil
}

func ensureChatMembership(app core.App, roomID, participantID string) error {
	records, err := app.FindRecordsByFilter("chat_memberships", "room = {:room} && participant = {:participant}", "", 1, 0,
		dbx.Params{"room": roomID, "participant": participantID})
	if err != nil {
		return err
	}
	if len(records) > 0 {
		records[0].Set("left_at", nil)
		return app.Save(records[0])
	}
	collection, err := app.FindCollectionByNameOrId("chat_memberships")
	if err != nil {
		return err
	}
	record := core.NewRecord(collection)
	record.Set("room", roomID)
	record.Set("participant", participantID)
	record.Set("joined_at", time.Now().UTC())
	return app.Save(record)
}
