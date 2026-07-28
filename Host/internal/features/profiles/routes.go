package profiles

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	platformaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/audit"
	platformauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/auth"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	platformrealtime "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const (
	requestsCollection = "profile_requests"
	requestLifetime    = 10 * time.Minute
)

type requestProfileRequest struct {
	DisplayName string `json:"displayName"`
}

type redeemRequest struct {
	Secret string `json:"secret"`
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

type updateProfileRequest struct {
	DisplayName *string `json:"displayName"`
	Biography   *string `json:"bio"`
	Accent      *string `json:"accent"`
}

func Register(event *core.ServeEvent) {
	authGroup := event.Router.Group("/api/app/v1/auth/player/requests")
	authGroup.POST("", requestProfile)
	authGroup.GET("/{requestId}", requestStatus)
	authGroup.POST("/{requestId}/redeem", redeem)

	profileGroup := event.Router.Group("/api/app/v1/profiles")
	profileGroup.BindFunc(platformauth.RequirePlayer)
	profileGroup.GET("/me", me)
	profileGroup.PATCH("/me", updateMe)
	profileGroup.POST("/me/avatar", updateAvatar)
	profileGroup.GET("/me/history", privateHistory)
	profileGroup.GET("/{profileId}/summary", summary)
	event.Router.GET("/api/app/v1/profiles/{profileId}/avatar", avatar)

	adminGroup := event.Router.Group("/api/app/v1/admin")
	adminGroup.BindFunc(platformauth.RequireGameMaster)
	adminGroup.GET("/profile-requests", pendingRequests)
	adminGroup.GET("/profiles", listProfiles)
	adminGroup.GET("/profiles/{profileId}", adminProfileDetail)
	adminGroup.POST("/profile-requests/{requestId}/approve", approve)
	adminGroup.POST("/profile-requests/{requestId}/reject", reject)
	adminGroup.POST("/profiles/{profileId}/disable", disable)
	adminGroup.POST("/profiles/{profileId}/restore", restore)
}

func listProfiles(event *core.RequestEvent) error {
	records, err := event.App.FindRecordsByFilter(
		platformauth.PlayerProfilesCollection,
		"",
		"display_name",
		500,
		0,
	)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projected := make([]map[string]any, len(records))
	for index, record := range records {
		projected[index] = projectAdminProfile(record)
	}
	return event.JSON(http.StatusOK, projected)
}

func RegisterJobs(app core.App) {
	app.Cron().MustAdd("sgh_expire_profile_requests", "* * * * *", func() {
		records, err := app.FindRecordsByFilter(
			requestsCollection,
			"status = 'pending' && expires_at < {:now}",
			"",
			100,
			0,
			dbx.Params{"now": time.Now().UTC()},
		)
		if err != nil {
			app.Logger().Error("failed to load expired profile requests", "error", err)
			return
		}
		for _, record := range records {
			record.Set("status", "expired")
			if err := app.Save(record); err != nil {
				app.Logger().Error("failed to expire profile request", "requestId", record.Id, "error", err)
			} else {
				publishProfileRequest(app, record, "profile_request.expired")
			}
		}
	})
}

func requestProfile(event *core.RequestEvent) error {
	var request requestProfileRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("profile_request.invalid", "Enter the profile name you want to use.", nil))
	}
	displayName, normalizedName, err := NormalizeName(request.DisplayName)
	if err != nil {
		return httpx.WriteError(event, result.Invalid("profile_request.invalid_name", "Enter a valid profile name.", result.FieldErrors{
			"displayName": {err.Error()},
		}))
	}

	secret, err := newSecret()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	collection, err := event.App.FindCollectionByNameOrId(requestsCollection)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	record := core.NewRecord(collection)
	record.Set("requested_name", displayName)
	record.Set("normalized_name", normalizedName)
	record.Set("secret_hash", hashSecret(secret))
	record.Set("status", "pending")
	record.Set("expires_at", time.Now().UTC().Add(requestLifetime))

	existing, existingErr := event.App.FindFirstRecordByData(platformauth.PlayerProfilesCollection, "normalized_name", normalizedName)
	if existingErr == nil {
		record.Set("request_type", "recover")
		record.Set("profile", existing.Id)
	} else {
		record.Set("request_type", "new")
	}
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	publishProfileRequest(event.App, record, "profile_request.created")
	return event.JSON(http.StatusAccepted, map[string]any{
		"requestId":     record.Id,
		"secret":        secret,
		"realtimeTopic": profileRequestTopic(record),
		"status":        "pending",
		"expiresAt":     record.GetDateTime("expires_at").Time().UTC(),
	})
}

