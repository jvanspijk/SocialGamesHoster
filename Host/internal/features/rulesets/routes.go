package rulesets

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

// auditRecord is replaced only by package tests to verify that the final
// ready-state transaction rolls back when durable auditing fails.
var auditRecord = applicationaudit.Record

type createRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	MinPlayers      int    `json:"minPlayers"`
	MaxPlayers      int    `json:"maxPlayers"`
	SourceRulesetID string `json:"sourceRulesetId"`
}

type updateVersionRequest struct {
	Definition DefinitionV1 `json:"definition"`
}

type preparedVersionAsset struct {
	key, kind, displayName, accessibilityText, mimeType, checksum string
	metadata                                                      any
	file                                                          *filesystem.File
}

func RegisterRoutes(event *core.ServeEvent, applicationVersion string) {
	if err := cleanupExpiredEditSessions(event.App, time.Now().UTC()); err != nil {
		event.App.Logger().Error("expired ruleset edit session cleanup failed", "error", err)
	}
	group := event.Router.Group("/api/app/v1")
	group.BindFunc(actorauth.RequireGameMaster)
	group.GET("/rulesets", listRulesets)
	group.POST("/rulesets", createRuleset)
	group.GET("/rulesets/{id}", getRuleset)
	group.DELETE("/rulesets/{id}", deleteRuleset)
	group.POST("/rulesets/{id}/validate", validateRuleset)
	group.POST("/rulesets/{id}/save", saveRuleset)
	group.POST("/rulesets/import", func(event *core.RequestEvent) error {
		return importBundle(event, applicationVersion)
	})
	group.GET("/rulesets/{id}/export", func(event *core.RequestEvent) error {
		return exportLatestSavedBundle(event, applicationVersion)
	})
	registerEditSessionRoutes(group)
	registerAssetPreviewRoute(event)
}

func listRulesets(event *core.RequestEvent) error {
	records, err := event.App.FindRecordsByFilter("rulesets", "", "archived,name", 200, 0)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	response := make([]map[string]any, len(records))
	for i, record := range records {
		if record.GetBool("archived") {
			continue
		}
		saved, findErr := latestSavedVersion(event.App, record)
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		response[i] = projectRulesetWithStatus(record, validationReportFromRecord(saved))
	}
	visible := response[:0]
	for _, item := range response {
		if item != nil {
			visible = append(visible, item)
		}
	}
	return event.JSON(http.StatusOK, visible)
}

func createRuleset(event *core.RequestEvent) error {
	var request createRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid", "The ruleset could not be read.", nil))
	}
	request.Name = strings.TrimSpace(request.Name)
	if len([]rune(request.Name)) < 1 || len([]rune(request.Name)) > 120 {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid_name", "Enter a ruleset name between 1 and 120 characters.", nil))
	}
	if request.MinPlayers < 1 || request.MaxPlayers > 30 || request.MinPlayers > request.MaxPlayers {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid_players", "Choose a player range between 1 and 30.", nil))
	}
	definition := blankDefinition(request.Name, request.Description, request.MinPlayers, request.MaxPlayers)
	preparedAssets := []preparedVersionAsset(nil)
	if request.SourceRulesetID != "" {
		source, findErr := event.App.FindRecordById("rulesets", request.SourceRulesetID)
		if findErr != nil || source.GetBool("archived") {
			return httpx.WriteError(event, result.Invalid("ruleset.invalid_source", "Choose an existing ruleset to duplicate.", nil))
		}
		sourceVersion, findErr := latestSavedVersion(event.App, source)
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		definition, findErr = definitionFromRecord(sourceVersion)
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		definition.Metadata.Name = request.Name
		definition.Metadata.Description = request.Description
		definition.Metadata.MinPlayers = request.MinPlayers
		definition.Metadata.MaxPlayers = request.MaxPlayers
		preparedAssets, findErr = prepareVersionAssets(event.App, sourceVersion.Id)
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
	}
	slug, err := generatedSlug(event.App, request.Name)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	var rulesetRecord *core.Record
	var versionRecord *core.Record
	assetKeys := make(map[string]struct{}, len(preparedAssets))
	for _, asset := range preparedAssets {
		assetKeys[asset.key] = struct{}{}
	}
	report := Validate(definition, assetKeys)
	err = event.App.RunInTransaction(func(txApp core.App) error {
		var err error
		rulesetRecord, versionRecord, err = createDraftRecords(txApp, slug, definition, event.Auth.Id, nil)
		if err != nil {
			return err
		}
		versionRecord.Set("validation_report", report)
		if err := txApp.Save(versionRecord); err != nil {
			return err
		}
		if err := stagePreparedVersionAssets(txApp, versionRecord.Id, preparedAssets); err != nil {
			return err
		}
		rulesetRecord.Set("latest_saved_version", versionRecord.Id)
		if err := txApp.Save(rulesetRecord); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		// Duplicate slugs are the only expected create failure; audit/storage
		// failures retain their standard internal-error contract.
		if _, exists := err.(result.AppError); exists {
			return httpx.WriteErrorFrom(event, err)
		}
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := uploadStagedVersionAssets(event.App, versionRecord.Id, preparedAssets); err != nil {
		_ = deleteStagedDraft(event.App, versionRecord.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := markVersionAssetsReady(tx, versionRecord.Id); err != nil {
			return err
		}
		current, err := tx.FindRecordById("rulesets", rulesetRecord.Id)
		if err != nil {
			return err
		}
		if report.Valid() {
			current.Set("latest_published_version", versionRecord.Id)
		}
		if err := tx.Save(current); err != nil {
			return err
		}
		return auditRecord(tx, event.Auth, "", "ruleset.created", "ruleset", current.Id, nil, event.Get(httpx.TraceIDKey))
	}); err != nil {
		_ = deleteStagedDraft(event.App, versionRecord.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, projectRulesetWithStatus(rulesetRecord, report))
}

func getRuleset(event *core.RequestEvent) error {
	record, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil || record.GetBool("archived") {
		return rulesetNotFound(event)
	}
	saved, err := latestSavedVersion(event.App, record)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	definition, err := definitionFromRecord(saved)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	report := validationReportFromRecord(saved)
	return event.JSON(http.StatusOK, map[string]any{
		"ruleset":    projectRulesetWithStatus(record, report),
		"definition": definition,
		"validation": report,
	})
}

func updateVersion(event *core.RequestEvent) error {
	record, err := event.App.FindRecordById("ruleset_versions", event.Request.PathValue("id"))
	if err != nil {
		return versionNotFound(event)
	}
	if record.GetString("state") != "draft" {
		return httpx.WriteError(event, result.Conflict("ruleset.published_immutable", "Published ruleset versions cannot be changed."))
	}
	var request updateVersionRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid", "The ruleset definition could not be read.", nil))
	}
	record.Set("definition", request.Definition)
	record.Set("definition_checksum", "")
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.save_failed", "The draft could not be saved.", nil))
	}
	return event.JSON(http.StatusOK, projectVersion(record))
}

