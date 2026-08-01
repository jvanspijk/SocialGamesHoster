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
		return app.Save(assets)
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