func requestStatus(event *core.RequestEvent) error {
	record, appError := authorizedRequest(event)
	if appError != nil {
		return httpx.WriteError(event, *appError)
	}
	expireIfNeeded(event.App, record)
	return event.JSON(http.StatusOK, map[string]any{
		"requestId": record.Id,
		"status":    record.GetString("status"),
		"expiresAt": record.GetDateTime("expires_at").Time().UTC(),
		"reason":    record.GetString("rejection_reason"),
	})
}

func redeem(event *core.RequestEvent) error {
	var request redeemRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("profile_request.invalid", "The approval secret is missing.", nil))
	}
	requestID := event.Request.PathValue("requestId")
	var profile *core.Record
	transactionErr := event.App.RunInTransaction(func(txApp core.App) error {
		record, err := txApp.FindRecordById(requestsCollection, requestID)
		if err != nil || !secretMatches(record, request.Secret) {
			return result.AppError{Code: "profile_request.not_found", Message: "This profile request is unavailable.", Status: http.StatusNotFound}
		}
		if expireIfNeeded(txApp, record) {
			return result.AppError{Code: "profile_request.expired", Message: "This profile request has expired.", Status: http.StatusGone}
		}
		if record.GetString("status") != "approved" {
			return result.AppError{Code: "profile_request.not_approved", Message: "This profile request has not been approved.", Status: http.StatusConflict}
		}
		profile, err = txApp.FindRecordById(platformauth.PlayerProfilesCollection, record.GetString("profile"))
		if err != nil || !profile.GetBool("active") {
			return result.AppError{Code: "profile.disabled", Message: "This profile is not available.", Status: http.StatusForbidden}
		}
		record.Set("status", "consumed")
		record.Set("consumed_at", time.Now().UTC())
		return txApp.Save(record)
	})
	if transactionErr != nil {
		if appError, ok := transactionErr.(result.AppError); ok {
			return httpx.WriteError(event, appError)
		}
		return httpx.WriteError(event, result.Internal(transactionErr))
	}
	token, err := profile.NewAuthToken()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	publishProfileRequest(event.App, approvedRequestRecord(event.App, requestID), "profile_request.consumed")
	return event.JSON(http.StatusOK, map[string]any{
		"token":   token,
		"profile": projectPrivateProfile(profile),
	})
}

func pendingRequests(event *core.RequestEvent) error {
	records, err := event.App.FindRecordsByFilter(requestsCollection, "status = 'pending'", "-created", 100, 0)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	response := make([]map[string]any, 0, len(records))
	for _, record := range records {
		expireIfNeeded(event.App, record)
		if record.GetString("status") != "pending" {
			continue
		}
		response = append(response, map[string]any{
			"id":            record.Id,
			"requestType":   record.GetString("request_type"),
			"requestedName": record.GetString("requested_name"),
			"createdAt":     record.GetDateTime("created").Time().UTC(),
			"expiresAt":     record.GetDateTime("expires_at").Time().UTC(),
		})
	}
	return event.JSON(http.StatusOK, response)
}

