package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId(gameMastersCollection)
		if err != nil {
			return err
		}
		password, ok := collection.Fields.GetByName(core.FieldNamePassword).(*core.PasswordField)
		if !ok {
			return nil
		}
		password.Min = 6
		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId(gameMastersCollection)
		if err != nil {
			return err
		}
		password, ok := collection.Fields.GetByName(core.FieldNamePassword).(*core.PasswordField)
		if !ok {
			return nil
		}
		password.Min = 8
		return app.Save(collection)
	}, "1710000002_game_master_password_minimum.go")
}
