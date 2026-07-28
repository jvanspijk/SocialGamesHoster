package fixtures

import (
	"embed"
	"encoding/json"
	"path"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

// Bundles contains the validated demonstration rulesets seeded on first launch.
//
//go:embed *.sghrules
var Bundles embed.FS

func Seed(app core.App, gameMasterID string) error {
	items := []struct {
		name string
		slug string
	}{
		{name: "blackjack.sghrules", slug: "blackjack"},
		{name: "echo-location.sghrules", slug: "echo-location"},
	}
	return app.RunInTransaction(func(txApp core.App) error {
		rulesetsCollection, err := txApp.FindCollectionByNameOrId("rulesets")
		if err != nil {
			return err
		}
		versionsCollection, err := txApp.FindCollectionByNameOrId("ruleset_versions")
		if err != nil {
			return err
		}
		assetsCollection, err := txApp.FindCollectionByNameOrId("ruleset_assets")
		if err != nil {
			return err
		}
		for _, item := range items {
			if _, err := txApp.FindFirstRecordByData("rulesets", "slug", item.slug); err == nil {
				continue
			}
			data, err := Bundles.ReadFile(item.name)
			if err != nil {
				return err
			}
			imported, err := rulesets.ReadBundle(data)
			if err != nil {
				return err
			}
			canonical, err := json.Marshal(imported.Definition)
			if err != nil {
				return err
			}
			logical := core.NewRecord(rulesetsCollection)
			logical.Set("slug", item.slug)
			logical.Set("name", imported.Definition.Metadata.Name)
			logical.Set("created_by", gameMasterID)
			if err := txApp.Save(logical); err != nil {
				return err
			}
			version := core.NewRecord(versionsCollection)
			version.Set("ruleset", logical.Id)
			version.Set("version_number", 1)
			version.Set("state", "published")
			version.Set("schema_version", 1)
			version.Set("definition", imported.Definition)
			version.Set("definition_checksum", rulesets.DefinitionChecksum(canonical))
			version.Set("created_by", gameMasterID)
			version.Set("published_by", gameMasterID)
			version.Set("published_at", time.Now().UTC())
			version.Set("source_metadata", map[string]any{"embeddedFixture": item.name})
			if err := txApp.Save(version); err != nil {
				return err
			}
			for _, manifestAsset := range imported.Manifest.Assets {
				content := imported.Assets[manifestAsset.AssetKey]
				file, err := filesystem.NewFileFromBytes(content, path.Base(manifestAsset.Path))
				if err != nil {
					return err
				}
				asset := core.NewRecord(assetsCollection)
				asset.Set("ruleset_version", version.Id)
				asset.Set("asset_key", manifestAsset.AssetKey)
				asset.Set("kind", manifestAsset.Kind)
				asset.Set("file", file)
				asset.Set("mime_type", manifestAsset.MIMEType)
				asset.Set("checksum", manifestAsset.Checksum)
				asset.Set("metadata", map[string]any{})
				if err := txApp.Save(asset); err != nil {
					return err
				}
			}
			logical.Set("latest_published_version", version.Id)
			if err := txApp.Save(logical); err != nil {
				return err
			}
		}
		return nil
	})
}
