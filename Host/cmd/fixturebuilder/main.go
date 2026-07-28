package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

var assetKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)

type fixture struct {
	source    string
	assetRoot string
	output    string
	slug      string
}

func main() {
	fixtures := []fixture{
		{source: "Host/fixtures/sources/blackjack.json", output: "Host/fixtures/blackjack.sghrules", slug: "blackjack"},
		{
			source:    "Host/fixtures/sources/echo-location.json",
			assetRoot: "Host/fixtures/sources/echo-location-assets/generated",
			output:    "Host/fixtures/echo-location.sghrules",
			slug:      "echo-location",
		},
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
	manifestAssets, assets, err := discoverAssets(item.assetRoot)
	if err != nil {
		return err
	}
	assetKeys := make(map[string]struct{}, len(assets))
	for key := range assets {
		assetKeys[key] = struct{}{}
	}
	report := rulesets.Validate(definition, assetKeys)
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
		Assets:                    manifestAssets,
	}, definition, assets)
	if err != nil {
		return err
	}
	if _, err := rulesets.ReadBundle(bundle); err != nil {
		return fmt.Errorf("%s: generated bundle is invalid: %w", item.source, err)
	}
	if err := os.WriteFile(item.output, bundle, 0o644); err != nil {
		return err
	}
	fmt.Println(filepath.ToSlash(item.output))
	return nil
}

func discoverAssets(root string) ([]rulesets.BundleAssetManifest, map[string][]byte, error) {
	if root == "" {
		return nil, map[string][]byte{}, nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", root, err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	manifest := make([]rulesets.BundleAssetManifest, 0, len(entries))
	assets := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, nil, fmt.Errorf("%s: nested asset directory %q is not supported", root, entry.Name())
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		kind, mimeType, ok := supportedAssetType(extension)
		if !ok {
			return nil, nil, fmt.Errorf("%s: asset %q has an unsupported file type", root, entry.Name())
		}
		key := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if !assetKeyPattern.MatchString(key) {
			return nil, nil, fmt.Errorf("%s: asset %q does not infer a valid stable key", root, entry.Name())
		}
		if _, duplicate := assets[key]; duplicate {
			return nil, nil, fmt.Errorf("%s: asset key %q is duplicated", root, key)
		}
		content, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, nil, err
		}
		assets[key] = content
		manifest = append(manifest, rulesets.BundleAssetManifest{
			Path:     "assets/" + entry.Name(),
			AssetKey: key,
			Kind:     kind,
			MIMEType: mimeType,
		})
	}
	return manifest, assets, nil
}

func supportedAssetType(extension string) (kind, mimeType string, ok bool) {
	switch extension {
	case ".jpg", ".jpeg":
		return "image", "image/jpeg", true
	case ".png":
		return "image", "image/png", true
	case ".webp":
		return "image", "image/webp", true
	case ".mp3":
		return "audio", "audio/mpeg", true
	case ".m4a":
		return "audio", "audio/mp4", true
	case ".ogg":
		return "audio", "audio/ogg", true
	case ".wav":
		return "audio", "audio/wav", true
	default:
		return "", "", false
	}
}
