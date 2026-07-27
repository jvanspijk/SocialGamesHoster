package setup

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/jvanspijk/SocialGamesHoster/Host/fixtures"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/owner"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/profiles"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const collectionName = "game_masters"

type OwnerRequest struct {
	Username               string `json:"username"`
	DisplayName            string `json:"displayName"`
	Password               string `json:"password"`
	TrustedLANAcknowledged bool   `json:"trustedLanAcknowledged"`
}

func Register(router *core.ServeEvent) {
	group := router.Router.Group("/api/app/v1/setup")
	group.GET("/status", status)
	group.GET("/join-qr", joinQRCode)
	group.POST("/owner", createOwner)
}

func status(event *core.RequestEvent) error {
	count, err := event.App.CountRecords(collectionName)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, map[string]any{
		"needsOwner": count == 0,
		"joinUrl":    owner.JoinURL(event.App),
	})
}

func joinQRCode(event *core.RequestEvent) error {
	image, err := qrcode.Encode(owner.JoinURL(event.App), qrcode.Medium, 320)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	event.Response.Header().Set("Cache-Control", "no-store")
	return event.Blob(http.StatusOK, "image/png", image)
}

func createOwner(event *core.RequestEvent) error {
	if !isLoopback(event.RemoteIP()) {
		return httpx.WriteError(event, result.Forbidden("setup.loopback_only", "First-time setup is available only on the host computer."))
	}
	count, err := event.App.CountRecords(collectionName)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if count != 0 {
		return httpx.WriteError(event, result.Conflict("setup.complete", "The owner account has already been created."))
	}

	var request OwnerRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("request.invalid", "The submitted setup details are invalid.", nil))
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
	if len(request.Password) < 6 {
		fields["password"] = []string{"Use at least 6 characters."}
	}
	if !request.TrustedLANAcknowledged {
		fields["trustedLanAcknowledged"] = []string{"Confirm that this host will be used only on a network you trust."}
	}
	if len(fields) > 0 {
		return httpx.WriteError(event, result.Invalid("setup.invalid", "Please correct the highlighted setup details.", fields))
	}

	collection, err := event.App.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	record := core.NewRecord(collection)
	record.Set("username", username)
	record.Set("display_name", displayName)
	record.Set("is_owner", true)
	record.Set("active", true)
	record.Set("last_login_at", time.Now().UTC())
	record.SetPassword(request.Password)
	if err := event.App.Save(record); err != nil {
		event.App.Logger().Error("failed to create first owner", "error", err)
		return httpx.WriteError(event, result.Invalid("setup.save_failed", "The owner account could not be created.", result.FieldErrors{
			"username": {"Choose a different username and try again."},
		}))
	}
	settings, err := owner.EnsureSettings(event.App)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	settings.Set("trusted_lan_acknowledged", true)
	if err := event.App.Save(settings); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := fixtures.Seed(event.App, record.Id); err != nil {
		event.App.Logger().Error("failed to seed demonstration rulesets", "error", err)
		return httpx.WriteError(event, result.Internal(err))
	}
	token, err := record.NewAuthToken()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, map[string]any{
		"token": token,
		"gameMaster": map[string]any{
			"id":          record.Id,
			"username":    username,
			"displayName": displayName,
			"isOwner":     true,
		},
	})
}

func isLoopback(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil && ip.IsLoopback()
}
