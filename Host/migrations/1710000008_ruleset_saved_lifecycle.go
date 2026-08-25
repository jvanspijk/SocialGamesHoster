package migrations

import (
	"path"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"

	domainrulesets "github.com/jvanspijk/SocialGamesHoster/Host/internal/domain/rulesets"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		logical, err := app.FindCollectionByNameOrId("rulesets")
		if err != nil {
			return err
		}
		versions, err := app.FindCollectionByNameOrId("ruleset_versions")
		if err != nil {
			return err
		}
		assets, err := app.FindCollectionByNameOrId("ruleset_assets")
		if err != nil {
			return err
		}
		if logical.Fields.GetByName("latest_saved_version") == nil {
			logical.Fields.Add(relation("latest_saved_version", versions.Id, false))
		}
		if versions.Fields.GetByName("validation_report") == nil {
			versions.Fields.Add(&core.JSONField{Name: "validation_report", MaxSize: 256 << 10})
		}
		if assets.Fields.GetByName("display_name") == nil {
			assets.Fields.Add(text("display_name", false, 0, 160))
		}
		if err := app.Save(logical); err != nil {
			return err
		}
		if err := app.Save(versions); err != nil {
			return err
		}
		if err := app.Save(assets); err != nil {
			return err
		}

		for _, record := range mustFindAll(app, "ruleset_assets") {
			if strings.TrimSpace(record.GetString("display_name")) != "" {
				continue
			}
			name := strings.TrimSpace(path.Base(record.GetString("file")))
			if name == "." || name == "" {
				if record.GetString("kind") == "audio" {
					name = "Untitled audio"
				} else {
					name = "Untitled image"
				}
			}
			record.Set("display_name", name)
			if err := app.Save(record); err != nil {
				return err
			}
		}
		for _, logicalRecord := range mustFindAll(app, "rulesets") {
			if logicalRecord.GetString("latest_saved_version") != "" {
				continue
			}
			items, err := app.FindRecordsByFilter("ruleset_versions", "ruleset = {:ruleset}", "-version_number", 1, 0, map[string]any{"ruleset": logicalRecord.Id})
			if err != nil || len(items) == 0 {
				if err != nil {
					return err
				}
				continue
			}
			latest := items[0]
			definition, err := decodeMigrationDefinition(latest)
			if err != nil {
				return err
			}
			assetKeys := map[string]struct{}{}
			for _, asset := range mustFindAll(app, "ruleset_assets") {
				if asset.GetString("ruleset_version") == latest.Id && asset.GetString("storage_state") == "ready" {
					assetKeys[asset.GetString("asset_key")] = struct{}{}
				}
			}
			logicalRecord.Set("latest_saved_version", latest.Id)
			report := domainrulesets.Validate(definition, assetKeys)
			latest.Set("validation_report", report)
			if !report.Valid() {
				logicalRecord.Set("latest_published_version", "")
			}
			if err := app.Save(latest); err != nil {
				return err
			}
			if err := app.Save(logicalRecord); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		logical, err := app.FindCollectionByNameOrId("rulesets")
		if err == nil {
			logical.Fields.RemoveByName("latest_saved_version")
			if err := app.Save(logical); err != nil {
				return err
			}
		}
		versions, err := app.FindCollectionByNameOrId("ruleset_versions")
		if err == nil {
			versions.Fields.RemoveByName("validation_report")
			if err := app.Save(versions); err != nil {
				return err
			}
		}
		assets, err := app.FindCollectionByNameOrId("ruleset_assets")
		if err == nil {
			assets.Fields.RemoveByName("display_name")
			return app.Save(assets)
		}
		return nil
	}, "1710000008_ruleset_saved_lifecycle.go")
}

func mustFindAll(app core.App, collection string) []*core.Record {
	records, _ := app.FindAllRecords(collection)
	return records
}

func decodeMigrationDefinition(record *core.Record) (domainrulesets.DefinitionV1, error) {
	var definition domainrulesets.DefinitionV1
	err := record.UnmarshalJSONField("definition", &definition)
	return definition, err
}