type saveRulesetRequest struct {
	Definition DefinitionV1 `json:"definition"`
	SessionID  string       `json:"sessionId,omitempty"`
}

// validateRuleset evaluates an editor working copy without changing the saved
// ruleset. Assets are resolved from the latest saved revision until edit
// sessions provide a staged effective asset set in Stage 3.
func validateRuleset(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil || logical.GetBool("archived") {
		return rulesetNotFound(event)
	}
	var request saveRulesetRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid", "The ruleset definition could not be read.", nil))
	}
	assetKeys := map[string]struct{}{}
	if request.SessionID != "" {
		event.Request.SetPathValue("sessionId", request.SessionID)
		session, sessionErr := ownedEditSession(event)
		if sessionErr != nil {
			return httpx.WriteErrorFrom(event, sessionErr)
		}
		assets, sessionErr := effectiveSessionAssets(event.App, session)
		if sessionErr != nil {
			return httpx.WriteError(event, result.Internal(sessionErr))
		}
		for key := range assets {
			assetKeys[key] = struct{}{}
		}
		request.Definition.AssetAccessibility = map[string]AssetAccessibility{}
		for key, asset := range assets {
			if asset.accessibilityText != "" {
				request.Definition.AssetAccessibility[key] = AssetAccessibility{Description: asset.accessibilityText}
			}
		}
		touchEditSession(session, time.Now().UTC())
		if err := event.App.Save(session); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	} else {
		saved, findErr := latestSavedVersion(event.App, logical)
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		assetKeys, err = versionAssetKeys(event.App, saved.Id)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	return event.JSON(http.StatusOK, Validate(request.Definition, assetKeys))
}

