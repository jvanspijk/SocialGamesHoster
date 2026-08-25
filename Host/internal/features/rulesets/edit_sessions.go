package rulesets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const editSessionLifetime = 7 * 24 * time.Hour

type effectiveAsset struct {
	key, kind, displayName, accessibilityText, mimeType, checksum string
	metadata                                                      any
	record                                                        *core.Record
	staged                                                        bool
}

type assetUsage struct {
	Label   string `json:"label"`
	Section string `json:"section"`
	ItemID  string `json:"itemId,omitempty"`
}

func registerEditSessionRoutes(group *router.RouterGroup[*core.RequestEvent]) {
	group.POST("/rulesets/{id}/edit-session", openEditSession)
	group.DELETE("/rulesets/{id}/edit-session/{sessionId}", discardEditSession)
	group.GET("/rulesets/{id}/edit-session/{sessionId}/assets", listSessionAssets)
	group.POST("/rulesets/{id}/edit-session/{sessionId}/assets", uploadSessionAsset)
	group.PATCH("/rulesets/{id}/edit-session/{sessionId}/assets/{assetKey}", updateSessionAsset)
	group.DELETE("/rulesets/{id}/edit-session/{sessionId}/assets/{assetKey}", deleteSessionAsset)
	group.GET("/ruleset-edit-assets/{id}", previewSessionAsset)
}

