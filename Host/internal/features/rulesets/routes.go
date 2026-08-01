package rulesets

import (
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

type createRequest struct {
	Slug       string       `json:"slug"`
	Definition DefinitionV1 `json:"definition"`
}

type updateVersionRequest struct {
	Definition DefinitionV1 `json:"definition"`
}

func RegisterRoutes(event *core.ServeEvent, applicationVersion string) {
	group := event.Router.Group("/api/app/v1")
	group.BindFunc(actorauth.RequireGameMaster)
	group.GET("/rulesets", listRulesets)
	group.POST("/rulesets", createRuleset)
	group.GET("/rulesets/{id}", getRuleset)
	group.DELETE("/rulesets/{id}", deleteRuleset)
	group.POST("/rulesets/{id}/draft", createDraft)
	group.POST("/rulesets/{id}/save", saveRuleset)
	group.POST("/rulesets/{id}/archive", archiveRuleset)
	group.POST("/rulesets/import", func(event *core.RequestEvent) error {
		return importBundle(event, applicationVersion)
	})
	group.PATCH("/ruleset-versions/{id}", updateVersion)
	group.POST("/ruleset-versions/{id}/validate", validateVersion)
	group.POST("/ruleset-versions/{id}/publish", publishVersion)
	group.POST("/ruleset-versions/{id}/duplicate", duplicateVersion)
	group.GET("/ruleset-versions/{id}/export", func(event *core.RequestEvent) error {
		return exportBundle(event, applicationVersion)
	})
	registerAssetRoutes(group)
	registerAssetPreviewRoute(event)
}

func listRulesets(event *core.RequestEvent) error {
	records, err := event.App.FindRecordsByFilter("rulesets", "", "archived,name", 200, 0)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	response := make([]map[string]any, len(records))
	for i, record := range records {
		response[i] = projectRuleset(record)
	}
	return event.JSON(http.StatusOK, response)
}

func createRuleset(event *core.RequestEvent) error {
	var request createRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid", "The ruleset could not be read.", nil))
	}
	request.Slug = strings.ToLower(strings.TrimSpace(request.Slug))
	if !stableIDPattern.MatchString(request.Slug) {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid_slug", "Use a short lowercase slug containing letters, numbers, hyphens, or underscores.", nil))
	}
	var rulesetRecord *core.Record
	var versionRecord *core.Record
	err := event.App.RunInTransaction(func(txApp core.App) error {
		var err error
		rulesetRecord, versionRecord, err = createDraftRecords(txApp, request.Slug, request.Definition, event.Auth.Id, nil)
		return err
	})
	if err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.save_failed", "The ruleset could not be created. The slug may already be in use.", nil))
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", "ruleset.created", "ruleset", rulesetRecord.Id,
		map[string]any{"slug": request.Slug}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusCreated, map[string]any{
		"ruleset": projectRuleset(rulesetRecord),
		"draft":   projectVersion(versionRecord),
	})
}

