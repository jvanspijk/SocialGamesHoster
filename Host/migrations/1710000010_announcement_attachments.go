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
		items, err := app.FindCollectionByNameOrId("attention_items")
		if err != nil {
			return err
		}
		masters, err := app.FindCollectionByNameOrId("game_masters")
		if err != nil {
			return err
		}

		gameRelation := relation("game", games.Id, true)
		gameRelation.CascadeDelete = true
		announcementRelation := relation("announcement", items.Id, false)
		announcementRelation.CascadeDelete = true
		attachments := core.NewBaseCollection("announcement_attachments")
		lockRules(attachments)
		attachments.Fields.Add(
			gameRelation,
			announcementRelation,
			relation("creator", masters.Id, true),
			selectField("kind", true, "image", "audio"),
			&core.FileField{Name: "file", Required: true, MaxSize: 5 << 20, MaxSelect: 1, MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "audio/mpeg", "audio/mp4", "audio/ogg", "audio/wav"}, Protected: true},
			text("mime_type", true, 1, 100),
			&core.TextField{Name: "checksum", Required: true, Max: 64},
			selectField("storage_state", true, "staging", "ready"),
		)
		attachments.AddIndex("idx_announcement_attachments_kind", true, "announcement,kind", "announcement != ''")
		attachments.AddIndex("idx_announcement_attachments_game", false, "game,storage_state", "")
		if err := app.Save(attachments); err != nil {
			return err
		}

		items.Fields.Add(
			relation("image_attachment", attachments.Id, false),
			relation("audio_attachment", attachments.Id, false),
		)
		return app.Save(items)
	}, func(app core.App) error {
		if items, err := app.FindCollectionByNameOrId("attention_items"); err == nil {
			items.Fields.RemoveByName("image_attachment")
			items.Fields.RemoveByName("audio_attachment")
			if err := app.Save(items); err != nil {
				return err
			}
		}
		if attachments, err := app.FindCollectionByNameOrId("announcement_attachments"); err == nil {
			return app.Delete(attachments)
		}
		return nil
	}, "1710000010_announcement_attachments.go")
}