func approve(event *core.RequestEvent) error {
	requestID := event.Request.PathValue("requestId")
	var approvedProfile *core.Record
	err := event.App.RunInTransaction(func(txApp core.App) error {
		requestRecord, err := txApp.FindRecordById(requestsCollection, requestID)
		if err != nil {
			return result.AppError{Code: "profile_request.not_found", Message: "Profile request not found.", Status: http.StatusNotFound}
		}
		if expireIfNeeded(txApp, requestRecord) {
			return result.AppError{Code: "profile_request.expired", Message: "This profile request has expired.", Status: http.StatusGone}
		}
		if requestRecord.GetString("status") != "pending" {
			return result.AppError{Code: "profile_request.already_decided", Message: "This request has already been decided.", Status: http.StatusConflict}
		}

		if requestRecord.GetString("request_type") == "recover" {
			approvedProfile, err = txApp.FindRecordById(platformauth.PlayerProfilesCollection, requestRecord.GetString("profile"))
			if err != nil {
				return err
			}
			approvedProfile.RefreshTokenKey()
			approvedProfile.Set("active", true)
			approvedProfile.Set("approved_at", time.Now().UTC())
			approvedProfile.Set("approved_by", event.Auth.Id)
			if err := txApp.Save(approvedProfile); err != nil {
				return err
			}
		} else {
			collection, err := txApp.FindCollectionByNameOrId(platformauth.PlayerProfilesCollection)
			if err != nil {
				return err
			}
			password, err := newSecret()
			if err != nil {
				return err
			}
			approvedProfile = core.NewRecord(collection)
			approvedProfile.Set("display_name", requestRecord.GetString("requested_name"))
			approvedProfile.Set("normalized_name", requestRecord.GetString("normalized_name"))
			approvedProfile.Set("active", true)
			approvedProfile.Set("approved_at", time.Now().UTC())
			approvedProfile.Set("approved_by", event.Auth.Id)
			approvedProfile.SetPassword(password)
			if err := txApp.Save(approvedProfile); err != nil {
				return err
			}
			requestRecord.Set("profile", approvedProfile.Id)
		}

		requestRecord.Set("status", "approved")
		requestRecord.Set("decided_by", event.Auth.Id)
		requestRecord.Set("decided_at", time.Now().UTC())
		return txApp.Save(requestRecord)
	})
	if err != nil {
		if appError, ok := err.(result.AppError); ok {
			return httpx.WriteError(event, appError)
		}
		return httpx.WriteError(event, result.Internal(err))
	}
	approvedRequest := approvedRequestRecord(event.App, requestID)
	publishProfileRequest(event.App, approvedRequest, "profile_request.approved")
	profileEventKind := "profile.updated"
	if approvedRequest != nil && approvedRequest.GetString("request_type") == "recover" {
		profileEventKind = "profile.recovered"
	}
	publishProfile(event.App, approvedProfile, profileEventKind)
	_ = platformaudit.Record(event.App, event.Auth, "", "profile_request.approved", "player_profile", approvedProfile.Id,
		map[string]any{"requestId": requestID}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusOK, map[string]any{
		"requestId": requestID,
		"status":    "approved",
		"profile":   projectAdminProfile(approvedProfile),
	})
}

func reject(event *core.RequestEvent) error {
	var request rejectRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("profile_request.invalid", "The rejection could not be read.", nil))
	}
	request.Reason = strings.TrimSpace(request.Reason)
	if len([]rune(request.Reason)) > 280 {
		return httpx.WriteError(event, result.Invalid("profile_request.invalid_reason", "The rejection reason is too long.", nil))
	}
	record, err := event.App.FindRecordById(requestsCollection, event.Request.PathValue("requestId"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "profile_request.not_found", Message: "Profile request not found.", Status: http.StatusNotFound})
	}
	if record.GetString("status") != "pending" {
		return httpx.WriteError(event, result.Conflict("profile_request.already_decided", "This request has already been decided."))
	}
	record.Set("status", "rejected")
	record.Set("rejection_reason", request.Reason)
	record.Set("decided_by", event.Auth.Id)
	record.Set("decided_at", time.Now().UTC())
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	publishProfileRequest(event.App, record, "profile_request.rejected")
	_ = platformaudit.Record(event.App, event.Auth, "", "profile_request.rejected", "profile_request", record.Id,
		nil, event.Get(httpx.TraceIDKey))
	return event.NoContent(http.StatusNoContent)
}

