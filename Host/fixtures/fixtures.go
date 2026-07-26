package fixtures

import (
	"embed"
	"encoding/json"
	"time"

	"github.com/pocketbase/pocketbase/core"

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
		{name: "town-of-salem.sghrules", slug: "town-of-salem"},
		{name: "blackjack.sghrules", slug: "blackjack"},
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
			logical.Set("latest_published_version", version.Id)
			if err := txApp.Save(logical); err != nil {
				return err
			}
		}
		return nil
	})
}
