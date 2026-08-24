package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		rulesets, err := app.FindCollectionByNameOrId("rulesets")
		if err != nil {
			return err
		}
		versions, err := app.FindCollectionByNameOrId("ruleset_versions")
		if err != nil {
			return err
		}
		masters, err := app.FindCollectionByNameOrId("game_masters")
		if err != nil {
			return err
		}

		sessions := core.NewBaseCollection("ruleset_edit_sessions")
		lockRules(sessions)
		sessions.Fields.Add(
			relation("ruleset", rulesets.Id, true),
			relation("base_version", versions.Id, true),
			relation("creator", masters.Id, true),
			&core.DateField{Name: "activity_at", Required: true},
			&core.DateField{Name: "expires_at", Required: true},
		)
		sessions.AddIndex("idx_ruleset_edit_sessions_owner", true, "ruleset,creator", "")
		sessions.AddIndex("idx_ruleset_edit_sessions_expiry", false, "expires_at", "")
		if err := app.Save(sessions); err != nil {
			return err
		}

		changes := core.NewBaseCollection("ruleset_asset_changes")
		lockRules(changes)
		changes.Fields.Add(
			relation("session", sessions.Id, true),
			text("asset_key", true, 1, 64),
			selectField("operation", true, "add", "replace", "update", "delete"),
			selectField("kind", true, "image", "audio"),
			&core.FileField{Name: "file", MaxSize: 5 << 20, MaxSelect: 1, MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "audio/mpeg", "audio/mp4", "audio/ogg", "audio/wav"}, Protected: true},
			text("display_name", true, 1, 160),
			&core.TextField{Name: "accessibility_text", Max: 1000},
			&core.TextField{Name: "mime_type", Max: 100},
			&core.TextField{Name: "checksum", Max: 64},
			&core.JSONField{Name: "metadata", MaxSize: 16 << 10},
		)
		changes.AddIndex("idx_ruleset_asset_changes_key", true, "session,asset_key", "")
		if err := app.Save(changes); err != nil {
			return err
		}

		assets, err := app.FindCollectionByNameOrId("ruleset_assets")
		if err != nil {
			return err
		}
		if assets.Fields.GetByName("accessibility_text") == nil {
			assets.Fields.Add(&core.TextField{Name: "accessibility_text", Max: 1000})
			if err := app.Save(assets); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		for _, name := range []string{"ruleset_asset_changes", "ruleset_edit_sessions"} {
			if collection, err := app.FindCollectionByNameOrId(name); err == nil {
				if err := app.Delete(collection); err != nil {
					return err
				}
			}
		}
		if assets, err := app.FindCollectionByNameOrId("ruleset_assets"); err == nil {
			assets.Fields.RemoveByName("accessibility_text")
			return app.Save(assets)
		}
		return nil
	}, "1710000009_ruleset_edit_sessions.go")
}
