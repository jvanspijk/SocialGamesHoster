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
		gameMasters, err := app.FindCollectionByNameOrId(gameMastersCollection)
		if err != nil {
			return err
		}
		participants, err := app.FindCollectionByNameOrId("participants")
		if err != nil {
			return err
		}

		items := core.NewBaseCollection("attention_items")
		lockRules(items)
		items.Fields.Add(
			relation("game", games.Id, true),
			selectField("kind", true, "announcement"),
			relation("sender", gameMasters.Id, true),
			text("sender_label_snapshot", true, 1, 120),
			&core.TextField{Name: "content", Required: true, Hidden: true, Min: 1, Max: 1000},
			selectField("audience", true, "all", "team", "player"),
			&core.TextField{Name: "target_id", Hidden: true, Max: 64},
			&core.TextField{Name: "cue_key", Max: 32},
		)
		items.AddIndex("idx_attention_items_game_created", false, "game,created DESC,id DESC", "")
		if err := app.Save(items); err != nil {
			return err
		}

		receipts := core.NewBaseCollection("attention_receipts")
		lockRules(receipts)
		receipts.Fields.Add(
			relation("attention_item", items.Id, true),
			relation("participant", participants.Id, true),
			&core.DateField{Name: "acknowledged_at"},
		)
		receipts.AddIndex("idx_attention_receipts_unique", true, "attention_item,participant", "")
		receipts.AddIndex("idx_attention_receipts_participant", false, "participant,acknowledged_at,created", "")
		return app.Save(receipts)
	}, func(app core.App) error {
		for _, name := range []string{"attention_receipts", "attention_items"} {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}
			if err := app.Delete(collection); err != nil {
				return err
			}
		}
		return nil
	}, "1710000003_attention_items.go")
}