func getRuleset(event *core.RequestEvent) error {
	record, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return rulesetNotFound(event)
	}
	versions, err := event.App.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset}", "-version_number", 100, 0, dbx.Params{"ruleset": record.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	projectedVersions := make([]map[string]any, len(versions))
	for i, version := range versions {
		projectedVersions[i] = projectVersion(version)
	}
	return event.JSON(http.StatusOK, map[string]any{
		"ruleset":  projectRuleset(record),
		"versions": projectedVersions,
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
}

func saveRuleset(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return rulesetNotFound(event)
	}
	var request saveRulesetRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.invalid", "The ruleset definition could not be read.", nil))
	}
	var saved *core.Record
	var report ValidationReport
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
		} else {
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
			if err := cloneVersionAssets(tx, source.Id, saved.Id); err != nil {
				return err
			}
		}
		saved.Set("definition", request.Definition)
		saved.Set("definition_checksum", "")
		if err := tx.Save(saved); err != nil {
			return err
		}
		assetKeys, err := versionAssetKeys(tx, saved.Id)
		if err != nil {
			return err
		}
		report = Validate(request.Definition, assetKeys)
		currentLogical.Set("name", request.Definition.Metadata.Name)
		if !report.Valid() {
			currentLogical.Set("latest_published_version", "")
			return tx.Save(currentLogical)
		}
		canonical, err := json.Marshal(request.Definition)
		if err != nil {
			return err
		}
		saved.Set("state", "published")
		saved.Set("definition_checksum", checksum(canonical))
		saved.Set("published_by", event.Auth.Id)
		saved.Set("published_at", time.Now().UTC())
		if err := tx.Save(saved); err != nil {
			return err
		}
		currentLogical.Set("latest_published_version", saved.Id)
		return tx.Save(currentLogical)
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	availability := "invalid"
	action := "ruleset.saved_invalid"
	if report.Valid() {
		availability = "ready"
		action = "ruleset.saved_ready"
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", action, "ruleset_version", saved.Id,
		map[string]any{"rulesetId": logical.Id, "availability": availability}, event.Get(httpx.TraceIDKey))
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
		return nil
	})
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", "ruleset.published", "ruleset_version", published.Id,
		map[string]any{"rulesetId": published.GetString("ruleset")}, event.Get(httpx.TraceIDKey))
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
	collection, err := event.App.FindCollectionByNameOrId("ruleset_versions")
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	draft := core.NewRecord(collection)
	draft.Set("ruleset", logical.Id)
	draft.Set("version_number", source.GetInt("version_number")+1)
	draft.Set("state", "draft")
	draft.Set("schema_version", 1)
	draft.Set("definition", definition)
	draft.Set("created_by", event.Auth.Id)
	if err := event.App.Save(draft); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if err := cloneVersionAssets(event.App, source.Id, draft.Id); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", "ruleset.draft_created", "ruleset_version", draft.Id,
		map[string]any{"rulesetId": logical.Id}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusCreated, projectVersion(draft))
}

func cloneVersionAssets(app core.App, sourceVersionID, targetVersionID string) error {
	sourceAssets, err := app.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": sourceVersionID})
	if err != nil || len(sourceAssets) == 0 {
		return err
	}
	collection, err := app.FindCollectionByNameOrId("ruleset_assets")
	if err != nil {
		return err
	}
	fsys, err := app.NewFilesystem()
	if err != nil {
		return err
	}
	defer fsys.Close()

	for _, source := range sourceAssets {
		filename := source.GetString("file")
		reader, err := fsys.GetReader(source.BaseFilesPath() + "/" + filename)
		if err != nil {
			return err
		}
		content, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		clonedFile, err := filesystem.NewFileFromBytes(content, path.Base(filename))
		if err != nil {
			return err
		}
		target := core.NewRecord(collection)
		target.Set("ruleset_version", targetVersionID)
		target.Set("asset_key", source.GetString("asset_key"))
		target.Set("kind", source.GetString("kind"))
		target.Set("file", clonedFile)
		target.Set("mime_type", source.GetString("mime_type"))
		target.Set("checksum", source.GetString("checksum"))
		target.Set("metadata", source.Get("metadata"))
		if err := app.Save(target); err != nil {
			return err
		}
	}
	return nil
}

func archiveRuleset(event *core.RequestEvent) error {
	record, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return rulesetNotFound(event)
	}
	record.Set("archived", true)
	if err := event.App.Save(record); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", "ruleset.archived", "ruleset", record.Id,
		nil, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusOK, projectRuleset(record))
}