func me(event *core.RequestEvent) error {
	event.Auth.Set("last_seen_at", time.Now().UTC())
	if err := event.App.Save(event.Auth); err != nil {
		event.App.Logger().Warn("failed to update profile last seen time", "profileId", event.Auth.Id, "error", err)
	}
	publishProfile(event.App, event.Auth, "profile.updated")
	return event.JSON(http.StatusOK, projectPrivateProfile(event.Auth))
}

func updateMe(event *core.RequestEvent) error {
	var request updateProfileRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("profile.invalid", "The profile changes could not be read.", nil))
	}
	if request.DisplayName != nil {
		displayName, normalizedName, err := NormalizeName(*request.DisplayName)
		if err != nil {
			return httpx.WriteError(event, result.Invalid("profile.invalid_name", "Enter a valid profile name.", result.FieldErrors{"displayName": {err.Error()}}))
		}
		event.Auth.Set("display_name", displayName)
		event.Auth.Set("normalized_name", normalizedName)
	}
	if request.Biography != nil {
		biography := strings.TrimSpace(*request.Biography)
		if len([]rune(biography)) > MaxBiographyLength {
			return httpx.WriteError(event, result.Invalid("profile.invalid_bio", "Biography must be 280 characters or fewer.", nil))
		}
		event.Auth.Set("bio", biography)
	}
	if request.Accent != nil {
		event.Auth.Set("accent", *request.Accent)
	}
	if err := event.App.Save(event.Auth); err != nil {
		return httpx.WriteError(event, result.Invalid("profile.save_failed", "The profile could not be saved.", nil))
	}
	if request.DisplayName != nil {
		if err := syncLiveParticipantNames(event.App, event.Auth); err != nil {
			event.App.Logger().Error("failed to update live participant display names", "profileId", event.Auth.Id, "error", err)
		}
	}
	publishProfile(event.App, event.Auth, "profile.updated")
	return event.JSON(http.StatusOK, projectPrivateProfile(event.Auth))
}

func syncLiveParticipantNames(app core.App, profile *core.Record) error {
	participants, err := app.FindRecordsByFilter("participants",
		"profile = {:profile} && status != 'kicked' && status != 'left'", "", 1000, 0,
		dbx.Params{"profile": profile.Id})
	if err != nil {
		return err
	}

	for _, participant := range participants {
		game, err := app.FindRecordById("games", participant.GetString("game"))
		if err != nil || !isLiveGame(game.GetString("status")) {
			continue
		}
		participant.Set("display_name_snapshot", profile.GetString("display_name"))
		if err := app.Save(participant); err != nil {
			return err
		}
		publishLiveParticipantNameChange(app, game, participant)
	}
	return nil
}

func isLiveGame(status string) bool {
	return status == "lobby" || status == "running" || status == "paused"
}

func publishLiveParticipantNameChange(app core.App, game, participant *core.Record) {
	event := platformrealtime.Event[map[string]any]{
		EventID: platformrealtime.NewEventID(), GameID: game.Id,
		Kind: "participant.updated", Payload: map[string]any{
			"id": participant.Id, "profileId": participant.GetString("profile"),
			"displayNameSnapshot": participant.GetString("display_name_snapshot"),
			"gameAlias":           participant.GetString("game_alias"),
			"seatNumber":          participant.GetInt("seat_number"), "status": participant.GetString("status"),
		},
	}
	for _, topic := range []string{"game:" + game.Id + ":public", "game:" + game.Id + ":game-masters"} {
		_ = platformrealtime.Publish(app, topic, event, func(auth *core.Record) bool {
			if auth == nil || !auth.GetBool("active") {
				return false
			}
			if auth.Collection().Name == platformauth.GameMastersCollection {
				return true
			}
			return auth.Collection().Name == platformauth.PlayerProfilesCollection && auth.Id == participant.GetString("profile")
		})
	}
}

