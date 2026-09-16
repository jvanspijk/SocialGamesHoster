package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		return renameParticipantNumberField(app, "seat_number", "player_number")
	}, func(app core.App) error {
		return renameParticipantNumberField(app, "player_number", "seat_number")
	}, "1710000012_player_number.go")
}

func renameParticipantNumberField(app core.App, from, to string) error {
	participants, err := app.FindCollectionByNameOrId("participants")
	if err != nil {
		return err
	}
	if participants.Fields.GetByName(to) != nil {
		return nil
	}
	field, ok := participants.Fields.GetByName(from).(*core.NumberField)
	if !ok {
		return fmt.Errorf("participants field %q is not a number field", from)
	}
	field.Name = to
	participants.RemoveIndex("idx_participants_game_seat")
	participants.RemoveIndex("idx_participants_game_player_number")
	indexName := "idx_participants_game_player_number"
	if to == "seat_number" {
		indexName = "idx_participants_game_seat"
	}
	participants.AddIndex(indexName, true, "game,"+to, "")
	return app.Save(participants)
}
