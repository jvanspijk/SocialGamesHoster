package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId(playerProfilesCollection)
		if err != nil {
			return err
		}
		collection.Fields.RemoveByName("accent")
		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId(playerProfilesCollection)
		if err != nil {
			return err
		}
		collection.Fields.Add(&core.SelectField{Name: "accent", Values: []string{"crimson", "forest", "navy", "gold", "plum"}})
		return app.Save(collection)
	}, "1710000015_remove_profile_accent.go")
}
