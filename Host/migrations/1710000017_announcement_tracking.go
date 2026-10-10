package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		items, err := app.FindCollectionByNameOrId("attention_items")
		if err != nil {
			return err
		}
		items.Fields.Add(&core.BoolField{Name: "tracking_cleared"})
		return app.Save(items)
	}, func(app core.App) error {
		items, err := app.FindCollectionByNameOrId("attention_items")
		if err != nil {
			return err
		}
		items.Fields.RemoveByName("tracking_cleared")
		return app.Save(items)
	}, "1710000017_announcement_tracking.go")
}