func updateAvatar(event *core.RequestEvent) error {
	files, err := event.FindUploadedFiles("file")
	if err != nil || len(files) != 1 {
		return httpx.WriteError(event, result.Invalid("profile.avatar_required", "Choose one JPEG, PNG, or WebP image.", nil))
	}
	file := files[0]
	if file.Size <= 0 || file.Size > 1<<20 {
		return httpx.WriteError(event, result.Invalid("profile.avatar_too_large", "Profile images must be 1 MB or smaller.", nil))
	}
	reader, err := file.Reader.Open()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	content, readErr := io.ReadAll(io.LimitReader(reader, (1<<20)+1))
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || len(content) > 1<<20 {
		return httpx.WriteError(event, result.Invalid("profile.avatar_invalid", "The profile image could not be read safely.", nil))
	}
	if _, width, height, err := rulesets.InspectProfileImageUpload(content); err != nil {
		return httpx.WriteError(event, result.Invalid("profile.avatar_invalid", err.Error(), nil))
	} else if width > 512 || height > 512 {
		return httpx.WriteError(event, result.Invalid("profile.avatar_dimensions", "Profile images must be 512 by 512 pixels or smaller.", nil))
	}
	event.Auth.Set("avatar", file)
	if err := event.App.Save(event.Auth); err != nil {
		return httpx.WriteError(event, result.Invalid("profile.avatar_save_failed", "The profile image could not be saved.", nil))
	}
	return event.JSON(http.StatusOK, projectPrivateProfile(event.Auth))
}

func avatar(event *core.RequestEvent) error {
	profile, err := event.App.FindRecordById(platformauth.PlayerProfilesCollection, event.Request.PathValue("profileId"))
	if err != nil || !profile.GetBool("active") || profile.GetString("avatar") == "" {
		return httpx.WriteError(event, result.AppError{Code: "profile.avatar_not_found", Message: "Profile image not found.", Status: http.StatusNotFound})
	}
	if event.Auth == nil || !event.Auth.GetBool("active") {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "Sign in to view this profile image.", Status: http.StatusUnauthorized})
	}
	if event.Auth.Collection().Name != platformauth.GameMastersCollection &&
		(event.Auth.Collection().Name != platformauth.PlayerProfilesCollection ||
			!partyMembers(event.App, event.Auth.Id, profile.Id)) {
		return httpx.WriteError(event, result.Forbidden("profile.forbidden", "This profile image is visible only to party members."))
	}
	fsys, err := event.App.NewFilesystem()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer fsys.Close()
	reader, err := fsys.GetReader(profile.BaseFilesPath() + "/" + profile.GetString("avatar"))
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer reader.Close()
	content, err := io.ReadAll(io.LimitReader(reader, (1<<20)+1))
	if err != nil || len(content) > 1<<20 {
		return httpx.WriteError(event, result.Internal(errors.New("stored profile image exceeds limit")))
	}
	mimeType, err := rulesets.InspectProfileImage(content)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	event.Response.Header().Set("Content-Disposition", `inline; filename="`+filepath.Base(profile.GetString("avatar"))+`"`)
	event.Response.Header().Set("Cache-Control", "private, max-age=300")
	return event.Blob(http.StatusOK, mimeType, content)
}

func summary(event *core.RequestEvent) error {
	profile, err := event.App.FindRecordById(platformauth.PlayerProfilesCollection, event.Request.PathValue("profileId"))
	if err != nil || !profile.GetBool("active") {
		return httpx.WriteError(event, result.AppError{Code: "profile.not_found", Message: "Profile not found.", Status: http.StatusNotFound})
	}
	if !partyMembers(event.App, event.Auth.Id, profile.Id) {
		return httpx.WriteError(event, result.Forbidden("profile.forbidden", "This profile is visible only to party members."))
	}
	projection, err := projectPublicProfile(event.App, profile)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, projection)
}

