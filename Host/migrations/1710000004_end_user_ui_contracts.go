package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		games, err := app.FindCollectionByNameOrId("games")
		if err != nil {
			return err
		}
		games.Fields.Add(
			&core.BoolField{Name: "roles_visible"},
			intField("role_visibility_revision", 0),
			&core.TextField{Name: "completion_previous_status", Max: 16},
		)
		if err := app.Save(games); err != nil {
			return err
		}

		participants, err := app.FindCollectionByNameOrId("participants")
		if err != nil {
			return err
		}
		participants.Fields.Add(intField("role_revision", 0))
		if err := app.Save(participants); err != nil {
			return err
		}

		rooms, err := app.FindCollectionByNameOrId("chat_rooms")
		if err != nil {
			return err
		}
		rooms.Fields.Add(&core.BoolField{Name: "players_can_post"})
		if err := app.Save(rooms); err != nil {
			return err
		}
		_, err = app.DB().NewQuery(`
			UPDATE chat_rooms
			SET players_can_post = CASE WHEN manually_locked = 1 THEN 0 ELSE 1 END
		`).Execute()
		return err
	}, func(app core.App) error {
		if rooms, err := app.FindCollectionByNameOrId("chat_rooms"); err == nil {
			rooms.Fields.RemoveByName("players_can_post")
			if err := app.Save(rooms); err != nil {
				return err
			}
		}
		if participants, err := app.FindCollectionByNameOrId("participants"); err == nil {
			participants.Fields.RemoveByName("role_revision")
			if err := app.Save(participants); err != nil {
				return err
			}
		}
		if games, err := app.FindCollectionByNameOrId("games"); err == nil {
			games.Fields.RemoveByName("roles_visible")
			games.Fields.RemoveByName("role_visibility_revision")
			games.Fields.RemoveByName("completion_previous_status")
			if err := app.Save(games); err != nil {
				return err
			}
		}
		return nil
	}, "1710000004_end_user_ui_contracts.go")
}
