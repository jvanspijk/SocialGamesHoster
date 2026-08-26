package games

import (
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/abilities"
	chatfeature "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/chat"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type createGameRequest struct {
	Name      string `json:"name"`
	RulesetID string `json:"rulesetId"`
}

func listGames(event *core.RequestEvent) error {
	records, err := event.App.FindRecordsByFilter("games", "", "-created", 200, 0)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	participants, err := event.App.FindRecordsByFilter(
		"participants",
		gamepolicyapp.CurrentParticipantStatusFilter,
		"",
		6000,
		0,
	)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	playerCounts := make(map[string]int, len(records))
	for _, participant := range participants {
		playerCounts[participant.GetString("game")]++
	}
	response := make([]map[string]any, len(records))
	for index, record := range records {
		definition, err := snapshot(record)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		response[index] = projectGame(record)
		response[index]["playerCount"] = playerCounts[record.Id]
		response[index]["maxPlayers"] = definition.Metadata.MaxPlayers
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
	logical, err := event.App.FindRecordById("rulesets", request.RulesetID)
	if err != nil || logical.GetBool("archived") {
		return httpx.WriteError(event, result.Invalid("game.invalid_ruleset", "Choose a valid ruleset.", nil))
	}
	versionID := logical.GetString("latest_published_version")
	version, err := event.App.FindRecordById("ruleset_versions", versionID)
	if err != nil || versionID == "" {
		return httpx.WriteError(event, result.Invalid("game.invalid_ruleset", "Choose a valid ruleset.", nil))
	}
	collection, err := event.App.FindCollectionByNameOrId("games")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	record := core.NewRecord(collection)
	record.Set("name", request.Name)
	record.Set("status", StatusLobby)
	record.Set("ruleset_version", version.Id)
	record.Set("ruleset_snapshot", version.Get("definition"))
	record.Set("joining_open", true)
	record.Set("revision", 1)
	record.Set("round_number", 0)
	record.Set("timer_state", "inactive")
	record.Set("roles_visible", false)
	record.Set("role_visibility_revision", 0)
	record.Set("created_by", event.Auth.Id)
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Save(record); err != nil {
			return err
		}
		definition, err := snapshot(record)
		if err != nil {
			return err
		}
		if err := chatfeature.EnsureLobbyRoom(tx, record.Id, definition); err != nil {
			return err
		}
		if err := applicationaudit.Record(tx, event.Auth, record.Id, "game.created", "game", record.Id, nil, event.Get(httpx.TraceIDKey)); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, record.Id, "game.lobby_opened", "game", record.Id, nil, event.Get(httpx.TraceIDKey))
	}); err != nil {
		event.App.Logger().Error("failed to create game with open lobby", "gameId", record.Id, "error", err)
		return httpx.WriteError(event, result.Conflict("game.live_game_exists", "Another lobby or game is already live."))
	}
	publishGame(event.App, record, "game.lobby_opened", projectGame(record))
	publishLobbyOpened(event.App, record)
	return event.JSON(http.StatusCreated, projectGame(record))
}

func duplicateGame(event *core.RequestEvent) error {
	source, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
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
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Save(record); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, record.Id, "game.duplicated", "game", source.Id, nil, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, projectGame(record))
}