func privateHistory(event *core.RequestEvent) error {
	projection, err := historyProjection(event.App, event.Auth)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, projection)
}

func adminProfileDetail(event *core.RequestEvent) error {
	profile, err := event.App.FindRecordById(platformauth.PlayerProfilesCollection, event.Request.PathValue("profileId"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "profile.not_found", Message: "Profile not found.", Status: http.StatusNotFound})
	}
	projection, err := historyProjection(event.App, profile)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projection["profile"] = projectAdminProfile(profile)
	return event.JSON(http.StatusOK, projection)
}

func historyProjection(app core.App, profile *core.Record) (map[string]any, error) {
	participants, err := app.FindRecordsByFilter("participants", "profile = {:profile}", "-created", 500, 0,
		dbx.Params{"profile": profile.Id})
	if err != nil {
		return nil, err
	}
	games := make([]map[string]any, 0, len(participants))
	for _, participant := range participants {
		game, err := app.FindRecordById("games", participant.GetString("game"))
		if err != nil || (game.GetString("status") != "review" && game.GetString("status") != "archived") {
			continue
		}
		roleName, rulesetName := historicalNames(game, participant.GetString("role_key"))
		awards, err := app.FindRecordsByFilter("achievement_awards",
			"profile = {:profile} && game = {:game}", "created", 100, 0,
			dbx.Params{"profile": profile.Id, "game": game.Id})
		if err != nil {
			return nil, err
		}
		projectedAwards := make([]map[string]any, len(awards))
		for index, award := range awards {
			projectedAwards[index] = map[string]any{
				"id": award.Id, "achievementId": award.GetString("achievement_key"),
				"title": award.GetString("title_snapshot"), "description": award.GetString("description_snapshot"),
				"assetKey": award.GetString("asset_key"), "points": award.GetInt("points_snapshot"),
				"awardedAt": award.GetDateTime("created").Time().UTC(),
			}
		}
		games = append(games, map[string]any{
			"id": game.Id, "name": game.GetString("name"), "rulesetName": rulesetName,
			"rulesetVersionId": game.GetString("ruleset_version"), "roleName": roleName,
			"outcome": participant.GetString("outcome"), "startedAt": dateOrNil(game, "started_at"),
			"endedAt": dateOrNil(game, "ended_at"), "achievements": projectedAwards,
		})
	}
	statistics, err := profileStatistics(app, profile.Id, true)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"profile":    projectPrivateProfile(profile),
		"games":      games,
		"statistics": statistics,
	}, nil
}

func partyMembers(app core.App, viewerProfileID, targetProfileID string) bool {
	if viewerProfileID == targetProfileID {
		return true
	}
	participants, err := app.FindRecordsByFilter(
		"participants",
		"profile = {:profile}",
		"",
		1000,
		0,
		dbx.Params{"profile": viewerProfileID},
	)
	if err != nil {
		return false
	}
	for _, participant := range participants {
		count, err := app.CountRecords("participants", dbx.HashExp{
			"game": participant.GetString("game"), "profile": targetProfileID,
		})
		if err == nil && count > 0 {
			return true
		}
	}
	return false
}

func disable(event *core.RequestEvent) error {
	return setActive(event, false)
}

func restore(event *core.RequestEvent) error {
	return setActive(event, true)
}

func setActive(event *core.RequestEvent, active bool) error {
	profile, err := event.App.FindRecordById(platformauth.PlayerProfilesCollection, event.Request.PathValue("profileId"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "profile.not_found", Message: "Profile not found.", Status: http.StatusNotFound})
	}
	profile.Set("active", active)
	profile.RefreshTokenKey()
	if err := event.App.Save(profile); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	kind := "profile.disabled"
	if active {
		kind = "profile.recovered"
	}
	publishProfile(event.App, profile, kind)
	_ = platformaudit.Record(event.App, event.Auth, "", kind, "player_profile", profile.Id,
		map[string]any{"active": active}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusOK, projectAdminProfile(profile))
}

