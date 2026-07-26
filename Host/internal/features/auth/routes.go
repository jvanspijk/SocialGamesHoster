package authfeature

import (
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"

	platformauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/auth"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1/auth")
	group.POST("/game-master/login", login)
	group.POST("/refresh", refresh)
	group.POST("/logout", func(event *core.RequestEvent) error {
		return event.NoContent(http.StatusNoContent)
	})
	RegisterOwnerRoutes(event)
}

func login(event *core.RequestEvent) error {
	var request loginRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("auth.invalid", "Username or password is incorrect.", nil))
	}
	record, err := event.App.FindFirstRecordByData(platformauth.GameMastersCollection, "username", strings.ToLower(strings.TrimSpace(request.Username)))
	if err != nil || !record.ValidatePassword(request.Password) || !record.GetBool("active") {
		return httpx.WriteError(event, result.AppError{
			Code:    "auth.invalid",
			Message: "Username or password is incorrect.",
			Status:  http.StatusUnauthorized,
		})
	}
	record.Set("last_login_at", time.Now().UTC())
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return writeAuth(event, record)
}

func refresh(event *core.RequestEvent) error {
	if event.Auth == nil || !event.Auth.GetBool("active") {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "Your session is no longer valid.", Status: http.StatusUnauthorized})
	}
	return writeAuth(event, event.Auth)
}

func writeAuth(event *core.RequestEvent, record *core.Record) error {
	token, err := record.NewAuthToken()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, map[string]any{
		"token": token,
		"actor": map[string]any{
			"id":          record.Id,
			"type":        record.Collection().Name,
			"displayName": record.GetString("display_name"),
			"isOwner":     record.GetBool("is_owner"),
		},
	})
}
