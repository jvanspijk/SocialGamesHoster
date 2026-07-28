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
			&core.DateField{Name: "ability_phase_locked_at"},
			intField("ability_phase_instance", 0),
		)
		if err := app.Save(games); err != nil {
			return err
		}

		participants, err := app.FindCollectionByNameOrId("participants")
		if err != nil {
			return err
		}
		choices := core.NewBaseCollection("ability_choices")
		lockRules(choices)
		choices.Fields.Add(
			relation("game", games.Id, true),
			relation("participant", participants.Id, true),
			text("phase_key", true, 1, 32),
			intField("round_number", 0),
			intField("phase_instance", 0),
			text("ability_key", true, 1, 32),
			selectField("status", true, "activated", "finalized"),
			&core.DateField{Name: "activated_at", Required: true},
			&core.DateField{Name: "finalized_at"},
		)
		choices.AddIndex("idx_ability_choices_unique", true, "game,participant,phase_instance,ability_key", "")
		choices.AddIndex("idx_ability_choices_phase", false, "game,phase_instance,status", "")
		if err := app.Save(choices); err != nil {
			return err
		}

		items, err := app.FindCollectionByNameOrId("attention_items")
		if err != nil {
			return err
		}
		items.Fields.Add(
			&core.TextField{Name: "image_asset_key", Max: 64},
			&core.TextField{Name: "image_description", Max: 500},
			&core.TextField{Name: "audio_asset_key", Max: 64},
			&core.TextField{Name: "audio_alternative", Max: 1000},
		)
		return app.Save(items)
	}, func(app core.App) error {
		if items, err := app.FindCollectionByNameOrId("attention_items"); err == nil {
			items.Fields.RemoveByName("image_asset_key")
			items.Fields.RemoveByName("image_description")
			items.Fields.RemoveByName("audio_asset_key")
			items.Fields.RemoveByName("audio_alternative")
			if err := app.Save(items); err != nil {
				return err
			}
		}
		if choices, err := app.FindCollectionByNameOrId("ability_choices"); err == nil {
			if err := app.Delete(choices); err != nil {
				return err
			}
		}
		if games, err := app.FindCollectionByNameOrId("games"); err == nil {
			games.Fields.RemoveByName("ability_phase_locked_at")
			games.Fields.RemoveByName("ability_phase_instance")
			if err := app.Save(games); err != nil {
				return err
			}
		}
		return nil
	}, "1710000006_ability_activation_and_announcement_media.go")
}