func saveRuleset(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil || logical.GetBool("archived") {
		return rulesetNotFound(event)
	}
	var request saveRulesetRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid", "The ruleset definition could not be read.", nil))
	}
	if request.SessionID != "" {
		return saveRulesetEditSession(event, logical, request)
	}
	// File reads are deliberately outside the short SQLite transaction. The
	// prepared files are in memory; only their record saves occur in the tx.
	assetSourceID := logical.GetString("latest_published_version")
	if assetSourceID == "" {
		published, findErr := event.App.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset} && state = 'published'", "-version_number", 1, 0, dbx.Params{"ruleset": logical.Id})
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		if len(published) > 0 {
			assetSourceID = published[0].Id
		}
	}
	preparedAssets, err := prepareVersionAssets(event.App, assetSourceID)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	var saved *core.Record
	var report ValidationReport
	createdSuccessor := false
	usedExistingDraft := false
	needsAssetFinalization := false
	previousPublishedID := logical.GetString("latest_published_version")
	err = event.App.RunInTransaction(func(tx core.App) error {
		currentLogical, err := tx.FindRecordById("rulesets", logical.Id)
		if err != nil {
			return err
		}
		drafts, err := tx.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset} && state = 'draft'", "", 1, 0, dbx.Params{"ruleset": logical.Id})
		if err != nil {
			return err
		}
		if len(drafts) > 0 {
			saved = drafts[0]
			usedExistingDraft = true
		} else {
			createdSuccessor = true
			sourceID := currentLogical.GetString("latest_published_version")
			source, err := tx.FindRecordById("ruleset_versions", sourceID)
			if err != nil {
				return err
			}
			collection, err := tx.FindCollectionByNameOrId("ruleset_versions")
			if err != nil {
				return err
			}
			saved = core.NewRecord(collection)
			saved.Set("ruleset", logical.Id)
			saved.Set("version_number", source.GetInt("version_number")+1)
			saved.Set("state", "draft")
			saved.Set("schema_version", 1)
			saved.Set("created_by", event.Auth.Id)
			saved.Set("definition", request.Definition)
			saved.Set("definition_checksum", "")
			if err := tx.Save(saved); err != nil {
				return err
			}
			if err := stagePreparedVersionAssets(tx, saved.Id, preparedAssets); err != nil {
				return err
			}
		}
		staged, err := tx.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && storage_state = 'staging'", "", MaxBundleFiles, 0, dbx.Params{"version": saved.Id})
		if err != nil {
			return err
		}
		needsAssetFinalization = len(staged) > 0
		saved.Set("definition", request.Definition)
		saved.Set("definition_checksum", "")
		if err := tx.Save(saved); err != nil {
			return err
		}
		assetKeys, err := versionAssetKeys(tx, saved.Id)
		if err != nil {
			return err
		}
		// Successor files remain staged until their post-commit upload succeeds,
		// but their already validated keys still participate in draft validation.
		if createdSuccessor || needsAssetFinalization {
			for _, asset := range preparedAssets {
				assetKeys[asset.key] = struct{}{}
			}
		}
		report = Validate(request.Definition, assetKeys)
		saved.Set("validation_report", report)
		currentLogical.Set("name", request.Definition.Metadata.Name)
		currentLogical.Set("latest_saved_version", saved.Id)
		if !report.Valid() {
			currentLogical.Set("latest_published_version", "")
			if err := tx.Save(saved); err != nil {
				return err
			}
			if err := tx.Save(currentLogical); err != nil {
				return err
			}
			if usedExistingDraft && !needsAssetFinalization {
				return auditRecord(tx, event.Auth, "", "ruleset.saved_invalid", "ruleset_version", saved.Id, map[string]any{"rulesetId": logical.Id, "availability": "invalid"}, event.Get(httpx.TraceIDKey))
			}
			return nil
		}
		canonical, err := json.Marshal(request.Definition)
		if err != nil {
			return err
		}
		saved.Set("definition_checksum", checksum(canonical))
		if !createdSuccessor && !needsAssetFinalization {
			saved.Set("state", "published")
			saved.Set("published_by", event.Auth.Id)
			saved.Set("published_at", time.Now().UTC())
		}
		if err := tx.Save(saved); err != nil {
			return err
		}
		if !createdSuccessor && !needsAssetFinalization {
			currentLogical.Set("latest_published_version", saved.Id)
		}
		if err := tx.Save(currentLogical); err != nil {
			return err
		}
		if usedExistingDraft && !needsAssetFinalization {
			return auditRecord(tx, event.Auth, "", "ruleset.saved_ready", "ruleset_version", saved.Id, map[string]any{"rulesetId": logical.Id, "availability": "ready"}, event.Get(httpx.TraceIDKey))
		}
		return nil
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	availability := "invalid"
	if report.Valid() {
		availability = "ready"
	}
	if (createdSuccessor || needsAssetFinalization) && report.Valid() {
		if err := uploadStagedVersionAssets(event.App, saved.Id, preparedAssets); err != nil {
			_ = discardStagedSuccessor(event.App, logical.Id, saved.Id, previousPublishedID)
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	action := "ruleset.saved_invalid"
	if report.Valid() {
		action = "ruleset.saved_ready"
	}
	if !usedExistingDraft || needsAssetFinalization {
		if err := event.App.RunInTransaction(func(tx core.App) error {
			if (createdSuccessor || needsAssetFinalization) && report.Valid() {
				if err := markVersionAssetsReady(tx, saved.Id); err != nil {
					return err
				}
				current, err := tx.FindRecordById("ruleset_versions", saved.Id)
				if err != nil {
					return err
				}
				current.Set("state", "published")
				current.Set("published_by", event.Auth.Id)
				current.Set("published_at", time.Now().UTC())
				if err := tx.Save(current); err != nil {
					return err
				}
				saved = current
				logicalCurrent, err := tx.FindRecordById("rulesets", logical.Id)
				if err != nil {
					return err
				}
				logicalCurrent.Set("latest_published_version", saved.Id)
				if err := tx.Save(logicalCurrent); err != nil {
					return err
				}
			}
			return auditRecord(tx, event.Auth, "", action, "ruleset_version", saved.Id, map[string]any{"rulesetId": logical.Id, "availability": availability}, event.Get(httpx.TraceIDKey))
		}); err != nil {
			if createdSuccessor {
				_ = discardStagedSuccessor(event.App, logical.Id, saved.Id, previousPublishedID)
			}
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	return event.JSON(http.StatusOK, map[string]any{
		"version":      projectVersion(saved),
		"validation":   report,
		"availability": availability,
	})
}

func validateVersion(event *core.RequestEvent) error {
	record, err := event.App.FindRecordById("ruleset_versions", event.Request.PathValue("id"))
	if err != nil {
		return versionNotFound(event)
	}
	definition, err := definitionFromRecord(record)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	assetKeys, err := versionAssetKeys(event.App, record.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, Validate(definition, assetKeys))
}

func publishVersion(event *core.RequestEvent) error {
	versionID := event.Request.PathValue("id")
	var published *core.Record
	err := event.App.RunInTransaction(func(txApp core.App) error {
		record, err := txApp.FindRecordById("ruleset_versions", versionID)
		if err != nil {
			return result.AppError{Code: "ruleset_version.not_found", Message: "Ruleset version not found.", Status: http.StatusNotFound}
		}
		if record.GetString("state") != "draft" {
			return result.AppError{Code: "ruleset.published_immutable", Message: "This version is already published.", Status: http.StatusConflict}
		}
		definition, err := definitionFromRecord(record)
		if err != nil {
			return err
		}
		assetKeys, err := versionAssetKeys(txApp, record.Id)
		if err != nil {
			return err
		}
		report := Validate(definition, assetKeys)
		if !report.Valid() {
			return validationAppError(report)
		}
		canonical, err := json.Marshal(definition)
		if err != nil {
			return err
		}
		record.Set("state", "published")
		record.Set("definition_checksum", checksum(canonical))
		record.Set("published_by", event.Auth.Id)
		record.Set("published_at", time.Now().UTC())
		if err := txApp.Save(record); err != nil {
			return err
		}
		logical, err := txApp.FindRecordById("rulesets", record.GetString("ruleset"))
		if err != nil {
			return err
		}
		logical.Set("name", definition.Metadata.Name)
		logical.Set("latest_published_version", record.Id)
		if err := txApp.Save(logical); err != nil {
			return err
		}
		published = record
		return auditRecord(txApp, event.Auth, "", "ruleset.published", "ruleset_version", published.Id, map[string]any{"rulesetId": published.GetString("ruleset")}, event.Get(httpx.TraceIDKey))
	})
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	return event.JSON(http.StatusOK, projectVersion(published))
}

func createDraft(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return rulesetNotFound(event)
	}
	existing, err := event.App.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset} && state = 'draft'", "", 1, 0, dbx.Params{"ruleset": logical.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if len(existing) > 0 {
		return event.JSON(http.StatusOK, projectVersion(existing[0]))
	}
	sourceID := logical.GetString("latest_published_version")
	if sourceID == "" {
		return httpx.WriteError(event, result.Conflict("ruleset.no_published_version", "Publish the first draft before creating a successor."))
	}
	source, err := event.App.FindRecordById("ruleset_versions", sourceID)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	definition, err := definitionFromRecord(source)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	preparedAssets, err := prepareVersionAssets(event.App, source.Id)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	var draft *core.Record
	created := false
	if err := event.App.RunInTransaction(func(tx core.App) error {
		current, err := tx.FindRecordById("rulesets", logical.Id)
		if err != nil {
			return err
		}
		existing, err := tx.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset} && state = 'draft'", "", 1, 0, dbx.Params{"ruleset": current.Id})
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			draft = existing[0]
			return nil
		}
		collection, err := tx.FindCollectionByNameOrId("ruleset_versions")
		if err != nil {
			return err
		}
		draft = core.NewRecord(collection)
		draft.Set("ruleset", current.Id)
		draft.Set("version_number", source.GetInt("version_number")+1)
		draft.Set("state", "draft")
		draft.Set("schema_version", 1)
		draft.Set("definition", definition)
		draft.Set("created_by", event.Auth.Id)
		if err := tx.Save(draft); err != nil {
			return err
		}
		created = true
		if err := stagePreparedVersionAssets(tx, draft.Id, preparedAssets); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if !created {
		return event.JSON(http.StatusOK, projectVersion(draft))
	}
	if err := uploadStagedVersionAssets(event.App, draft.Id, preparedAssets); err != nil {
		_ = deleteStagedDraft(event.App, draft.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := markVersionAssetsReady(tx, draft.Id); err != nil {
			return err
		}
		return auditRecord(tx, event.Auth, "", "ruleset.draft_created", "ruleset_version", draft.Id, map[string]any{"rulesetId": logical.Id}, event.Get(httpx.TraceIDKey))
	}); err != nil {
		_ = deleteStagedDraft(event.App, draft.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, projectVersion(draft))
}

func prepareVersionAssets(app core.App, sourceVersionID string) ([]preparedVersionAsset, error) {
	if sourceVersionID == "" {
		return nil, nil
	}
	sourceAssets, err := app.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": sourceVersionID, "storage_state": "ready"})
	if err != nil || len(sourceAssets) == 0 {
		return nil, err
	}
	fsys, err := app.NewFilesystem()
	if err != nil {
		return nil, err
	}
	defer fsys.Close()

	prepared := make([]preparedVersionAsset, 0, len(sourceAssets))
	for _, source := range sourceAssets {
		filename := source.GetString("file")
		reader, err := fsys.GetReader(source.BaseFilesPath() + "/" + filename)
		if err != nil {
			return nil, err
		}
		content, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		clonedFile, err := filesystem.NewFileFromBytes(content, path.Base(filename))
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, preparedVersionAsset{key: source.GetString("asset_key"), kind: source.GetString("kind"), displayName: source.GetString("display_name"), accessibilityText: source.GetString("accessibility_text"), file: clonedFile, mimeType: source.GetString("mime_type"), checksum: source.GetString("checksum"), metadata: source.Get("metadata")})
	}
	return prepared, nil
}

func stagePreparedVersionAssets(app core.App, targetVersionID string, prepared []preparedVersionAsset) error {
	if len(prepared) == 0 {
		return nil
	}
	collection, err := app.FindCollectionByNameOrId("ruleset_assets")
	if err != nil {
		return err
	}
	for _, asset := range prepared {
		target := core.NewRecord(collection)
		target.Set("ruleset_version", targetVersionID)
		target.Set("asset_key", asset.key)
		target.Set("kind", asset.kind)
		target.Set("display_name", asset.displayName)
		target.Set("accessibility_text", asset.accessibilityText)
		target.Set("storage_state", "staging")
		target.Set("mime_type", asset.mimeType)
		target.Set("checksum", asset.checksum)
		target.Set("metadata", asset.metadata)
		if err := app.Save(target); err != nil {
			return err
		}
	}
	return nil
}

func uploadStagedVersionAssets(app core.App, versionID string, prepared []preparedVersionAsset) error {
	for _, item := range prepared {
		records, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && asset_key = {:key}", "", 1, 0, dbx.Params{"version": versionID, "key": item.key})
		if err != nil || len(records) != 1 {
			if err != nil {
				return err
			}
			return fmt.Errorf("staged asset missing")
		}
		records[0].Set("file", item.file)
		if err := app.Save(records[0]); err != nil {
			return err
		}
	}
	return nil
}

func markVersionAssetsReady(app core.App, versionID string) error {
	records, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && storage_state = 'staging'", "", MaxBundleFiles, 0, dbx.Params{"version": versionID})
	if err != nil {
		return err
	}
	for _, record := range records {
		record.Set("storage_state", "ready")
		if err := app.Save(record); err != nil {
			return err
		}
	}
	return nil
}

func deleteStagedDraft(app core.App, versionID string) error {
	assets, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version}", "", MaxBundleFiles, 0, dbx.Params{"version": versionID})
	if err != nil {
		return err
	}
	for _, asset := range assets {
		_ = app.Delete(asset)
	}
	version, err := app.FindRecordById("ruleset_versions", versionID)
	if err != nil {
		return err
	}
	return app.Delete(version)
}

