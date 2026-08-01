package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		assets, err := app.FindCollectionByNameOrId("ruleset_assets")
		if err != nil {
			return err
		}
		if field := assets.Fields.GetByName("file"); field != nil {
			field.(*core.FileField).Required = false
		}
		if assets.Fields.GetByName("storage_state") == nil {
			assets.Fields.Add(selectField("storage_state", true, "staging", "ready"))
		}
		if err := app.Save(assets); err != nil {
			return err
		}
		// Existing assets predate staging and are complete, so preserve their
		// visibility on upgrade rather than treating them as transient work.
		records, err := app.FindAllRecords("ruleset_assets")
		if err != nil {
			return err
		}
		for _, record := range records {
			if record.GetString("storage_state") == "" {
				record.Set("storage_state", "ready")
				if err := app.Save(record); err != nil {
					return err
				}
			}
		}
		return nil
	}, func(app core.App) error {
		assets, err := app.FindCollectionByNameOrId("ruleset_assets")
		if err != nil {
			return err
		}
		assets.Fields.RemoveByName("storage_state")
		if field := assets.Fields.GetByName("file"); field != nil {
			field.(*core.FileField).Required = true
		}
		return app.Save(assets)
	}, "1710000007_staged_ruleset_assets.go")
}
