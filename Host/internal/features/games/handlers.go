package games

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/abilities"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type createGameRequest struct {
	Name             string `json:"name"`
	RulesetVersionID string `json:"rulesetVersionId"`
}

func listGames(event *core.RequestEvent) error {
	records, err := event.App.FindRecordsByFilter("games", "", "-created", 200, 0)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	response := make([]map[string]any, len(records))
	for index, record := range records {
		response[index] = projectGame(record)
	}
	return event.JSON(http.StatusOK, response)
}

func createGame(event *core.RequestEvent) error {
	var request createGameRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game.invalid", "The game could not be read.", nil))
	}
	request.Name = strings.TrimSpace(request.Name)
	if len([]rune(request.Name)) < 1 || len([]rune(request.Name)) > 120 {
		return httpx.WriteError(event, result.Invalid("game.invalid_name", "Enter a game name between 1 and 120 characters.", nil))
	}
	version, err := event.App.FindRecordById("ruleset_versions", request.RulesetVersionID)
	if err != nil || version.GetString("state") != "published" {
		return httpx.WriteError(event, result.Invalid("game.invalid_ruleset", "Choose a ready ruleset.", nil))
	}
	logical, err := event.App.FindRecordById("rulesets", version.GetString("ruleset"))
	if err != nil || logical.GetString("latest_published_version") != version.Id {
		return httpx.WriteError(event, result.Invalid("game.invalid_ruleset", "Choose the current ready version of a ruleset.", nil))
	}
	collection, err := event.App.FindCollectionByNameOrId("games")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	record := core.NewRecord(collection)
	record.Set("name", request.Name)
	record.Set("status", StatusDraft)
	record.Set("ruleset_version", version.Id)
	record.Set("ruleset_snapshot", version.Get("definition"))
	record.Set("joining_open", false)
	record.Set("revision", 0)
	record.Set("round_number", 0)
	record.Set("timer_state", "inactive")
	record.Set("roles_visible", false)
	record.Set("role_visibility_revision", 0)
	record.Set("created_by", event.Auth.Id)
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = audit(event.App, event.Auth, record.Id, "game.created", "game", record.Id, nil, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusCreated, projectGame(record))
}

func duplicateGame(event *core.RequestEvent) error {
	source, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	collection, err := event.App.FindCollectionByNameOrId("games")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	record := core.NewRecord(collection)
	record.Set("name", source.GetString("name")+" (copy)")
	record.Set("status", StatusDraft)
	record.Set("ruleset_version", source.GetString("ruleset_version"))
	record.Set("ruleset_snapshot", source.Get("ruleset_snapshot"))
	record.Set("joining_open", false)
	record.Set("revision", 0)
	record.Set("round_number", 0)
	record.Set("timer_state", "inactive")
	record.Set("roles_visible", false)
	record.Set("role_visibility_revision", 0)
	record.Set("created_by", event.Auth.Id)
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = audit(event.App, event.Auth, record.Id, "game.duplicated", "game", source.Id, nil, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusCreated, projectGame(record))
}

