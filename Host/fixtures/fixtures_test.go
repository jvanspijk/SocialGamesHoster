package fixtures

import (
	"testing"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

func TestEmbeddedBundlesValidate(t *testing.T) {
	for _, name := range []string{"blackjack.sghrules", "echo-location.sghrules"} {
		data, err := Bundles.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		imported, err := rulesets.ReadBundle(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		assetKeys := make(map[string]struct{}, len(imported.Assets))
		for key := range imported.Assets {
			assetKeys[key] = struct{}{}
		}
		if report := rulesets.Validate(imported.Definition, assetKeys); !report.Valid() {
			t.Fatalf("%s: %#v", name, report.Errors)
		}
		if name == "echo-location.sghrules" {
			if imported.Definition.Metadata.Name != "Echo Location" ||
				imported.Manifest.LogicalSourceRulesetID != "echo-location" {
				t.Fatal("echo-location.sghrules does not use the Echo Location identity")
			}
			if len(imported.Assets) != 47 {
				t.Fatalf("echo-location.sghrules contains %d assets, want 47", len(imported.Assets))
			}
			for _, asset := range imported.Manifest.Assets {
				if asset.Kind == "image" && asset.MIMEType != "image/webp" {
					t.Fatalf("image asset %q uses %q, want image/webp", asset.AssetKey, asset.MIMEType)
				}
				if asset.Kind == "audio" && asset.MIMEType != "audio/ogg" {
					t.Fatalf("audio asset %q uses %q, want audio/ogg", asset.AssetKey, asset.MIMEType)
				}
			}
			if imported.Definition.Metadata.MinPlayers != 3 ||
				imported.Definition.Metadata.MaxPlayers != 3 ||
				len(imported.Definition.Roles) != 3 ||
				len(imported.Definition.Abilities) != 4 ||
				len(imported.Definition.Chat.Channels) != 2 {
				t.Fatal("echo-location.sghrules does not contain the expected three-role game contract")
			}
			for _, ability := range imported.Definition.Abilities {
				if len(ability.ActivationPhaseIDs) != 1 ||
					ability.ActivationPhaseIDs[0] != "command" ||
					ability.CanCombineWithOtherAbilities {
					t.Fatalf("ability %q is not an exclusive Command choice", ability.ID)
				}
			}
			for _, key := range []string{
				"echo-location-cover",
				"safe-passage-badge",
				"lookout-practice",
				"sonar-practice",
				"dive-ambience",
			} {
				if _, ok := imported.Assets[key]; !ok {
					t.Fatalf("echo-location.sghrules is missing %q", key)
				}
			}
		}
	}
}
