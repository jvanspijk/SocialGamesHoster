package achievements

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func Register(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1")
	group.POST("/games/{id}/achievement-awards", award).BindFunc(actorauth.RequireGameMaster)
	group.DELETE("/games/{id}/achievement-awards/{awardId}", revoke).BindFunc(actorauth.RequireGameMaster)
}

type awardRequest struct {
	ParticipantID string `json:"participantId"`
	AchievementID string `json:"achievementId"`
	Note          string `json:"note"`
}

func award(event *core.RequestEvent) error {
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
	}
	status := gamepolicy.GameStatus(game.GetString("status"))
	if status == gamepolicy.GameDraft || status == gamepolicy.GameLobby || gamepolicy.IsArchived(status) {
		return httpx.WriteError(event, result.Conflict("achievement.not_allowed", "Achievements can only be awarded during play or review."))
	}
	var request awardRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("achievement.invalid", "The achievement award could not be read.", nil))
	}
	request.Note = strings.TrimSpace(request.Note)
	if len([]rune(request.Note)) > 1000 {
		return httpx.WriteError(event, result.Invalid("achievement.invalid_note", "The private note may contain at most 1000 characters.", nil))
	}
	participant, err := event.App.FindRecordById("participants", request.ParticipantID)
	if err != nil || participant.GetString("game") != game.Id {
		return httpx.WriteError(event, result.AppError{Code: "participant.not_found", Message: "Participant not found.", Status: http.StatusNotFound})
	}
	definition, err := definitionFromGame(game)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	var achievement *rulesets.Achievement
	for index := range definition.Achievements {
		if definition.Achievements[index].ID == request.AchievementID {
			achievement = &definition.Achievements[index]
			break
		}
	}
	if achievement == nil {
		return httpx.WriteError(event, result.Invalid("achievement.unknown", "That achievement is not part of the game's ruleset snapshot.", nil))
	}
	collection, err := event.App.FindCollectionByNameOrId("achievement_awards")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	record := core.NewRecord(collection)
	record.Set("profile", participant.GetString("profile"))
	record.Set("game", game.Id)
	record.Set("ruleset_version", game.GetString("ruleset_version"))
	record.Set("achievement_key", achievement.ID)
	record.Set("title_snapshot", achievement.Name)
	record.Set("description_snapshot", achievement.Description)
	record.Set("asset_key", achievement.ImageAssetKey)
	record.Set("points_snapshot", achievement.Points)
	record.Set("hidden_until_game_completed", achievement.HiddenUntilGameCompleted)
	record.Set("awarded_by", event.Auth.Id)
	record.Set("note", request.Note)
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Save(record); err != nil {
			return result.Conflict("achievement.duplicate", "This achievement has already been awarded to that player in this game.")
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "achievement.awarded", "achievement_award", record.Id,
			map[string]any{"achievementKey": achievement.ID, "participantId": participant.Id}, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	projected := projectAward(record, true)
	if awardVisibleDuringStatus(record.GetBool("hidden_until_game_completed"), game.GetString("status")) {
		publish(event.App, game, participant, "achievement.awarded", projected)
	}
	return event.JSON(http.StatusCreated, projected)
}

func revoke(event *core.RequestEvent) error {
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
	}
	if appError := gamepolicyapp.GameMutationError(game); appError != nil {
		return httpx.WriteError(event, *appError)
	}
	record, err := event.App.FindRecordById("achievement_awards", event.Request.PathValue("awardId"))
	if err != nil || record.GetString("game") != game.Id {
		return httpx.WriteError(event, result.AppError{Code: "achievement.not_found", Message: "Achievement award not found.", Status: http.StatusNotFound})
	}
	profileID := record.GetString("profile")
	participants, _ := event.App.FindRecordsByFilter("participants", "game = {:game} && profile = {:profile}", "", 1, 0,
		map[string]any{"game": game.Id, "profile": profileID})
	var participant *core.Record
	if len(participants) > 0 {
		participant = participants[0]
	}
	awardID := record.Id
	visible := awardVisibleDuringStatus(record.GetBool("hidden_until_game_completed"), game.GetString("status"))
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Delete(record); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "achievement.revoked", "achievement_award", awardID, nil, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	payload := map[string]any{"id": awardID}
	if visible {
		publish(event.App, game, participant, "achievement.revoked", payload)
	}
	return event.NoContent(http.StatusNoContent)
}

func definitionFromGame(game *core.Record) (rulesets.DefinitionV1, error) {
	data, err := json.Marshal(game.Get("ruleset_snapshot"))
	if err != nil {
		return rulesets.DefinitionV1{}, err
	}
	var definition rulesets.DefinitionV1
	err = json.Unmarshal(data, &definition)
	return definition, err
}

func projectAward(record *core.Record, includeNote bool) map[string]any {
	projected := map[string]any{
		"id": record.Id, "profileId": record.GetString("profile"), "gameId": record.GetString("game"),
		"achievementId": record.GetString("achievement_key"), "title": record.GetString("title_snapshot"),
		"description": record.GetString("description_snapshot"), "assetKey": record.GetString("asset_key"),
		"points":                   record.GetInt("points_snapshot"),
		"hiddenUntilGameCompleted": record.GetBool("hidden_until_game_completed"),
		"awardedAt":                record.GetDateTime("created").Time().UTC(),
	}
	if includeNote {
		projected["note"] = record.GetString("note")
	}
	return projected
}

func awardVisibleDuringStatus(hiddenUntilCompleted bool, gameStatus string) bool {
	return !hiddenUntilCompleted ||
		gameStatus == "review" ||
		gameStatus == "archived"
}

func publish(app core.App, game, participant *core.Record, kind string, payload any) {
	profileID := ""
	participantID := ""
	if participant != nil {
		profileID = participant.GetString("profile")
		participantID = participant.Id
	}
	authorize := func(auth *core.Record) bool {
		return auth != nil && auth.GetBool("active") &&
			(actorauth.IsGameMaster(auth) ||
				(actorauth.IsPlayer(auth) && auth.Id == profileID))
	}
	if participantID != "" {
		_ = realtime.Publish(app, "participant:"+participantID+":private", realtime.Event[any]{
			EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
			Kind: kind, Payload: payload,
		}, authorize)
	}
	if profileID != "" {
		_ = realtime.Publish(app, "profile:"+profileID, realtime.Event[any]{
			EventID: realtime.NewEventID(), GameID: game.Id, Kind: kind, Payload: payload,
		}, authorize)
	}
}