func deleteGame(event *core.RequestEvent) error {
	record, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	// Deletion is an explicit exception to archived-game immutability: it
	// removes the entire aggregate after an exact confirmation.
	if !canDeleteGame(Status(record.GetString("status"))) {
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

func canDeleteGame(status Status) bool {
	return status == StatusDraft || status == StatusReview || status == StatusArchived
}

// cancelGame permanently removes a lobby that was created by mistake. Started
// games use the completion flow so their roster and history remain available.
func cancelGame(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if !canCancelGame(Status(game.GetString("status"))) {
		return httpx.WriteError(event, result.Conflict("game.cancel_not_allowed", "Only a game that has not started can be cancelled."))
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := clearGameSession(tx, game.Id, true); err != nil {
			return err
		}
		current, err := tx.FindRecordById("games", game.Id)
		if err != nil {
			return err
		}
		return tx.Delete(current)
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.NoContent(http.StatusNoContent)
}

func canCancelGame(status Status) bool {
	return status == StatusLobby
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
	if err := chatfeature.ClearGameSession(app, gameID); err != nil {
		return err
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
		return httpx.WriteErrorFrom(event, err)
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
		if err := chatfeature.EnsureLobbyRoom(tx, game.Id, definition); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "game.lobby_opened", "game", game.Id, nil, event.Get(httpx.TraceIDKey))
	})
	if err != nil {
		event.App.Logger().Error("failed to open game lobby", "gameId", game.Id, "error", err)
		return httpx.WriteError(event, result.Conflict("game.live_game_exists", "Another lobby or game is already live."))
	}
	publishGame(event.App, game, "game.lobby_opened", projectGame(game))
	publishLobbyOpened(event.App, game)
	return event.JSON(http.StatusOK, projectGame(game))
}

func joinGame(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
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
			if gamepolicy.ParticipantStatus(participant.GetString("status")) == gamepolicy.ParticipantKicked {
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
		participant.Set("status", gamepolicy.ParticipantActive)
		participant.Set("outcome", "unset")
		participant.Set("joined_at", time.Now().UTC())
		if err := tx.Save(participant); err != nil {
			return err
		}
		if err := chatfeature.AddParticipant(tx, game.Id, participant); err != nil {
			return err
		}
		incrementRevision(game)
		if err := tx.Save(game); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "participant.joined", "participant", participant.Id, nil, event.Get(httpx.TraceIDKey))
	})
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	publishGame(event.App, game, "participant.joined", projectParticipant(participant, false))
	publishGameMasters(event.App, game, "participant.joined_private", projectParticipant(participant, true))
	return event.JSON(http.StatusOK, projectParticipant(participant, false))
}

func adminView(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
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
	rooms, attentionSummaries := chatfeature.AdminViewData(event.App, game.Id)
	assetRecords, _ := event.App.FindRecordsByFilter(
		"ruleset_assets", "ruleset_version = {:version} && storage_state = 'ready'", "asset_key", 100, 0,
		dbx.Params{"version": game.GetString("ruleset_version")},
	)
	projectedAssets := make([]map[string]any, 0, len(assetRecords))
	for _, asset := range assetRecords {
		projectedAssets = append(projectedAssets, map[string]any{
			"id": asset.Id, "assetKey": asset.GetString("asset_key"), "kind": asset.GetString("kind"),
			"displayName": asset.GetString("display_name"), "accessibilityText": asset.GetString("accessibility_text"),
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
		"rooms":           rooms,
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
	participant, err := gamepolicyapp.CurrentParticipantByGameAndProfile(event.App, game.Id, event.Auth.Id)
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_joined", Message: "Join the live game to view it.", Status: http.StatusForbidden})
	}
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
	rooms, err := chatfeature.VisibleRoomsForPlayer(event.App, game, participant, definition)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	participants, err := gameParticipants(event.App, game.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	party := make([]map[string]any, 0, len(participants))
	for _, member := range participants {
		if !gamepolicy.IsCurrentMember(gamepolicy.ParticipantStatus(member.GetString("status"))) {
			continue
		}
		party = append(party, map[string]any{
			"id": member.Id, "profileId": member.GetString("profile"),
			"displayName": member.GetString("display_name_snapshot"), "gameAlias": member.GetString("game_alias"),
			"seatNumber": member.GetInt("seat_number"), "status": member.GetString("status"),
		})
	}
	attentionItems, err := chatfeature.UnacknowledgedAttentionForParticipant(event.App, game.Id, participant.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	abilityChoices, err := abilities.ProjectPlayer(event.App, game, participant, definition)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	assetRecords, _ := event.App.FindRecordsByFilter(
		"ruleset_assets", "ruleset_version = {:version} && storage_state = 'ready'", "asset_key", 100, 0,
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
			"displayName": asset.GetString("display_name"), "accessibilityText": asset.GetString("accessibility_text"),
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

func nextSeat(participants []*core.Record) int {
	maximum := 0
	for _, participant := range participants {
		if participant.GetInt("seat_number") > maximum {
			maximum = participant.GetInt("seat_number")
		}
	}
	return maximum + 1
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
	participants, err := currentParticipants(app, gameID)
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