func approvedRequestRecord(app core.App, requestID string) *core.Record {
	record, _ := app.FindRecordById(requestsCollection, requestID)
	return record
}

func profileRequestTopic(record *core.Record) string {
	if record == nil {
		return ""
	}
	return "profile-request:" + record.Id + ":" + record.GetString("secret_hash")
}

func publishProfileRequest(app core.App, record *core.Record, kind string) {
	if record == nil {
		return
	}
	payload := map[string]any{
		"requestId": record.Id,
		"status":    record.GetString("status"),
		"reason":    record.GetString("rejection_reason"),
		"expiresAt": record.GetDateTime("expires_at").Time().UTC(),
	}
	event := platformrealtime.Event[map[string]any]{
		EventID: platformrealtime.NewEventID(), Kind: kind, Payload: payload,
	}
	if err := platformrealtime.Publish(app, profileRequestTopic(record), event, func(*core.Record) bool { return true }); err != nil {
		app.Logger().Error("profile request realtime publication failed", "requestId", record.Id, "error", err)
	}
	if err := platformrealtime.Publish(app, "profile-requests:game-masters", event, func(auth *core.Record) bool {
		return auth != nil && auth.Collection().Name == platformauth.GameMastersCollection && auth.GetBool("active")
	}); err != nil {
		app.Logger().Error("profile request admin publication failed", "requestId", record.Id, "error", err)
	}
}

func publishProfile(app core.App, profile *core.Record, kind string) {
	if profile == nil {
		return
	}
	event := platformrealtime.Event[map[string]any]{
		EventID: platformrealtime.NewEventID(),
		Kind:    kind,
		Payload: map[string]any{"profileId": profile.Id, "active": profile.GetBool("active")},
	}
	if err := platformrealtime.Publish(app, "profile:"+profile.Id, event, func(auth *core.Record) bool {
		return auth != nil && (auth.Id == profile.Id ||
			(auth.Collection().Name == platformauth.GameMastersCollection && auth.GetBool("active")))
	}); err != nil {
		app.Logger().Error("profile realtime publication failed", "profileId", profile.Id, "error", err)
	}
}

func authorizedRequest(event *core.RequestEvent) (*core.Record, *result.AppError) {
	secret := event.Request.Header.Get("X-Profile-Request-Secret")
	record, err := event.App.FindRecordById(requestsCollection, event.Request.PathValue("requestId"))
	if err != nil || !secretMatches(record, secret) {
		appError := result.AppError{Code: "profile_request.not_found", Message: "This profile request is unavailable.", Status: http.StatusNotFound}
		return nil, &appError
	}
	return record, nil
}

func expireIfNeeded(app core.App, record *core.Record) bool {
	if record.GetString("status") != "pending" || record.GetDateTime("expires_at").Time().After(time.Now().UTC()) {
		return record.GetString("status") == "expired"
	}
	record.Set("status", "expired")
	if err := app.Save(record); err != nil {
		app.Logger().Error("failed to expire profile request", "requestId", record.Id, "error", err)
	}
	return true
}

func newSecret() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashSecret(secret string) string {
	digest := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(digest[:])
}