func discardStagedSuccessor(app core.App, rulesetID, versionID, previousPublishedID string) error {
	// Compensate the externally uploaded blobs before removing their staging
	// records. Restore the logical pointer so a failed successor is invisible.
	if err := deleteStagedDraft(app, versionID); err != nil {
		return err
	}
	logical, err := app.FindRecordById("rulesets", rulesetID)
	if err != nil {
		return err
	}
	logical.Set("latest_published_version", previousPublishedID)
	return app.Save(logical)
}

func archiveRuleset(event *core.RequestEvent) error {
	record, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return rulesetNotFound(event)
	}
	record.Set("archived", true)
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Save(record); err != nil {
			return err
		}
		return auditRecord(tx, event.Auth, "", "ruleset.archived", "ruleset", record.Id, nil, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, projectRuleset(record))
}

func deleteRuleset(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return rulesetNotFound(event)
	}
	sessions, err := event.App.FindAllRecords("ruleset_edit_sessions", dbx.HashExp{"ruleset": logical.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	versions, err := event.App.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset}", "", 100, 0, dbx.Params{"ruleset": logical.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	for _, version := range versions {
		count, err := event.App.CountRecords("games", dbx.HashExp{"ruleset_version": version.Id})
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		if count > 0 {
			if err := event.App.RunInTransaction(func(tx core.App) error {
				current, err := tx.FindRecordById("rulesets", logical.Id)
				if err != nil {
					return err
				}
				current.Set("archived", true)
				current.Set("latest_published_version", "")
				if err := tx.Save(current); err != nil {
					return err
				}
				return auditRecord(tx, event.Auth, "", "ruleset.deleted", "ruleset", current.Id, nil, event.Get(httpx.TraceIDKey))
			}); err != nil {
				return httpx.WriteError(event, result.Internal(err))
			}
			for _, session := range sessions {
				if err := deleteEditSession(event.App, session); err != nil {
					event.App.Logger().Error("archived ruleset edit session cleanup failed", "sessionId", session.Id, "error", err)
				}
			}
			return event.NoContent(http.StatusNoContent)
		}
	}
	assetsToClean := []*core.Record{}
	for _, version := range versions {
		assets, findErr := event.App.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": version.Id})
		if findErr != nil {
			return httpx.WriteError(event, result.Internal(findErr))
		}
		assetsToClean = append(assetsToClean, assets...)
	}
	// FileField deletes are filesystem work and must complete before the
	// related versions are removed. Assets are hidden first so cleanup cannot
	// expose a broken record.
	if err := event.App.RunInTransaction(func(tx core.App) error {
		for _, asset := range assetsToClean {
			current, err := tx.FindRecordById("ruleset_assets", asset.Id)
			if err != nil {
				return err
			}
			current.Set("storage_state", "staging")
			if err := tx.Save(current); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	for _, asset := range assetsToClean {
		if err := event.App.Delete(asset); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	err = event.App.RunInTransaction(func(txApp core.App) error {
		// Audit first so an unavailable audit sink cannot destroy retryable edit
		// sessions. All following record deletions roll back together.
		if err := auditRecord(txApp, event.Auth, "", "ruleset.deleted", "ruleset", logical.Id, nil, event.Get(httpx.TraceIDKey)); err != nil {
			return err
		}
		for _, session := range sessions {
			if err := deleteEditSession(txApp, session); err != nil {
				return err
			}
		}
		logical, err := txApp.FindRecordById("rulesets", logical.Id)
		if err != nil {
			return err
		}
		logical.Set("latest_published_version", "")
		if err := txApp.Save(logical); err != nil {
			return err
		}
		for _, version := range versions {
			if err := txApp.Delete(version); err != nil {
				return err
			}
		}
		if err := txApp.Delete(logical); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.NoContent(http.StatusNoContent)
}

func duplicateVersion(event *core.RequestEvent) error {
	source, err := event.App.FindRecordById("ruleset_versions", event.Request.PathValue("id"))
	if err != nil {
		return versionNotFound(event)
	}
	definition, err := definitionFromRecord(source)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	slug := importSlug(definition.Metadata.Name, source.Id)
	var logical, draft *core.Record
	err = event.App.RunInTransaction(func(txApp core.App) error {
		var err error
		logical, draft, err = createDraftRecords(txApp, slug, definition, event.Auth.Id, map[string]any{"duplicatedFrom": source.Id})
		if err != nil {
			return err
		}
		return auditRecord(txApp, event.Auth, "", "ruleset.duplicated", "ruleset", logical.Id, map[string]any{"sourceVersionId": source.Id}, event.Get(httpx.TraceIDKey))
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, map[string]any{"ruleset": projectRuleset(logical), "draft": projectVersion(draft)})
}

func importBundle(event *core.RequestEvent, applicationVersion string) error {
	data, err := io.ReadAll(io.LimitReader(event.Request.Body, MaxBundleSize+1))
	if err != nil || len(data) > MaxBundleSize {
		return httpx.WriteError(event, result.Invalid("bundle.too_large", "The ruleset bundle exceeds 25 MB.", nil))
	}
	imported, err := ReadBundle(data)
	if err != nil {
		event.App.Logger().Warn("ruleset bundle validation failed", "error", err)
		return httpx.WriteError(event, result.Invalid("bundle.invalid", "The ruleset bundle could not be imported. Check that it is a supported, complete bundle.", nil))
	}
	definition, err := normalizeDefinitionIdentifiers(imported.Definition)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	slug, err := generatedSlug(event.App, imported.Manifest.Name)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	preparedAssets := make([]preparedVersionAsset, 0, len(imported.Manifest.Assets))
	if definition.AssetAccessibility == nil {
		definition.AssetAccessibility = map[string]AssetAccessibility{}
	}
	for _, manifestAsset := range imported.Manifest.Assets {
		content := imported.Assets[manifestAsset.AssetKey]
		file, err := filesystem.NewFileFromBytes(content, path.Base(manifestAsset.Path))
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		mimeType, metadata, err := validateDeclaredAsset(manifestAsset.Kind, manifestAsset.MIMEType, content)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		preparedAssets = append(preparedAssets, preparedVersionAsset{key: manifestAsset.AssetKey, kind: manifestAsset.Kind, displayName: manifestAsset.DisplayName, accessibilityText: manifestAsset.AccessibilityText, file: file, mimeType: mimeType, checksum: manifestAsset.Checksum, metadata: metadata})
		if manifestAsset.AccessibilityText != "" {
			definition.AssetAccessibility[manifestAsset.AssetKey] = AssetAccessibility{Description: manifestAsset.AccessibilityText}
		}
	}
	assetKeys := make(map[string]struct{}, len(preparedAssets))
	for _, asset := range preparedAssets {
		assetKeys[asset.key] = struct{}{}
	}
	report := Validate(definition, assetKeys)
	var logical, draft *core.Record
	err = event.App.RunInTransaction(func(txApp core.App) error {
		source := map[string]any{
			"sourceApplicationVersion": imported.Manifest.SourceApplicationVersion,
			"logicalSourceRulesetId":   imported.Manifest.LogicalSourceRulesetID,
			"sourceVersionNumber":      imported.Manifest.SourceVersionNumber,
			"importedByVersion":        applicationVersion,
		}
		var err error
		logical, draft, err = createDraftRecords(txApp, slug, definition, event.Auth.Id, source)
		if err != nil {
			return err
		}
		return stagePreparedVersionAssets(txApp, draft.Id, preparedAssets)
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := uploadStagedVersionAssets(event.App, draft.Id, preparedAssets); err != nil {
		_ = deleteStagedDraft(event.App, draft.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := markVersionAssetsReady(tx, draft.Id); err != nil {
			return err
		}
		currentDraft, err := tx.FindRecordById("ruleset_versions", draft.Id)
		if err != nil {
			return err
		}
		currentDraft.Set("validation_report", report)
		if err := tx.Save(currentDraft); err != nil {
			return err
		}
		currentLogical, err := tx.FindRecordById("rulesets", logical.Id)
		if err != nil {
			return err
		}
		currentLogical.Set("latest_saved_version", draft.Id)
		if report.Valid() {
			currentLogical.Set("latest_published_version", draft.Id)
		}
		if err := tx.Save(currentLogical); err != nil {
			return err
		}
		return auditRecord(tx, event.Auth, "", "ruleset.imported", "ruleset", logical.Id, map[string]any{"sourceVersion": imported.Manifest.SourceVersionNumber}, event.Get(httpx.TraceIDKey))
	}); err != nil {
		_ = deleteStagedDraft(event.App, draft.Id)
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusCreated, projectRulesetWithStatus(logical, report))
}

func exportBundle(event *core.RequestEvent, applicationVersion string) error {
	version, err := event.App.FindRecordById("ruleset_versions", event.Request.PathValue("id"))
	if err != nil {
		return versionNotFound(event)
	}
	return exportBundleRecord(event, applicationVersion, version)
}

func exportLatestSavedBundle(event *core.RequestEvent, applicationVersion string) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil || logical.GetBool("archived") {
		return rulesetNotFound(event)
	}
	version, err := latestSavedVersion(event.App, logical)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return exportBundleRecord(event, applicationVersion, version)
}

func exportBundleRecord(event *core.RequestEvent, applicationVersion string, version *core.Record) error {
	definition, err := definitionFromRecord(version)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	logical, err := event.App.FindRecordById("rulesets", version.GetString("ruleset"))
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	assetRecords, err := event.App.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": version.Id, "storage_state": "ready"})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	manifest := BundleManifest{
		SourceApplicationVersion:  applicationVersion,
		MinimumApplicationVersion: applicationVersion,
		LogicalSourceRulesetID:    logical.Id,
		SourceVersionNumber:       version.GetInt("version_number"),
		Name:                      definition.Metadata.Name,
		Description:               definition.Metadata.Description,
		Assets:                    make([]BundleAssetManifest, 0, len(assetRecords)),
	}
	assets := map[string][]byte{}
	fsys, err := event.App.NewFilesystem()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer fsys.Close()
	for _, asset := range assetRecords {
		filename := asset.GetString("file")
		reader, err := fsys.GetReader(asset.BaseFilesPath() + "/" + filename)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		assetPath := "assets/" + filename
		manifest.Assets = append(manifest.Assets, BundleAssetManifest{
			Path:              assetPath,
			AssetKey:          asset.GetString("asset_key"),
			Kind:              asset.GetString("kind"),
			MIMEType:          asset.GetString("mime_type"),
			DisplayName:       asset.GetString("display_name"),
			AccessibilityText: asset.GetString("accessibility_text"),
		})
		assets[asset.GetString("asset_key")] = content
	}
	data, err := WriteBundle(manifest, definition, assets)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	filename := logical.GetString("slug") + "-v" + fmt.Sprint(version.GetInt("version_number")) + ".sghrules"
	event.Response.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return event.Blob(http.StatusOK, "application/vnd.socialgameshoster.ruleset+zip", data)
}

func createDraftRecords(app core.App, slug string, definition DefinitionV1, actorID string, sourceMetadata any) (*core.Record, *core.Record, error) {
	rulesetsCollection, err := app.FindCollectionByNameOrId("rulesets")
	if err != nil {
		return nil, nil, err
	}
	versionsCollection, err := app.FindCollectionByNameOrId("ruleset_versions")
	if err != nil {
		return nil, nil, err
	}
	logical := core.NewRecord(rulesetsCollection)
	logical.Set("slug", slug)
	logical.Set("name", definition.Metadata.Name)
	logical.Set("archived", false)
	logical.Set("created_by", actorID)
	if err := app.Save(logical); err != nil {
		return nil, nil, err
	}
	version := core.NewRecord(versionsCollection)
	version.Set("ruleset", logical.Id)
	version.Set("version_number", 1)
	version.Set("state", "draft")
	version.Set("schema_version", 1)
	version.Set("definition", definition)
	version.Set("created_by", actorID)
	if sourceMetadata != nil {
		version.Set("source_metadata", sourceMetadata)
	}
	if err := app.Save(version); err != nil {
		return nil, nil, err
	}
	return logical, version, nil
}

func definitionFromRecord(record *core.Record) (DefinitionV1, error) {
	var definition DefinitionV1
	err := record.UnmarshalJSONField("definition", &definition)
	return definition, err
}

func versionAssetKeys(app core.App, versionID string) (map[string]struct{}, error) {
	records, err := app.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": versionID, "storage_state": "ready"})
	if err != nil {
		return nil, err
	}
	keys := make(map[string]struct{}, len(records))
	for _, record := range records {
		keys[record.GetString("asset_key")] = struct{}{}
	}
	return keys, nil
}

func projectRuleset(record *core.Record) map[string]any {
	return map[string]any{
		"id":   record.Id,
		"name": record.GetString("name"),
	}
}

func projectRulesetWithStatus(record *core.Record, report ValidationReport) map[string]any {
	status := "valid"
	if !report.Valid() {
		status = "invalid"
	}
	return map[string]any{
		"id":         record.Id,
		"name":       record.GetString("name"),
		"status":     status,
		"issueCount": len(report.Errors),
	}
}

func latestSavedVersion(app core.App, logical *core.Record) (*core.Record, error) {
	if id := logical.GetString("latest_saved_version"); id != "" {
		return app.FindRecordById("ruleset_versions", id)
	}
	versions, err := app.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset}", "-version_number", 1, 0, dbx.Params{"ruleset": logical.Id})
	if err != nil || len(versions) == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("ruleset has no saved definition")
	}
	return versions[0], nil
}

func validationReportFromRecord(record *core.Record) ValidationReport {
	var report ValidationReport
	if err := record.UnmarshalJSONField("validation_report", &report); err == nil && (len(report.Errors) > 0 || len(report.Warnings) > 0) {
		return report
	}
	return ValidationReport{}
}

func blankDefinition(name, description string, minPlayers, maxPlayers int) DefinitionV1 {
	return DefinitionV1{
		SchemaVersion: 1,
		Metadata:      Metadata{Name: name, Description: description, MinPlayers: minPlayers, MaxPlayers: maxPlayers},
		Teams:         []Team{}, Categories: []Category{}, Abilities: []Ability{}, Roles: []Role{}, Phases: []Phase{},
		KnowledgeRules: []KnowledgeRule{}, CompositionBands: []CompositionBand{}, CompositionModifiers: []CompositionModifier{},
		Chat:         ChatPolicy{DefaultPolicy: ChatPolicyDefaults{Teams: map[string]RoomPermission{}}, PhaseOverrides: map[string]ChatPolicyOverride{}, Channels: []ChatChannel{}},
		Achievements: []Achievement{}, AudioCues: []AudioCue{}, AssetAccessibility: map[string]AssetAccessibility{},
	}
}

func normalizeDefinitionIdentifiers(definition DefinitionV1) (DefinitionV1, error) {
	assign := func(prefix string, value *string) error {
		if *value != "" {
			return nil
		}
		generated, err := newOpaqueID(prefix)
		if err != nil {
			return err
		}
		*value = generated
		return nil
	}
	for index := range definition.Teams {
		if err := assign("team", &definition.Teams[index].ID); err != nil {
			return definition, err
		}
	}
	for index := range definition.Categories {
		if err := assign("category", &definition.Categories[index].ID); err != nil {
			return definition, err
		}
	}
	for index := range definition.Abilities {
		if err := assign("ability", &definition.Abilities[index].ID); err != nil {
			return definition, err
		}
	}
	for index := range definition.Roles {
		if err := assign("role", &definition.Roles[index].ID); err != nil {
			return definition, err
		}
	}
	for index := range definition.Phases {
		if err := assign("phase", &definition.Phases[index].ID); err != nil {
			return definition, err
		}
	}
	for bandIndex := range definition.CompositionBands {
		if err := assign("band", &definition.CompositionBands[bandIndex].ID); err != nil {
			return definition, err
		}
		for slotIndex := range definition.CompositionBands[bandIndex].Slots {
			if err := assign("slot", &definition.CompositionBands[bandIndex].Slots[slotIndex].ID); err != nil {
				return definition, err
			}
		}
	}
	for index := range definition.CompositionModifiers {
		if err := assign("modifier", &definition.CompositionModifiers[index].ID); err != nil {
			return definition, err
		}
	}
	for index := range definition.Chat.Channels {
		if err := assign("channel", &definition.Chat.Channels[index].ID); err != nil {
			return definition, err
		}
	}
	for index := range definition.Achievements {
		if err := assign("achievement", &definition.Achievements[index].ID); err != nil {
			return definition, err
		}
	}
	for index := range definition.AudioCues {
		if err := assign("audio", &definition.AudioCues[index].ID); err != nil {
			return definition, err
		}
	}
	return definition, nil
}

func newOpaqueID(prefix string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	encoded := hex.EncodeToString(bytes)
	return prefix + "_" + encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}

func generatedSlug(app core.App, name string) (string, error) {
	base := strings.TrimSuffix(importSlug(name, ""), "-"+checksum([]byte(""))[:6])
	for suffix := 0; ; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s-%d", base, suffix+1)
		}
		count, err := app.CountRecords("rulesets", dbx.HashExp{"slug": candidate})
		if err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
}

func projectVersion(record *core.Record) map[string]any {
	definition, _ := definitionFromRecord(record)
	return map[string]any{
		"id":                 record.Id,
		"rulesetId":          record.GetString("ruleset"),
		"versionNumber":      record.GetInt("version_number"),
		"state":              record.GetString("state"),
		"schemaVersion":      record.GetInt("schema_version"),
		"definition":         definition,
		"definitionChecksum": record.GetString("definition_checksum"),
		"publishedAt":        record.GetDateTime("published_at").Time().UTC(),
	}
}

func validationAppError(report ValidationReport) result.AppError {
	fields := result.FieldErrors{}
	for _, issue := range report.Errors {
		fields[issue.Path] = append(fields[issue.Path], issue.Message)
	}
	return result.Invalid("ruleset.validation_failed", "Correct the ruleset validation errors before continuing.", fields)
}

func importSlug(name, unique string) string {
	base := strings.ToLower(strings.TrimSpace(name))
	var cleaned strings.Builder
	for _, character := range base {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			cleaned.WriteRune(character)
		} else if cleaned.Len() > 0 && !strings.HasSuffix(cleaned.String(), "-") {
			cleaned.WriteByte('-')
		}
		if cleaned.Len() >= 24 {
			break
		}
	}
	value := strings.Trim(cleaned.String(), "-")
	if value == "" || value[0] < 'a' || value[0] > 'z' {
		value = "ruleset"
	}
	suffix := checksum([]byte(unique))[:6]
	return value + "-" + suffix
}

func rulesetNotFound(event *core.RequestEvent) error {
	return httpx.WriteError(event, result.AppError{Code: "ruleset.not_found", Message: "Ruleset not found.", Status: http.StatusNotFound})
}

func versionNotFound(event *core.RequestEvent) error {
	return httpx.WriteError(event, result.AppError{Code: "ruleset_version.not_found", Message: "Ruleset version not found.", Status: http.StatusNotFound})
}
