package setup

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/jvanspijk/SocialGamesHoster/Host/fixtures"
	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/owner"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/profiles"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const collectionName = actorauth.GameMastersCollection

type OwnerRequest struct {
	Username               string `json:"username"`
	DisplayName            string `json:"displayName"`
	Password               string `json:"password"`
	TrustedLANAcknowledged bool   `json:"trustedLanAcknowledged"`
}

type ownerRecoveryRequest struct {
	Username               string `json:"username"`
	DisplayName            string `json:"displayName"`
	Password               string `json:"password"`
	TrustedLANAcknowledged bool   `json:"trustedLanAcknowledged"`
	Confirmation           string `json:"confirmation"`
}

func Register(router *core.ServeEvent, applicationVersion string) {
	group := router.Router.Group("/api/app/v1/setup")
	group.GET("/status", status(applicationVersion))
	group.GET("/join-qr", joinQRCode)
	group.POST("/owner", createOwner)
	group.POST("/owner-recovery", recoverOwner)
}

func status(applicationVersion string) func(*core.RequestEvent) error {
	return func(event *core.RequestEvent) error {
		count, err := event.App.CountRecords(collectionName)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		owners, err := event.App.FindRecordsByFilter(collectionName, "is_owner = true", "", 1, 0)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		hasUsableOwner := len(owners) > 0 && owners[0].GetBool("active")
		return event.JSON(http.StatusOK, map[string]any{
			"needsOwner":             count == 0,
			"hasUsableOwner":         hasUsableOwner,
			"ownerRecoveryAvailable": len(owners) > 0 && isLoopback(event.RemoteIP()),
			"joinUrl":                owner.JoinURL(event.App),
			"version":                applicationVersion,
		})
	}
}

func recoverOwner(event *core.RequestEvent) error {
	if !isLoopback(event.RemoteIP()) {
		return httpx.WriteError(event, result.Forbidden("setup.loopback_only", "Owner recovery is available only on the host computer."))
	}
	var request ownerRecoveryRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("request.invalid", "The submitted recovery details are invalid.", nil))
	}
	if request.Confirmation != "RECOVER OWNERSHIP" {
		return httpx.WriteError(event, result.Invalid("setup.recovery_confirmation_required", `Type "RECOVER OWNERSHIP" to confirm.`, nil))
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
		return httpx.WriteError(event, result.Invalid("setup.recovery_invalid", "Please correct the highlighted recovery details.", fields))
	}

	backupName := "pre_owner_recovery_sgh_" + time.Now().UTC().Format("20060102_150405") + ".zip"
	if err := event.App.CreateBackup(event.Request.Context(), backupName); err != nil {
		event.App.Logger().Error("owner recovery backup failed", "error", err)
		return httpx.WriteError(event, result.Conflict("setup.recovery_backup_failed", "A recovery backup could not be created, so ownership was not changed."))
	}
	replacement, err := replaceOwners(event.App, username, displayName, request.Password, event.Get(httpx.TraceIDKey))
	if err != nil {
		event.App.Logger().Error("owner recovery failed", "error", err)
		return httpx.WriteError(event, result.Internal(err))
	}
	token, err := replacement.NewAuthToken()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, map[string]any{"token": token, "gameMaster": map[string]any{"id": replacement.Id, "username": username, "displayName": displayName, "isOwner": true}})
}

// replaceOwners transfers every game-master relation before removing only owner accounts.
// A recovery backup is deliberately created by the route before this transaction begins.
func replaceOwners(app core.App, username, displayName, password string, traceID any) (*core.Record, error) {
	masters, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return nil, err
	}
	obsolete, err := app.FindRecordsByFilter(masters, "is_owner = true", "", 0, 0)
	if err != nil {
		return nil, err
	}
	if len(obsolete) == 0 {
		return nil, result.Conflict("setup.recovery_unavailable", "Owner recovery is unavailable until an owner account exists.")
	}
	references, err := app.FindCollectionReferences(masters)
	if err != nil {
		return nil, err
	}
	var replacement *core.Record
	err = app.RunInTransaction(func(tx core.App) error {
		collection, err := tx.FindCollectionByNameOrId(collectionName)
		if err != nil {
			return err
		}
		replacement = core.NewRecord(collection)
		replacement.Set("username", username)
		replacement.Set("display_name", displayName)
		replacement.Set("is_owner", true)
		replacement.Set("active", true)
		replacement.Set("last_login_at", time.Now().UTC())
		replacement.SetPassword(password)
		if err := tx.Save(replacement); err != nil {
			return err
		}
		for collection, fields := range references {
			records, err := tx.FindRecordsByFilter(collection, "", "", 0, 0)
			if err != nil {
				return err
			}
			for _, record := range records {
				changed := false
				for _, field := range fields {
					relation, ok := field.(*core.RelationField)
					if !ok {
						continue
					}
					if relation.IsMultiple() {
						values := record.GetStringSlice(relation.Name)
						for i, value := range values {
							if isObsoleteOwner(obsolete, value) {
								values[i] = replacement.Id
								changed = true
							}
						}
						if changed {
							record.Set(relation.Name, values)
						}
					} else if isObsoleteOwner(obsolete, record.GetString(relation.Name)) {
						record.Set(relation.Name, replacement.Id)
						changed = true
					}
				}
				if changed {
					if err := tx.Save(record); err != nil {
						return err
					}
				}
			}
		}
		for _, record := range obsolete {
			if err := tx.Delete(record); err != nil {
				return err
			}
		}
		return applicationaudit.Record(tx, replacement, "", "owner.recovered", "game_master", replacement.Id, map[string]any{"replacedOwnerCount": len(obsolete)}, traceID)
	})
	return replacement, err
}

func isObsoleteOwner(records []*core.Record, id string) bool {
	for _, record := range records {
		if record.Id == id {
			return true
		}
	}
	return false
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
