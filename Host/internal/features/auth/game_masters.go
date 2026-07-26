package authfeature

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/profiles"
	platformaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/audit"
	platformauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/auth"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type createGameMasterRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

type updateGameMasterRequest struct {
	DisplayName *string `json:"displayName"`
	Active      *bool   `json:"active"`
	MakeOwner   bool    `json:"makeOwner"`
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func RegisterOwnerRoutes(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1/owner/game-masters")
	group.BindFunc(platformauth.RequireOwner)
	group.GET("", listGameMasters)
	group.POST("", createGameMaster)
	group.PATCH("/{id}", updateGameMaster)
	group.POST("/{id}/reset-password", resetGameMasterPassword)
	group.DELETE("/{id}", deleteGameMaster)
}

func listGameMasters(event *core.RequestEvent) error {
	records, err := event.App.FindRecordsByFilter(platformauth.GameMastersCollection, "", "username", 100, 0)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	response := make([]map[string]any, len(records))
	for i, record := range records {
		response[i] = projectGameMaster(record)
	}
	return event.JSON(http.StatusOK, response)
}

func createGameMaster(event *core.RequestEvent) error {
	var request createGameMasterRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game_master.invalid", "The account details could not be read.", nil))
	}
	username := strings.ToLower(strings.TrimSpace(request.Username))
	displayName, _, nameErr := profiles.NormalizeName(request.DisplayName)
	fields := result.FieldErrors{}
	if len(username) < 3 || len(username) > 32 {
		fields["username"] = []string{"Use between 3 and 32 characters."}
	}
	if nameErr != nil {
		fields["displayName"] = []string{nameErr.Error()}
	}
	if len(request.Password) < 10 {
		fields["password"] = []string{"Use at least 10 characters."}
	}
	if len(fields) > 0 {
		return httpx.WriteError(event, result.Invalid("game_master.invalid", "Correct the highlighted account details.", fields))
	}
	collection, err := event.App.FindCollectionByNameOrId(platformauth.GameMastersCollection)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	record := core.NewRecord(collection)
	record.Set("username", username)
	record.Set("display_name", displayName)
	record.Set("active", true)
	record.Set("is_owner", false)
	record.SetPassword(request.Password)
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Invalid("game_master.save_failed", "The account could not be created.", result.FieldErrors{
			"username": {"Choose a different username."},
		}))
	}
	_ = platformaudit.Record(event.App, event.Auth, "", "game_master.created", "game_master", record.Id,
		map[string]any{"username": username}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusCreated, projectGameMaster(record))
}

func updateGameMaster(event *core.RequestEvent) error {
	target, err := event.App.FindRecordById(platformauth.GameMastersCollection, event.Request.PathValue("id"))
	if err != nil {
		return gameMasterNotFound(event)
	}
	var request updateGameMasterRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game_master.invalid", "The account changes could not be read.", nil))
	}
	if request.DisplayName != nil {
		displayName, _, err := profiles.NormalizeName(*request.DisplayName)
		if err != nil {
			return httpx.WriteError(event, result.Invalid("game_master.invalid_name", "Enter a valid display name.", nil))
		}
		target.Set("display_name", displayName)
	}
	if request.Active != nil {
		if target.GetBool("is_owner") && !*request.Active {
			return httpx.WriteError(event, result.Conflict("game_master.owner_required", "Transfer ownership before disabling the owner account."))
		}
		target.Set("active", *request.Active)
		if !*request.Active {
			target.RefreshTokenKey()
		}
	}
	if request.MakeOwner && !target.GetBool("is_owner") {
		if err := transferOwnership(event.App, event.Auth.Id, target.Id); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		updated, err := event.App.FindRecordById(platformauth.GameMastersCollection, target.Id)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		_ = platformaudit.Record(event.App, event.Auth, "", "game_master.ownership_transferred", "game_master", target.Id,
			nil, event.Get(httpx.TraceIDKey))
		return event.JSON(http.StatusOK, projectGameMaster(updated))
	}
	if err := event.App.Save(target); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = platformaudit.Record(event.App, event.Auth, "", "game_master.updated", "game_master", target.Id,
		map[string]any{"active": target.GetBool("active")}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusOK, projectGameMaster(target))
}

func transferOwnership(app core.App, currentOwnerID, nextOwnerID string) error {
	return app.RunInTransaction(func(txApp core.App) error {
		current, err := txApp.FindRecordById(platformauth.GameMastersCollection, currentOwnerID)
		if err != nil {
			return err
		}
		next, err := txApp.FindRecordById(platformauth.GameMastersCollection, nextOwnerID)
		if err != nil {
			return err
		}
		next.Set("is_owner", true)
		next.Set("active", true)
		next.RefreshTokenKey()
		if err := txApp.Save(next); err != nil {
			return err
		}
		current.Set("is_owner", false)
		current.RefreshTokenKey()
		return txApp.Save(current)
	})
}

func resetGameMasterPassword(event *core.RequestEvent) error {
	target, err := event.App.FindRecordById(platformauth.GameMastersCollection, event.Request.PathValue("id"))
	if err != nil {
		return gameMasterNotFound(event)
	}
	var request resetPasswordRequest
	if err := event.BindBody(&request); err != nil || len(request.Password) < 10 {
		return httpx.WriteError(event, result.Invalid("game_master.invalid_password", "Use a password of at least 10 characters.", nil))
	}
	target.SetPassword(request.Password)
	if err := event.App.Save(target); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = platformaudit.Record(event.App, event.Auth, "", "game_master.password_reset", "game_master", target.Id,
		nil, event.Get(httpx.TraceIDKey))
	return event.NoContent(http.StatusNoContent)
}

func deleteGameMaster(event *core.RequestEvent) error {
	target, err := event.App.FindRecordById(platformauth.GameMastersCollection, event.Request.PathValue("id"))
	if err != nil {
		return gameMasterNotFound(event)
	}
	if target.Id == event.Auth.Id {
		return httpx.WriteError(event, result.Conflict("game_master.cannot_delete_self", "You cannot delete the account you are currently using."))
	}
	if target.GetBool("is_owner") {
		return httpx.WriteError(event, result.Conflict("game_master.owner_required", "Transfer ownership before deleting the owner account."))
	}
	if err := event.App.Delete(target); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = platformaudit.Record(event.App, event.Auth, "", "game_master.deleted", "game_master", target.Id,
		map[string]any{"username": target.GetString("username")}, event.Get(httpx.TraceIDKey))
	return event.NoContent(http.StatusNoContent)
}

func projectGameMaster(record *core.Record) map[string]any {
	return map[string]any{
		"id":          record.Id,
		"username":    record.GetString("username"),
		"displayName": record.GetString("display_name"),
		"isOwner":     record.GetBool("is_owner"),
		"active":      record.GetBool("active"),
		"lastLoginAt": record.GetDateTime("last_login_at").Time().UTC(),
	}
}

func gameMasterNotFound(event *core.RequestEvent) error {
	return httpx.WriteError(event, result.AppError{Code: "game_master.not_found", Message: "Game-master account not found.", Status: http.StatusNotFound})
}