func secretMatches(record *core.Record, secret string) bool {
	if record == nil || secret == "" {
		return false
	}
	expected := record.GetString("secret_hash")
	actual := hashSecret(secret)
	return len(expected) == len(actual) && subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

func projectPrivateProfile(record *core.Record) map[string]any {
	return map[string]any{
		"id":          record.Id,
		"displayName": record.GetString("display_name"),
		"avatar":      avatarURL(record),
		"bio":         record.GetString("bio"),
		"accent":      record.GetString("accent"),
		"active":      record.GetBool("active"),
	}
}

func projectPublicProfile(app core.App, record *core.Record) (map[string]any, error) {
	participants, err := app.FindRecordsByFilter("participants", "profile = {:profile}", "", 1000, 0,
		dbx.Params{"profile": record.Id})
	if err != nil {
		return nil, err
	}
	statistics := map[string]int{
		"gamesPlayed": len(participants), "completedGames": 0, "wins": 0, "draws": 0,
		"achievementCount": 0, "achievementPoints": 0,
	}
	for _, participant := range participants {
		game, err := app.FindRecordById("games", participant.GetString("game"))
		if err != nil || game.GetString("status") != "archived" {
			continue
		}
		statistics["completedGames"]++
		switch participant.GetString("outcome") {
		case "win":
			statistics["wins"]++
		case "draw":
			statistics["draws"]++
		}
	}
	awards, err := app.FindRecordsByFilter("achievement_awards", "profile = {:profile}", "-created", 100, 0,
		dbx.Params{"profile": record.Id})
	if err != nil {
		return nil, err
	}
	projectedAwards := make([]map[string]any, 0, len(awards))
	for _, award := range awards {
		if !awardVisible(app, award) {
			continue
		}
		points := award.GetInt("points_snapshot")
		statistics["achievementCount"]++
		statistics["achievementPoints"] += points
		projectedAwards = append(projectedAwards, map[string]any{
			"id": award.Id, "achievementId": award.GetString("achievement_key"),
			"title": award.GetString("title_snapshot"), "description": award.GetString("description_snapshot"),
			"assetKey": award.GetString("asset_key"), "points": points,
			"awardedAt": award.GetDateTime("created").Time().UTC(),
		})
	}
	projection := map[string]any{
		"id":           record.Id,
		"displayName":  record.GetString("display_name"),
		"avatar":       avatarURL(record),
		"bio":          record.GetString("bio"),
		"accent":       record.GetString("accent"),
		"statistics":   statistics,
		"achievements": projectedAwards,
	}
	if statistics["completedGames"] > 0 {
		projection["winRate"] = float64(statistics["wins"]) / float64(statistics["completedGames"])
	}
	return projection, nil
}

func profileStatistics(app core.App, profileID string, completedOnly bool) (map[string]int, error) {
	awards, err := app.FindRecordsByFilter("achievement_awards", "profile = {:profile}", "", 1000, 0,
		dbx.Params{"profile": profileID})
	if err != nil {
		return nil, err
	}
	statistics := map[string]int{"achievementCount": 0, "achievementPoints": 0}
	for _, award := range awards {
		if completedOnly && !awardVisible(app, award) {
			continue
		}
		statistics["achievementCount"]++
		statistics["achievementPoints"] += award.GetInt("points_snapshot")
	}
	return statistics, nil
}

func awardVisible(app core.App, award *core.Record) bool {
	if !award.GetBool("hidden_until_game_completed") {
		return true
	}
	game, err := app.FindRecordById("games", award.GetString("game"))
	if err != nil {
		return false
	}
	status := game.GetString("status")
	return status == "review" || status == "archived"
}

func avatarURL(record *core.Record) string {
	if record.GetString("avatar") == "" {
		return ""
	}
	version := record.GetDateTime("updated").Time().UTC().Unix()
	return "/api/app/v1/profiles/" + record.Id + "/avatar?v=" + fmt.Sprint(version)
}

func projectAdminProfile(record *core.Record) map[string]any {
	projection := projectPrivateProfile(record)
	projection["normalizedName"] = record.GetString("normalized_name")
	projection["approvedAt"] = record.GetDateTime("approved_at").Time().UTC()
	return projection
}

func historicalNames(game *core.Record, roleKey string) (string, string) {
	data, err := json.Marshal(game.Get("ruleset_snapshot"))
	if err != nil {
		return "", ""
	}
	var definition rulesets.DefinitionV1
	if json.Unmarshal(data, &definition) != nil {
		return "", ""
	}
	roleName := ""
	for _, role := range definition.Roles {
		if role.ID == roleKey {
			roleName = role.Name
			break
		}
	}
	return roleName, definition.Metadata.Name
}

func dateOrNil(record *core.Record, field string) any {
	value := record.GetDateTime(field)
	if value.IsZero() {
		return nil
	}
	return value.Time().UTC()
}
