package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

type fixture struct {
	source string
	output string
	slug   string
}

func main() {
	fixtures := []fixture{
		{source: "Host/fixtures/sources/town-of-salem.json", output: "Host/fixtures/town-of-salem.sghrules", slug: "town-of-salem"},
		{source: "Host/fixtures/sources/blackjack.json", output: "Host/fixtures/blackjack.sghrules", slug: "blackjack"},
	}
	for _, item := range fixtures {
		if err := build(item); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func build(item fixture) error {
	data, err := os.ReadFile(item.source)
	if err != nil {
		return err
	}
	var definition rulesets.DefinitionV1
	if err := json.Unmarshal(data, &definition); err != nil {
		return err
	}
	report := rulesets.Validate(definition, map[string]struct{}{})
	if !report.Valid() {
		return fmt.Errorf("%s: %s at %s", item.source, report.Errors[0].Message, report.Errors[0].Path)
	}
	bundle, err := rulesets.WriteBundle(rulesets.BundleManifest{
		SourceApplicationVersion:  "1.0.0",
		MinimumApplicationVersion: "1.0.0",
		LogicalSourceRulesetID:    item.slug,
		SourceVersionNumber:       1,
		Name:                      definition.Metadata.Name,
		Description:               definition.Metadata.Description,
	}, definition, map[string][]byte{})
	if err != nil {
		return err
	}
	if err := os.WriteFile(item.output, bundle, 0o644); err != nil {
		return err
	}
	fmt.Println(filepath.ToSlash(item.output))
	return nil
}
