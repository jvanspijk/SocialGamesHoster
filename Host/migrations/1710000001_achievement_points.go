package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("achievement_awards")
		if err != nil {
			return err
		}
		if collection.Fields.GetByName("points_snapshot") == nil {
			collection.Fields.Add(&core.NumberField{
				Name: "points_snapshot", OnlyInt: true, Min: number(0), Max: number(10000),
			})
		}
		if collection.Fields.GetByName("hidden_until_game_completed") == nil {
			collection.Fields.Add(&core.BoolField{Name: "hidden_until_game_completed"})
		}
		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("achievement_awards")
		if err != nil {
			return err
		}
		collection.Fields.RemoveByName("points_snapshot")
		collection.Fields.RemoveByName("hidden_until_game_completed")
		return app.Save(collection)
	}, "1710000001_achievement_points.go")
}