func deleteRuleset(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil {
		return rulesetNotFound(event)
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
			return httpx.WriteError(event, result.Conflict("ruleset.in_use", "Archive this ruleset instead; one or more games use it."))
		}
	}
	err = event.App.RunInTransaction(func(txApp core.App) error {
		logical, err := txApp.FindRecordById("rulesets", logical.Id)
		if err != nil {
			return err
		}
		logical.Set("latest_published_version", "")
		if err := txApp.Save(logical); err != nil {
			return err
		}
		for _, version := range versions {
			assets, err := txApp.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": version.Id})
			if err != nil {
				return err
			}
			for _, asset := range assets {
				if err := txApp.Delete(asset); err != nil {
					return err
				}
			}
			if err := txApp.Delete(version); err != nil {
				return err
			}
		}
		return txApp.Delete(logical)
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", "ruleset.deleted", "ruleset", logical.Id,
		map[string]any{"slug": logical.GetString("slug")}, event.Get(httpx.TraceIDKey))
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
		return err
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", "ruleset.duplicated", "ruleset", logical.Id,
		map[string]any{"sourceVersionId": source.Id}, event.Get(httpx.TraceIDKey))
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
		return httpx.WriteError(event, result.Invalid("bundle.invalid", err.Error(), nil))
	}
	slug := importSlug(imported.Manifest.Name, imported.Manifest.RulesetChecksum)
	var logical, draft *core.Record
	err = event.App.RunInTransaction(func(txApp core.App) error {
		source := map[string]any{
			"sourceApplicationVersion": imported.Manifest.SourceApplicationVersion,
			"logicalSourceRulesetId":   imported.Manifest.LogicalSourceRulesetID,
			"sourceVersionNumber":      imported.Manifest.SourceVersionNumber,
			"importedByVersion":        applicationVersion,
		}
		var err error
		logical, draft, err = createDraftRecords(txApp, slug, imported.Definition, event.Auth.Id, source)
		if err != nil {
			return err
		}
		collection, err := txApp.FindCollectionByNameOrId("ruleset_assets")
		if err != nil {
			return err
		}
		for _, manifestAsset := range imported.Manifest.Assets {
			content := imported.Assets[manifestAsset.AssetKey]
			file, err := filesystem.NewFileFromBytes(content, path.Base(manifestAsset.Path))
			if err != nil {
				return err
			}
			mimeType, metadata, err := validateDeclaredAsset(manifestAsset.Kind, manifestAsset.MIMEType, content)
			if err != nil {
				return err
			}
			record := core.NewRecord(collection)
			record.Set("ruleset_version", draft.Id)
			record.Set("asset_key", manifestAsset.AssetKey)
			record.Set("kind", manifestAsset.Kind)
			record.Set("file", file)
			record.Set("mime_type", mimeType)
			record.Set("checksum", manifestAsset.Checksum)
			record.Set("metadata", metadata)
			if err := txApp.Save(record); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	_ = applicationaudit.Record(event.App, event.Auth, "", "ruleset.imported", "ruleset", logical.Id,
		map[string]any{"sourceVersion": imported.Manifest.SourceVersionNumber}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusCreated, map[string]any{"ruleset": projectRuleset(logical), "draft": projectVersion(draft)})
}

func exportBundle(event *core.RequestEvent, applicationVersion string) error {
	version, err := event.App.FindRecordById("ruleset_versions", event.Request.PathValue("id"))
	if err != nil {
		return versionNotFound(event)
	}
	definition, err := definitionFromRecord(version)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	logical, err := event.App.FindRecordById("rulesets", version.GetString("ruleset"))
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	assetRecords, err := event.App.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": version.Id})
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
			Path:     assetPath,
			AssetKey: asset.GetString("asset_key"),
			Kind:     asset.GetString("kind"),
			MIMEType: asset.GetString("mime_type"),
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
	records, err := app.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": versionID})
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
		"id":                     record.Id,
		"slug":                   record.GetString("slug"),
		"name":                   record.GetString("name"),
		"archived":               record.GetBool("archived"),
		"latestPublishedVersion": record.GetString("latest_published_version"),
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

func writeValidationError(event *core.RequestEvent, report ValidationReport) error {
	return httpx.WriteError(event, validationAppError(report))
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