func deleteGame(event *core.RequestEvent) error {
	record, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	status := record.GetString("status")
	if status != string(StatusDraft) && status != string(StatusReview) && status != string(StatusArchived) {
		return httpx.WriteError(event, result.Conflict("game.delete_not_allowed", "Only draft, review, or archived games can be deleted."))
	}
	var request struct {
		Confirmation string `json:"confirmation"`
	}
	if err := event.BindBody(&request); err != nil || request.Confirmation != "DELETE "+record.Id {
		return httpx.WriteError(event, result.Invalid(
			"game.confirmation_required",
			`Type "DELETE `+record.Id+`" to confirm permanent deletion.`,
			nil,
		))
	}
	err = event.App.RunInTransaction(func(tx core.App) error {
		if err := clearGameSession(tx, record.Id, true); err != nil {
			return err
		}
		current, err := tx.FindRecordById("games", record.Id)
		if err != nil {
			return err
		}
		return tx.Delete(current)
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.NoContent(http.StatusNoContent)
}

// closeJoining stops new entries for an active game. Closing a lobby instead
// resets it to a draft, removing its temporary roster and chat session so it
// can be reopened cleanly.
func closeJoining(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	if game.GetString("status") == string(StatusRunning) || game.GetString("status") == string(StatusPaused) {
		game.Set("joining_open", false)
		if err := event.App.Save(game); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		_ = audit(event.App, event.Auth, game.Id, "game.joining_closed", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
		publishGame(event.App, game, "game.joining_closed", projectGame(game))
		return event.JSON(http.StatusOK, projectGame(game))
	}
	next, err := ApplyTransition(stateFromRecord(game), CancelLobby, time.Now().UTC())
	if err != nil {
		return httpx.WriteError(event, result.Conflict("game.transition_not_allowed", err.Error()))
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		current, err := tx.FindRecordById("games", game.Id)
		if err != nil {
			return err
		}
		if err := clearGameSession(tx, current.Id, false); err != nil {
			return err
		}
		applyState(current, next)
		current.Set("join_code", "")
		current.Set("joining_open", false)
		current.Set("round_number", 0)
		current.Set("phase_key", "")
		current.Set("phase_started_at", nil)
		current.Set("ability_phase_locked_at", nil)
		current.Set("ability_phase_instance", 0)
		current.Set("timer_state", "inactive")
		current.Set("timer_total_ms", 0)
		current.Set("timer_remaining_ms", 0)
		current.Set("timer_ends_at", nil)
		current.Set("timer_revision", current.GetInt("timer_revision")+1)
		current.Set("roles_visible", false)
		current.Set("role_visibility_revision", current.GetInt("role_visibility_revision")+1)
		current.Set("completion_previous_status", "")
		return tx.Save(current)
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	game, err = event.App.FindRecordById("games", game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = audit(event.App, event.Auth, game.Id, "game.lobby_cancelled", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "game.lobby_cancelled", projectGame(game))
	return event.JSON(http.StatusOK, projectGame(game))
}

func clearGameSession(app core.App, gameID string, includeAudit bool) error {
	abilityChoices, err := app.FindRecordsByFilter("ability_choices", "game = {:game}", "", 10000, 0, dbx.Params{"game": gameID})
	if err != nil {
		return err
	}
	for _, choice := range abilityChoices {
		if err := app.Delete(choice); err != nil {
			return err
		}
	}
	items, err := app.FindRecordsByFilter("attention_items", "game = {:game}", "", 10000, 0, dbx.Params{"game": gameID})
	if err != nil {
		return err
	}
	for _, item := range items {
		receipts, err := app.FindRecordsByFilter("attention_receipts", "attention_item = {:item}", "", 10000, 0, dbx.Params{"item": item.Id})
		if err != nil {
			return err
		}
		for _, receipt := range receipts {
			if err := app.Delete(receipt); err != nil {
				return err
			}
		}
		if err := app.Delete(item); err != nil {
			return err
		}
	}
	rooms, err := app.FindRecordsByFilter("chat_rooms", "game = {:game}", "", 500, 0, dbx.Params{"game": gameID})
	if err != nil {
		return err
	}
	for _, room := range rooms {
		for _, collection := range []string{"chat_messages", "chat_memberships"} {
			dependents, err := app.FindRecordsByFilter(collection, "room = {:room}", "", 10000, 0, dbx.Params{"room": room.Id})
			if err != nil {
				return err
			}
			for _, dependent := range dependents {
				if err := app.Delete(dependent); err != nil {
					return err
				}
			}
		}
		if err := app.Delete(room); err != nil {
			return err
		}
	}
	collections := []string{"achievement_awards", "participants"}
	if includeAudit {
		collections = append(collections, "game_audit")
	}
	for _, collection := range collections {
		dependents, err := app.FindRecordsByFilter(collection, "game = {:game}", "", 10000, 0, dbx.Params{"game": gameID})
		if err != nil {
			return err
		}
		for _, dependent := range dependents {
			if err := app.Delete(dependent); err != nil {
				return err
			}
		}
	}
	return nil
}

func openLobby(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	now := time.Now().UTC()
	next, err := ApplyTransition(stateFromRecord(game), OpenLobby, now)
	if err != nil {
		return httpx.WriteError(event, result.Conflict("game.transition_not_allowed", err.Error()))
	}
	err = event.App.RunInTransaction(func(tx core.App) error {
		game, err = tx.FindRecordById("games", game.Id)
		if err != nil {
			return err
		}
		applyState(game, next)
		game.Set("join_code", "")
		game.Set("joining_open", true)
		game.Set("roles_visible", false)
		if err := tx.Save(game); err != nil {
			return err
		}
		definition, err := snapshot(game)
		if err != nil {
			return err
		}
		if definition.Chat.DefaultPolicy.General != nil {
			_, err = ensureRoom(tx, game.Id, "general", "general", "General", "")
		}
		return err
	})
	if err != nil {
		event.App.Logger().Error("failed to open game lobby", "gameId", game.Id, "error", err)
		return httpx.WriteError(event, result.Conflict("game.live_game_exists", "Another lobby or game is already live."))
	}
	_ = audit(event.App, event.Auth, game.Id, "game.lobby_opened", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "game.lobby_opened", projectGame(game))
	publishLobbyOpened(event.App, game)
	return event.JSON(http.StatusOK, projectGame(game))
}

// openJoining reopens entry for a game already in progress. A lobby always
// opens with joining enabled, so this only applies after play has started.
func openJoining(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	if !canOpenJoining(Status(game.GetString("status"))) {
		return httpx.WriteError(event, result.Conflict(
			"game.joining_not_available",
			"Joining can only be reopened while a game is running or paused.",
		))
	}
	if game.GetBool("joining_open") {
		return event.JSON(http.StatusOK, projectGame(game))
	}
	game.Set("joining_open", true)
	if err := event.App.Save(game); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = audit(event.App, event.Auth, game.Id, "game.joining_opened", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "game.joining_opened", projectGame(game))
	publishLobbyOpened(event.App, game)
	return event.JSON(http.StatusOK, projectGame(game))
}

func joinGame(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	var participant *core.Record
	err = event.App.RunInTransaction(func(tx core.App) error {
		game, err = tx.FindRecordById("games", game.Id)
		if err != nil {
			return err
		}
		status := Status(game.GetString("status"))
		if !IsLive(status) || !game.GetBool("joining_open") {
			return result.AppError{Code: "game.joining_closed", Message: "This game is not accepting players.", Status: http.StatusConflict}
		}
		existing, err := tx.FindRecordsByFilter("participants", "game = {:game} && profile = {:profile}", "", 1, 0,
			dbx.Params{"game": game.Id, "profile": event.Auth.Id})
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			participant = existing[0]
			if participant.GetString("status") == "kicked" {
				return result.AppError{Code: "game.player_kicked", Message: "A game master removed this profile from the game.", Status: http.StatusForbidden}
			}
			return nil
		}
		participants, err := gameParticipants(tx, game.Id)
		if err != nil {
			return err
		}
		definition, err := snapshot(game)
		if err != nil {
			return err
		}
		if len(participants) >= definition.Metadata.MaxPlayers || len(participants) >= 30 {
			return result.AppError{Code: "game.full", Message: "This lobby is full.", Status: http.StatusConflict}
		}
		collection, err := tx.FindCollectionByNameOrId("participants")
		if err != nil {
			return err
		}
		participant = core.NewRecord(collection)
		participant.Set("game", game.Id)
		participant.Set("profile", event.Auth.Id)
		participant.Set("display_name_snapshot", event.Auth.GetString("display_name"))
		participant.Set("seat_number", nextSeat(participants))
		participant.Set("status", "active")
		participant.Set("outcome", "unset")
		participant.Set("joined_at", time.Now().UTC())
		if err := tx.Save(participant); err != nil {
			return err
		}
		dm, err := ensureRoom(
			tx,
			game.Id,
			"gm:"+participant.Id,
			"gm_dm",
			"Game master · "+participant.GetString("display_name_snapshot"),
			"",
		)
		if err != nil {
			return err
		}
		if err := ensureMembership(tx, dm.Id, participant.Id); err != nil {
			return err
		}
		if general, err := findRoom(tx, game.Id, "general"); err == nil {
			if err := ensureMembership(tx, general.Id, participant.Id); err != nil {
				return err
			}
		}
		incrementRevision(game)
		return tx.Save(game)
	})
	if err != nil {
		return writeGameError(event, err)
	}
	_ = audit(event.App, event.Auth, game.Id, "participant.joined", "participant", participant.Id, nil, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "participant.joined", projectParticipant(participant, false))
	publishGameMasters(event.App, game, "participant.joined_private", projectParticipant(participant, true))
	return event.JSON(http.StatusOK, projectParticipant(participant, false))
}

func adminView(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	participants, err := gameParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projected := make([]map[string]any, len(participants))
	for index, participant := range participants {
		projected[index] = projectParticipant(participant, true)
	}
	rooms, _ := event.App.FindRecordsByFilter("chat_rooms", "game = {:game} && kind != 'announcements'", "kind,label", 200, 0, dbx.Params{"game": game.Id})
	attentionItems, _ := event.App.FindRecordsByFilter("attention_items", "game = {:game}", "-created,-id", 50, 0, dbx.Params{"game": game.Id})
	attentionSummaries := make([]map[string]any, 0, len(attentionItems))
	for _, item := range attentionItems {
		if summary, summaryErr := projectAdminAttentionSummary(event.App, item); summaryErr == nil {
			attentionSummaries = append(attentionSummaries, summary)
		}
	}
	assetRecords, _ := event.App.FindRecordsByFilter(
		"ruleset_assets", "ruleset_version = {:version}", "asset_key", 100, 0,
		dbx.Params{"version": game.GetString("ruleset_version")},
	)
	projectedAssets := make([]map[string]any, 0, len(assetRecords))
	for _, asset := range assetRecords {
		projectedAssets = append(projectedAssets, map[string]any{
			"id": asset.Id, "assetKey": asset.GetString("asset_key"), "kind": asset.GetString("kind"),
		})
	}
	awards, _ := event.App.FindRecordsByFilter("achievement_awards", "game = {:game}", "-created", 500, 0, dbx.Params{"game": game.Id})
	projectedAwards := make([]map[string]any, len(awards))
	for index, award := range awards {
		projectedAwards[index] = map[string]any{
			"id": award.Id, "profileId": award.GetString("profile"),
			"achievementId": award.GetString("achievement_key"), "title": award.GetString("title_snapshot"),
			"description": award.GetString("description_snapshot"), "points": award.GetInt("points_snapshot"),
			"hiddenUntilGameCompleted": award.GetBool("hidden_until_game_completed"),
			"awardedAt":                dateValue(award, "created"),
		}
	}
	auditRecords, _ := event.App.FindRecordsByFilter("game_audit", "game = {:game}", "-created", 25, 0, dbx.Params{"game": game.Id})
	projectedAudit := make([]map[string]any, len(auditRecords))
	for index, entry := range auditRecords {
		projectedAudit[index] = map[string]any{
			"id": entry.Id, "actorLabel": entry.GetString("actor_label"), "action": entry.GetString("action"),
			"targetType": entry.GetString("target_type"), "detail": entry.Get("detail"),
			"createdAt": dateValue(entry, "created"),
		}
	}
	abilityProgress, abilityResults, abilityErr := abilities.ProjectAdmin(event.App, game, definition, participants)
	if abilityErr != nil {
		return httpx.WriteError(event, result.Internal(abilityErr))
	}
	return event.JSON(http.StatusOK, map[string]any{
		"game":            projectGame(game),
		"timer":           projectTimer(game),
		"ruleset":         game.Get("ruleset_snapshot"),
		"participants":    projected,
		"rooms":           projectRooms(event.App, rooms),
		"attentionItems":  attentionSummaries,
		"assets":          projectedAssets,
		"awards":          projectedAwards,
		"audit":           projectedAudit,
		"abilityProgress": abilityProgress,
		"abilityResults":  abilityResults,
	})
}

func projectTimer(game *core.Record) map[string]any {
	projected := map[string]any{
		"status":      game.GetString("timer_state"),
		"totalMs":     game.GetInt("timer_total_ms"),
		"remainingMs": game.GetInt("timer_remaining_ms"),
		"revision":    game.GetInt("timer_revision"),
		"serverTime":  time.Now().UTC(),
	}
	if endsAt := dateValue(game, "timer_ends_at"); endsAt != nil {
		projected["endsAt"] = endsAt
	}
	return projected
}

func playerView(event *core.RequestEvent) error {
	game, err := findLiveGame(event.App)
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.no_live_game", Message: "There is no live game.", Status: http.StatusNotFound})
	}
	records, err := event.App.FindRecordsByFilter("participants",
		"game = {:game} && profile = {:profile} && status != 'kicked' && status != 'left'", "", 1, 0,
		dbx.Params{"game": game.Id, "profile": event.Auth.Id})
	if err != nil || len(records) == 0 {
		return httpx.WriteError(event, result.AppError{Code: "game.not_joined", Message: "Join the live game to view it.", Status: http.StatusForbidden})
	}
	participant := records[0]
	definition, err := snapshot(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	roleAvailable := game.GetBool("roles_visible") && participant.GetString("role_key") != ""
	var role any
	knowledge := []map[string]any{}
	if roleAvailable {
		role = roleByID(definition, participant.GetString("role_key"))
		knowledge, err = projectKnowledge(event.App, game.Id, participant, definition)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	rooms, err := visibleRoomsForPlayer(event.App, game, participant, definition)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	participants, err := gameParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	party := make([]map[string]any, 0, len(participants))
	for _, member := range participants {
		if member.GetString("status") == "kicked" || member.GetString("status") == "left" {
			continue
		}
		party = append(party, map[string]any{
			"id": member.Id, "profileId": member.GetString("profile"),
			"displayName": member.GetString("display_name_snapshot"), "gameAlias": member.GetString("game_alias"),
			"seatNumber": member.GetInt("seat_number"), "status": member.GetString("status"),
		})
	}
	attentionItems, err := unacknowledgedAttentionForParticipant(event.App, game.Id, participant.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	abilityChoices, err := abilities.ProjectPlayer(event.App, game, participant, definition)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	assetRecords, _ := event.App.FindRecordsByFilter(
		"ruleset_assets", "ruleset_version = {:version}", "asset_key", 100, 0,
		dbx.Params{"version": game.GetString("ruleset_version")},
	)
	privateKeys := map[string]bool{}
	if !roleAvailable {
		for _, key := range privateRoleAssetKeys(definition) {
			privateKeys[key] = true
		}
	}
	assets := make([]map[string]any, 0, len(assetRecords))
	for _, asset := range assetRecords {
		if privateKeys[asset.GetString("asset_key")] {
			continue
		}
		assets = append(assets, map[string]any{
			"id": asset.Id, "assetKey": asset.GetString("asset_key"), "kind": asset.GetString("kind"),
			"checksum": asset.GetString("checksum"), "preview": "/api/app/v1/ruleset-assets/" + asset.Id,
		})
	}
	return event.JSON(http.StatusOK, map[string]any{
		"game": projectPlayerGame(game),
		"participant": map[string]any{
			"id": participant.Id, "displayName": participant.GetString("display_name_snapshot"),
			"gameAlias": participant.GetString("game_alias"), "seatNumber": participant.GetInt("seat_number"),
			"status": participant.GetString("status"),
		},
		"ruleset":        map[string]any{"name": definition.Metadata.Name, "description": definition.Metadata.Description},
		"roleAvailable":  roleAvailable,
		"roleRevision":   participant.GetInt("role_revision"),
		"role":           role,
		"knowledge":      knowledge,
		"rooms":          rooms,
		"party":          party,
		"attentionItems": attentionItems,
		"assets":         assets,
		"abilityChoices": abilityChoices,
	})
}

func uniqueJoinCode(app core.App) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for range 20 {
		var bytes [6]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			return "", err
		}
		for index := range bytes {
			bytes[index] = alphabet[int(bytes[index])%len(alphabet)]
		}
		code := string(bytes[:])
		records, err := app.FindRecordsByFilter("games", "join_code = {:code}", "", 1, 0, dbx.Params{"code": code})
		if err != nil {
			return "", err
		}
		if len(records) == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not generate a unique join code")
}

func nextSeat(participants []*core.Record) int {
	maximum := 0
	for _, participant := range participants {
		if participant.GetInt("seat_number") > maximum {
			maximum = participant.GetInt("seat_number")
		}
	}
	return maximum + 1
}

func ensureRoom(app core.App, gameID, key, kind, label, teamKey string) (*core.Record, error) {
	if room, err := findRoom(app, gameID, key); err == nil {
		return room, nil
	}
	collection, err := app.FindCollectionByNameOrId("chat_rooms")
	if err != nil {
		return nil, err
	}
	room := core.NewRecord(collection)
	room.Set("game", gameID)
	room.Set("room_key", key)
	room.Set("kind", kind)
	room.Set("label", label)
	room.Set("team_key", teamKey)
	room.Set("players_can_post", true)
	room.Set("manual_visibility_override", "default")
	room.Set("sender_display", "profile_name")
	if err := app.Save(room); err != nil {
		return nil, err
	}
	return room, nil
}

func findRoom(app core.App, gameID, key string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter("chat_rooms", "game = {:game} && room_key = {:key}", "", 1, 0,
		dbx.Params{"game": gameID, "key": key})
	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("room not found")
	}
	return records[0], nil
}

func ensureMembership(app core.App, roomID, participantID string) error {
	existing, err := app.FindRecordsByFilter("chat_memberships", "room = {:room} && participant = {:participant}", "", 1, 0,
		dbx.Params{"room": roomID, "participant": participantID})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		existing[0].Set("left_at", nil)
		return app.Save(existing[0])
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

func roleByID(definition rulesets.DefinitionV1, id string) any {
	if id == "" {
		return nil
	}
	for _, role := range definition.Roles {
		if role.ID == id {
			abilities := make([]rulesets.Ability, 0)
			for _, abilityID := range role.AbilityIDs {
				for _, ability := range definition.Abilities {
					if ability.ID == abilityID {
						abilities = append(abilities, ability)
					}
				}
			}
			var team *rulesets.Team
			for index := range definition.Teams {
				if definition.Teams[index].ID == role.TeamID {
					copy := definition.Teams[index]
					team = &copy
				}
			}
			return map[string]any{"id": role.ID, "name": role.Name, "description": role.Description,
				"winCondition": role.WinCondition, "imageAssetKey": role.ImageAssetKey, "team": team, "abilities": abilities}
		}
	}
	return nil
}

func projectPlayerGame(game *core.Record) map[string]any {
	return map[string]any{
		"id": game.Id, "name": game.GetString("name"), "status": game.GetString("status"),
		"revision": game.GetInt("revision"), "roundNumber": game.GetInt("round_number"),
		"phaseKey": game.GetString("phase_key"), "phaseStartedAt": dateValue(game, "phase_started_at"),
		"abilityPhaseLockedAt": dateValue(game, "ability_phase_locked_at"),
	}
}

func privateRoleAssetKeys(definition rulesets.DefinitionV1) []string {
	seen := map[string]bool{}
	keys := make([]string, 0)
	add := func(key string) {
		if key != "" && !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	for _, role := range definition.Roles {
		add(role.ImageAssetKey)
	}
	for _, team := range definition.Teams {
		add(team.ImageAssetKey)
	}
	for _, ability := range definition.Abilities {
		add(ability.ImageAssetKey)
	}
	return keys
}

func projectKnowledge(app core.App, gameID string, viewer *core.Record, definition rulesets.DefinitionV1) ([]map[string]any, error) {
	if viewer.GetString("role_key") == "" {
		return []map[string]any{}, nil
	}
	var viewerRole rulesets.Role
	for _, role := range definition.Roles {
		if role.ID == viewer.GetString("role_key") {
			viewerRole = role
		}
	}
	participants, err := activeParticipants(app, gameID)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0)
	for _, target := range participants {
		if target.Id == viewer.Id || target.GetString("role_key") == "" {
			continue
		}
		var targetRole rulesets.Role
		for _, role := range definition.Roles {
			if role.ID == target.GetString("role_key") {
				targetRole = role
			}
		}
		revealed := map[string]any{"participantId": target.Id, "seatNumber": target.GetInt("seat_number")}
		hasReveal := false
		for _, rule := range definition.KnowledgeRules {
			if !selectorMatches(viewerRole, rule.Viewer) || !selectorMatches(targetRole, rule.Target) {
				continue
			}
			for _, field := range rule.Reveal {
				switch field {
				case "identity":
					revealed["displayName"] = target.GetString("display_name_snapshot")
				case "role":
					revealed["role"] = map[string]any{"id": targetRole.ID, "name": targetRole.Name}
				case "team":
					revealed["teamId"] = targetRole.TeamID
				case "elimination_state":
					revealed["status"] = target.GetString("status")
				}
				hasReveal = true
			}
		}
		if hasReveal {
			result = append(result, revealed)
		}
	}
	return result, nil
}

func selectorMatches(role rulesets.Role, selector rulesets.Selector) bool {
	matchList := func(values []string, value string) bool {
		if len(values) == 0 {
			return true
		}
		for _, candidate := range values {
			if candidate == value {
				return true
			}
		}
		return false
	}
	if !matchList(selector.RoleIDs, role.ID) || !matchList(selector.TeamIDs, role.TeamID) {
		return false
	}
	for _, category := range selector.CategoryIDs {
		found := false
		for _, roleCategory := range role.CategoryIDs {
			found = found || category == roleCategory
		}
		if found {
			return true
		}
	}
	if len(selector.CategoryIDs) > 0 {
		return false
	}
	for _, tag := range selector.Tags {
		found := false
		for _, roleTag := range role.Tags {
			found = found || tag == roleTag
		}
		if found {
			return true
		}
	}
	return len(selector.Tags) == 0
}

func projectRooms(app core.App, rooms []*core.Record) []map[string]any {
	result := make([]map[string]any, 0, len(rooms))
	for _, room := range rooms {
		if room.GetString("kind") == "announcements" {
			continue
		}
		projected := map[string]any{
			"id": room.Id, "key": room.GetString("room_key"), "kind": room.GetString("kind"),
			"label": room.GetString("label"), "teamKey": room.GetString("team_key"),
			"playersCanPost": room.GetBool("players_can_post"),
		}
		projected["latestMessage"] = latestMessageSummary(app, room.Id)
		result = append(result, projected)
	}
	return result
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
		"id": message.Id, "createdAt": dateValue(message, "created"),
		"senderLabel": message.GetString("sender_label_snapshot"), "preview": preview,
	}
}

func decodeSnapshot(record *core.Record, target any) error {
	data, err := json.Marshal(record.Get("ruleset_snapshot"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
