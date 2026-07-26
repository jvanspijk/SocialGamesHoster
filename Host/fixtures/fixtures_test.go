package fixtures

import (
	"testing"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

func TestEmbeddedBundlesValidate(t *testing.T) {
	for _, name := range []string{"town-of-salem.sghrules", "blackjack.sghrules"} {
		data, err := Bundles.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		imported, err := rulesets.ReadBundle(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if report := rulesets.Validate(imported.Definition, map[string]struct{}{}); !report.Valid() {
			t.Fatalf("%s: %#v", name, report.Errors)
		}
	}
}