func openEditSession(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil || logical.GetBool("archived") {
		return rulesetNotFound(event)
	}
	if err := cleanupExpiredEditSessions(event.App, time.Now().UTC()); err != nil {
		event.App.Logger().Error("expired ruleset edit session cleanup failed", "error", err)
	}
	now := time.Now().UTC()
	items, err := event.App.FindRecordsByFilter("ruleset_edit_sessions", "ruleset = {:ruleset} && creator = {:creator}", "", 1, 0, dbx.Params{"ruleset": logical.Id, "creator": event.Auth.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	var session *core.Record
	if len(items) == 1 && items[0].GetDateTime("expires_at").Time().After(now) {
		session = items[0]
	} else {
		if len(items) == 1 {
			_ = deleteEditSession(event.App, items[0])
		}
		base, findErr := latestSavedVersion(event.App, logical)
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		collection, findErr := event.App.FindCollectionByNameOrId("ruleset_edit_sessions")
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		session = core.NewRecord(collection)
		session.Set("ruleset", logical.Id)
		session.Set("base_version", base.Id)
		session.Set("creator", event.Auth.Id)
	}
	touchEditSession(session, now)
	if err := event.App.Save(session); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projection := projectEditSession(session)
	changes, _ := event.App.CountRecords("ruleset_asset_changes", dbx.HashExp{"session": session.Id})
	projection["hasChanges"] = changes > 0
	return event.JSON(http.StatusOK, projection)
}

func discardEditSession(event *core.RequestEvent) error {
	session, err := ownedEditSession(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	if err := deleteEditSession(event.App, session); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.NoContent(http.StatusNoContent)
}

func listSessionAssets(event *core.RequestEvent) error {
	session, err := ownedEditSession(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	assets, err := effectiveSessionAssets(event.App, session)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	definition, _ := sessionDefinition(event)
	response := make([]map[string]any, 0, len(assets))
	for _, asset := range assets {
		response = append(response, projectEffectiveAsset(asset, scanAssetUsages(definition, asset.key)))
	}
	touchEditSession(session, time.Now().UTC())
	_ = event.App.Save(session)
	return event.JSON(http.StatusOK, response)
}

func uploadSessionAsset(event *core.RequestEvent) error {
	session, err := ownedEditSession(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	kind := strings.TrimSpace(event.Request.FormValue("kind"))
	mode := strings.TrimSpace(event.Request.FormValue("mode"))
	key := strings.TrimSpace(event.Request.FormValue("assetKey"))
	displayName := strings.TrimSpace(event.Request.FormValue("displayName"))
	accessibilityText := strings.TrimSpace(event.Request.FormValue("accessibilityText"))
	if kind != "image" && kind != "audio" {
		return httpx.WriteError(event, result.Invalid("asset.invalid_kind", "Choose an image or audio file.", nil))
	}
	if mode != "replace" {
		mode = "add"
		key, err = newOpaqueID("asset")
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	current, err := effectiveSessionAssets(event.App, session)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if mode == "replace" {
		asset, ok := current[key]
		if !ok {
			return httpx.WriteError(event, result.AppError{Code: "asset.not_found", Message: "Media item not found.", Status: http.StatusNotFound})
		}
		if asset.kind != kind {
			return httpx.WriteError(event, result.Invalid("asset.kind_mismatch", "Replace an image with an image or audio with audio.", nil))
		}
		if displayName == "" {
			displayName = asset.displayName
		}
		if accessibilityText == "" {
			accessibilityText = asset.accessibilityText
		}
	} else if len(current) >= MaxBundleFiles {
		return httpx.WriteError(event, result.Conflict("asset.limit_reached", "A ruleset can contain at most 100 media items."))
	}
	files, err := event.FindUploadedFiles("file")
	if err != nil || len(files) != 1 {
		return httpx.WriteError(event, result.Invalid("asset.file_required", "Choose one image or audio file.", nil))
	}
	uploaded := files[0]
	if displayName == "" {
		displayName = path.Base(uploaded.Name)
	}
	if len([]rune(displayName)) > 160 || displayName == "" {
		return httpx.WriteError(event, result.Invalid("asset.invalid_name", "Enter a media name up to 160 characters.", nil))
	}
	if len([]rune(accessibilityText)) > 1000 {
		return httpx.WriteError(event, result.Invalid("asset.invalid_accessibility", "Keep the image description or audio alternative under 1000 characters.", nil))
	}
	limit := int64(maxAudioSize)
	if kind == "image" {
		limit = maxImageSize
	}
	if uploaded.Size <= 0 || uploaded.Size > limit {
		return httpx.WriteError(event, result.Invalid("asset.too_large", "The selected media file exceeds its size limit.", nil))
	}
	stream, err := uploaded.Reader.Open()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	content, readErr := io.ReadAll(io.LimitReader(stream, limit+1))
	closeErr := stream.Close()
	if readErr != nil || closeErr != nil || int64(len(content)) > limit {
		return httpx.WriteError(event, result.Invalid("asset.invalid", "The selected media file could not be read safely.", nil))
	}
	mimeType, metadata, err := inspectAsset(kind, content)
	if err != nil {
		return httpx.WriteError(event, result.Invalid("asset.invalid", err.Error(), nil))
	}
	change, err := upsertAssetChange(event.App, session, key)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	operation := mode
	if change.GetString("operation") == "add" {
		operation = "add"
	}
	change.Set("operation", operation)
	change.Set("kind", kind)
	change.Set("display_name", displayName)
	change.Set("accessibility_text", accessibilityText)
	change.Set("mime_type", mimeType)
	digest := sha256.Sum256(content)
	change.Set("checksum", hex.EncodeToString(digest[:]))
	change.Set("metadata", metadata)
	change.Set("file", uploaded)
	if err := event.App.Save(change); err != nil {
		return httpx.WriteError(event, result.Invalid("asset.save_failed", "The media file could not be staged.", nil))
	}
	touchEditSession(session, time.Now().UTC())
	_ = event.App.Save(session)
	asset := effectiveAsset{key: key, kind: kind, displayName: displayName, accessibilityText: accessibilityText, mimeType: mimeType, checksum: change.GetString("checksum"), metadata: metadata, record: change, staged: true}
	return event.JSON(http.StatusCreated, projectEffectiveAsset(asset, nil))
}

type assetMetadataRequest struct {
	DisplayName       string `json:"displayName"`
	AccessibilityText string `json:"accessibilityText"`
}

func updateSessionAsset(event *core.RequestEvent) error {
	session, err := ownedEditSession(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	assets, err := effectiveSessionAssets(event.App, session)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	key := event.Request.PathValue("assetKey")
	asset, ok := assets[key]
	if !ok {
		return httpx.WriteError(event, result.AppError{Code: "asset.not_found", Message: "Media item not found.", Status: http.StatusNotFound})
	}
	var request assetMetadataRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("asset.invalid", "The media details could not be read.", nil))
	}
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	request.AccessibilityText = strings.TrimSpace(request.AccessibilityText)
	if request.DisplayName == "" || len([]rune(request.DisplayName)) > 160 || len([]rune(request.AccessibilityText)) > 1000 {
		return httpx.WriteError(event, result.Invalid("asset.invalid_metadata", "Enter a media name and keep accessibility text under 1000 characters.", nil))
	}
	change, err := upsertAssetChange(event.App, session, key)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if change.GetString("operation") == "" {
		change.Set("operation", "update")
	}
	change.Set("kind", asset.kind)
	change.Set("display_name", request.DisplayName)
	change.Set("accessibility_text", request.AccessibilityText)
	change.Set("mime_type", asset.mimeType)
	change.Set("checksum", asset.checksum)
	change.Set("metadata", asset.metadata)
	if err := event.App.Save(change); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	touchEditSession(session, time.Now().UTC())
	_ = event.App.Save(session)
	asset.displayName, asset.accessibilityText, asset.record, asset.staged = request.DisplayName, request.AccessibilityText, change, true
	return event.JSON(http.StatusOK, projectEffectiveAsset(asset, nil))
}

type deleteAssetRequest struct {
	Definition DefinitionV1 `json:"definition"`
}

func deleteSessionAsset(event *core.RequestEvent) error {
	session, err := ownedEditSession(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	assets, err := effectiveSessionAssets(event.App, session)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	key := event.Request.PathValue("assetKey")
	asset, ok := assets[key]
	if !ok {
		return httpx.WriteError(event, result.AppError{Code: "asset.not_found", Message: "Media item not found.", Status: http.StatusNotFound})
	}
	var request deleteAssetRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("asset.invalid", "The ruleset definition could not be read.", nil))
	}
	if usages := scanAssetUsages(request.Definition, key); len(usages) > 0 {
		return event.JSON(http.StatusConflict, map[string]any{"code": "asset.in_use", "message": "Remove this media item from its usages before deleting it.", "usages": usages})
	}
	change, err := upsertAssetChange(event.App, session, key)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if change.GetString("operation") == "add" {
		if err := event.App.Delete(change); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	} else {
		change.Set("operation", "delete")
		change.Set("kind", asset.kind)
		change.Set("display_name", asset.displayName)
		change.Set("accessibility_text", asset.accessibilityText)
		change.Set("mime_type", asset.mimeType)
		change.Set("checksum", asset.checksum)
		change.Set("metadata", asset.metadata)
		change.Set("file", nil)
		if err := event.App.Save(change); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	touchEditSession(session, time.Now().UTC())
	_ = event.App.Save(session)
	return event.NoContent(http.StatusNoContent)
}

func previewSessionAsset(event *core.RequestEvent) error {
	if event.Auth == nil || event.Auth.Collection().Name != "game_masters" || !event.Auth.GetBool("active") {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "Sign in to preview this media item.", Status: http.StatusUnauthorized})
	}
	change, err := event.App.FindRecordById("ruleset_asset_changes", event.Request.PathValue("id"))
	if err != nil || change.GetString("file") == "" {
		return httpx.WriteError(event, result.AppError{Code: "asset.not_found", Message: "Media preview not found.", Status: http.StatusNotFound})
	}
	session, err := event.App.FindRecordById("ruleset_edit_sessions", change.GetString("session"))
	if err != nil || session.GetString("creator") != event.Auth.Id || !session.GetDateTime("expires_at").Time().After(time.Now().UTC()) {
		return httpx.WriteError(event, result.Forbidden("asset.forbidden", "This media preview is not available."))
	}
	return serveRecordFile(event, change, maxAudioSize)
}

func ownedEditSession(event *core.RequestEvent) (*core.Record, error) {
	session, err := event.App.FindRecordById("ruleset_edit_sessions", event.Request.PathValue("sessionId"))
	if err != nil || session.GetString("ruleset") != event.Request.PathValue("id") {
		return nil, result.AppError{Code: "edit_session.not_found", Message: "Editing session not found. Reopen the ruleset and try again.", Status: http.StatusNotFound}
	}
	if session.GetString("creator") != event.Auth.Id {
		return nil, result.Forbidden("edit_session.forbidden", "This editing session belongs to another account.")
	}
	if !session.GetDateTime("expires_at").Time().After(time.Now().UTC()) {
		_ = deleteEditSession(event.App, session)
		return nil, result.AppError{Code: "edit_session.expired", Message: "The media editing session expired. Reopen the ruleset and upload staged files again.", Status: http.StatusGone}
	}
	return session, nil
}

func touchEditSession(session *core.Record, now time.Time) {
	session.Set("activity_at", now)
	session.Set("expires_at", now.Add(editSessionLifetime))
}

func projectEditSession(session *core.Record) map[string]any {
	return map[string]any{"id": session.Id, "expiresAt": session.GetDateTime("expires_at").Time().UTC()}
}

func upsertAssetChange(app core.App, session *core.Record, key string) (*core.Record, error) {
	items, err := app.FindRecordsByFilter("ruleset_asset_changes", "session = {:session} && asset_key = {:key}", "", 1, 0, dbx.Params{"session": session.Id, "key": key})
	if err != nil {
		return nil, err
	}
	if len(items) == 1 {
		return items[0], nil
	}
	collection, err := app.FindCollectionByNameOrId("ruleset_asset_changes")
	if err != nil {
		return nil, err
	}
	record := core.NewRecord(collection)
	record.Set("session", session.Id)
	record.Set("asset_key", key)
	return record, nil
}

func effectiveSessionAssets(app core.App, session *core.Record) (map[string]effectiveAsset, error) {
	records, err := app.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": session.GetString("base_version")})
	if err != nil {
		return nil, err
	}
	assets := make(map[string]effectiveAsset, len(records))
	for _, record := range records {
		if record.GetString("file") == "" {
			continue
		}
		assets[record.GetString("asset_key")] = effectiveAssetFromRecord(record, false)
	}
	changes, err := app.FindAllRecords("ruleset_asset_changes", dbx.HashExp{"session": session.Id})
	if err != nil {
		return nil, err
	}
	for _, change := range changes {
		key := change.GetString("asset_key")
		switch change.GetString("operation") {
		case "delete":
			delete(assets, key)
		case "add", "replace":
			assets[key] = effectiveAssetFromRecord(change, true)
		case "update":
			asset, ok := assets[key]
			if ok {
				asset.displayName = change.GetString("display_name")
				asset.accessibilityText = change.GetString("accessibility_text")
				assets[key] = asset
			}
		}
	}
	return assets, nil
}

func effectiveAssetFromRecord(record *core.Record, staged bool) effectiveAsset {
	name := strings.TrimSpace(record.GetString("display_name"))
	if name == "" {
		name = path.Base(record.GetString("file"))
	}
	return effectiveAsset{key: record.GetString("asset_key"), kind: record.GetString("kind"), displayName: name, accessibilityText: record.GetString("accessibility_text"), mimeType: record.GetString("mime_type"), checksum: record.GetString("checksum"), metadata: record.Get("metadata"), record: record, staged: staged}
}

func projectEffectiveAsset(asset effectiveAsset, usages []assetUsage) map[string]any {
	preview := "/api/app/v1/ruleset-assets/" + asset.record.Id
	if asset.staged && asset.record.GetString("file") != "" {
		preview = "/api/app/v1/ruleset-edit-assets/" + asset.record.Id
	}
	return map[string]any{"assetKey": asset.key, "displayName": asset.displayName, "accessibilityText": asset.accessibilityText, "kind": asset.kind, "mimeType": asset.mimeType, "checksum": asset.checksum, "metadata": asset.metadata, "preview": preview, "staged": asset.staged, "usages": usages}
}

func prepareEffectiveSessionAssets(app core.App, session *core.Record) ([]preparedVersionAsset, error) {
	assets, err := effectiveSessionAssets(app, session)
	if err != nil {
		return nil, err
	}
	prepared := make([]preparedVersionAsset, 0, len(assets))
	fsys, err := app.NewFilesystem()
	if err != nil {
		return nil, err
	}
	defer fsys.Close()
	for _, asset := range assets {
		filename := asset.record.GetString("file")
		reader, err := fsys.GetReader(asset.record.BaseFilesPath() + "/" + filename)
		if err != nil {
			return nil, err
		}
		content, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			if readErr != nil {
				return nil, readErr
			}
			return nil, closeErr
		}
		file, err := filesystem.NewFileFromBytes(content, path.Base(filename))
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, preparedVersionAsset{key: asset.key, kind: asset.kind, displayName: asset.displayName, accessibilityText: asset.accessibilityText, mimeType: asset.mimeType, checksum: asset.checksum, metadata: asset.metadata, file: file})
	}
	return prepared, nil
}

func saveRulesetEditSession(event *core.RequestEvent, logical *core.Record, request saveRulesetRequest) error {
	event.Request.SetPathValue("sessionId", request.SessionID)
	session, err := ownedEditSession(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	prepared, err := prepareEffectiveSessionAssets(event.App, session)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	assetKeys := make(map[string]struct{}, len(prepared))
	request.Definition.AssetAccessibility = map[string]AssetAccessibility{}
	for _, asset := range prepared {
		assetKeys[asset.key] = struct{}{}
		if asset.accessibilityText != "" {
			request.Definition.AssetAccessibility[asset.key] = AssetAccessibility{Description: asset.accessibilityText}
		}
	}
	report := Validate(request.Definition, assetKeys)
	canonical, err := json.Marshal(request.Definition)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}

	var saved *core.Record
	if err := event.App.RunInTransaction(func(tx core.App) error {
		versions, err := tx.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset}", "-version_number", 1, 0, dbx.Params{"ruleset": logical.Id})
		if err != nil {
			return err
		}
		number := 1
		if len(versions) > 0 {
			number = versions[0].GetInt("version_number") + 1
		}
		collection, err := tx.FindCollectionByNameOrId("ruleset_versions")
		if err != nil {
			return err
		}
		saved = core.NewRecord(collection)
		saved.Set("ruleset", logical.Id)
		saved.Set("version_number", number)
		// Validity is represented by the logical usable pointer. Internal
		// revisions are immutable after every explicit Save.
		saved.Set("state", "published")
		saved.Set("schema_version", 1)
		saved.Set("definition", request.Definition)
		saved.Set("definition_checksum", checksum(canonical))
		saved.Set("validation_report", report)
		saved.Set("created_by", event.Auth.Id)
		saved.Set("published_by", event.Auth.Id)
		saved.Set("published_at", time.Now().UTC())
		if err := tx.Save(saved); err != nil {
			return err
		}
		return stagePreparedVersionAssets(tx, saved.Id, prepared)
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := uploadStagedVersionAssets(event.App, saved.Id, prepared); err != nil {
		_ = deleteStagedDraft(event.App, saved.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	availability := "invalid"
	action := "ruleset.saved_invalid"
	if report.Valid() {
		availability = "ready"
		action = "ruleset.saved_ready"
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := markVersionAssetsReady(tx, saved.Id); err != nil {
			return err
		}
		current, err := tx.FindRecordById("rulesets", logical.Id)
		if err != nil {
			return err
		}
		current.Set("name", request.Definition.Metadata.Name)
		current.Set("latest_saved_version", saved.Id)
		if report.Valid() {
			current.Set("latest_published_version", saved.Id)
		} else {
			current.Set("latest_published_version", "")
		}
		if err := tx.Save(current); err != nil {
			return err
		}
		if err := auditRecord(tx, event.Auth, "", action, "ruleset_version", saved.Id, map[string]any{"rulesetId": logical.Id, "availability": availability}, event.Get(httpx.TraceIDKey)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		_ = deleteStagedDraft(event.App, saved.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := deleteEditSession(event.App, session); err != nil {
		event.App.Logger().Error("consumed ruleset edit session cleanup failed", "sessionId", session.Id, "error", err)
	}
	return event.JSON(http.StatusOK, map[string]any{"validation": report, "availability": availability})
}

func deleteEditSession(app core.App, session *core.Record) error {
	changes, err := app.FindAllRecords("ruleset_asset_changes", dbx.HashExp{"session": session.Id})
	if err != nil {
		return err
	}
	for _, change := range changes {
		if err := app.Delete(change); err != nil {
			return err
		}
	}
	return app.Delete(session)
}

func cleanupExpiredEditSessions(app core.App, now time.Time) error {
	items, err := app.FindRecordsByFilter("ruleset_edit_sessions", "expires_at <= {:now}", "", 200, 0, dbx.Params{"now": now})
	if err != nil {
		return err
	}
	for _, session := range items {
		if err := deleteEditSession(app, session); err != nil {
			return err
		}
	}
	return nil
}

func serveRecordFile(event *core.RequestEvent, record *core.Record, limit int64) error {
	fsys, err := event.App.NewFilesystem()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer fsys.Close()
	reader, err := fsys.GetReader(record.BaseFilesPath() + "/" + record.GetString("file"))
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer reader.Close()
	content, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil || int64(len(content)) > limit {
		return httpx.WriteError(event, result.Internal(errors.New("media storage content exceeds limit")))
	}
	event.Response.Header().Set("Cache-Control", "private, no-store")
	event.Response.Header().Set("Content-Disposition", `inline; filename="`+path.Base(record.GetString("file"))+`"`)
	return event.Blob(http.StatusOK, record.GetString("mime_type"), content)
}

func sessionDefinition(event *core.RequestEvent) (DefinitionV1, error) {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return DefinitionV1{}, err
	}
	version, err := latestSavedVersion(event.App, logical)
	if err != nil {
		return DefinitionV1{}, err
	}
	return definitionFromRecord(version)
}

func scanAssetUsages(definition DefinitionV1, key string) []assetUsage {
	usages := []assetUsage{}
	if definition.Metadata.CoverAssetKey == key {
		usages = append(usages, assetUsage{Label: "Ruleset cover", Section: "metadata"})
	}
	for _, team := range definition.Teams {
		if team.ImageAssetKey == key {
			usages = append(usages, assetUsage{Label: "Team · " + team.Name, Section: "teams", ItemID: team.ID})
		}
	}
	for _, role := range definition.Roles {
		if role.ImageAssetKey == key {
			usages = append(usages, assetUsage{Label: "Role · " + role.Name, Section: "roles", ItemID: role.ID})
		}
	}
	for _, ability := range definition.Abilities {
		if ability.ImageAssetKey == key {
			usages = append(usages, assetUsage{Label: "Ability · " + ability.Name, Section: "roles", ItemID: ability.ID})
		}
	}
	for _, achievement := range definition.Achievements {
		if achievement.ImageAssetKey == key {
			usages = append(usages, assetUsage{Label: "Achievement · " + achievement.Name, Section: "achievements", ItemID: achievement.ID})
		}
	}
	for _, cue := range definition.AudioCues {
		if cue.AssetKey != key {
			continue
		}
		used := false
		for _, phase := range definition.Phases {
			if phase.AudioCueID == cue.ID {
				used = true
				usages = append(usages, assetUsage{Label: fmt.Sprintf("Audio cue · %s → Phase · %s", cue.Name, phase.Name), Section: "phases", ItemID: phase.ID})
			}
		}
		if !used {
			usages = append(usages, assetUsage{Label: "Audio cue · " + cue.Name, Section: "audio", ItemID: cue.ID})
		}
	}
	return usages
}
