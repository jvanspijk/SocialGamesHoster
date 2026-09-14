package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

const mediaMaxSize = 64 << 20

var mediaFileFields = []struct {
	collection    string
	field         string
	previousLimit int64
}{
	{collection: "player_profiles", field: "avatar", previousLimit: 1 << 20},
	{collection: "ruleset_assets", field: "file", previousLimit: 5 << 20},
	{collection: "ruleset_asset_changes", field: "file", previousLimit: 5 << 20},
	{collection: "announcement_attachments", field: "file", previousLimit: 5 << 20},
}

func init() {
	pbmigrations.Register(func(app core.App) error {
		for _, target := range mediaFileFields {
			collection, err := app.FindCollectionByNameOrId(target.collection)
			if err != nil {
				return err
			}
			field, ok := collection.Fields.GetByName(target.field).(*core.FileField)
			if !ok {
				return fmt.Errorf("collection %q field %q is not a file field", target.collection, target.field)
			}
			field.MaxSize = mediaMaxSize
			if err := app.Save(collection); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		for _, target := range mediaFileFields {
			collection, err := app.FindCollectionByNameOrId(target.collection)
			if err != nil {
				return err
			}
			field, ok := collection.Fields.GetByName(target.field).(*core.FileField)
			if !ok {
				return fmt.Errorf("collection %q field %q is not a file field", target.collection, target.field)
			}
			field.MaxSize = target.previousLimit
			if err := app.Save(collection); err != nil {
				return err
			}
		}
		return nil
	}, "1710000011_media_and_bundle_limits.go")
}
