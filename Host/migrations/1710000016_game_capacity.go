package migrations

import (
	"fmt"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/domain/gamecapacity"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		participants, err := app.FindCollectionByNameOrId("participants")
		if err != nil {
			return err
		}
		field, ok := participants.Fields.GetByName("player_number").(*core.NumberField)
		if !ok {
			return fmt.Errorf("participants player_number is not a number field")
		}
		maximum := float64(gamecapacity.MaxPlayers)
		field.Max = &maximum
		return app.Save(participants)
	}, func(app core.App) error {
		count, err := app.CountRecords("participants", dbx.NewExp("player_number > 30"))
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("cannot restore 30-player capacity while higher player numbers exist")
		}
		participants, err := app.FindCollectionByNameOrId("participants")
		if err != nil {
			return err
		}
		maximum := float64(30)
		participants.Fields.GetByName("player_number").(*core.NumberField).Max = &maximum
		return app.Save(participants)
	}, "1710000016_game_capacity.go")
}
