package migrations

import (
	"slices"

	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		rooms, err := app.FindCollectionByNameOrId("chat_rooms")
		if err != nil {
			return err
		}
		kind, ok := rooms.Fields.GetByName("kind").(*core.SelectField)
		if !ok {
			return nil
		}
		if !slices.Contains(kind.Values, "custom") {
			kind.Values = append(kind.Values, "custom")
		}
		return app.Save(rooms)
	}, func(app core.App) error {
		rooms, err := app.FindCollectionByNameOrId("chat_rooms")
		if err != nil {
			return nil
		}
		kind, ok := rooms.Fields.GetByName("kind").(*core.SelectField)
		if !ok {
			return nil
		}
		kind.Values = slices.DeleteFunc(kind.Values, func(value string) bool { return value == "custom" })
		return app.Save(rooms)
	}, "1710000005_ruleset_chat_channels.go")
}
