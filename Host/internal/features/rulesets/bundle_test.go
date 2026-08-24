package rulesets

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestBundleRoundTrip(t *testing.T) {
	definition := testDefinition()
	definition.Phases = []Phase{{ID: "night", Name: "Night", Order: 1}}
	definition.Abilities = []Ability{{
		ID: "inspect", Name: "Inspect", Description: "Inspect one player.",
		ActivationPhaseIDs: []string{"night"}, CanCombineWithOtherAbilities: true,
	}}
	definition.Chat.Channels = []ChatChannel{{
		ID: "town_square", Name: "Town square",
		ReaderTeamIDs: []string{"town"}, MessageRestriction: ChatEmojiOnly,
		Visible: true, Sendable: true, GameMasterMaySend: true,
		SenderDisplay: SenderGameAlias, PhaseOverrides: map[string]ChatChannelPhaseOverride{},
	}}
	manifest := BundleManifest{
		SourceApplicationVersion:  "test",
		MinimumApplicationVersion: "1.0.0",
		LogicalSourceRulesetID:    "test",
		SourceVersionNumber:       1,
		Name:                      "Test",
		Assets: []BundleAssetManifest{{
			Path: "assets/cover.png", AssetKey: "cover", Kind: "image", MIMEType: "image/png",
			DisplayName: "Village cover", AccessibilityText: "A quiet village at dusk",
		}},
	}
	definition.Metadata.CoverAssetKey = "cover"
	definition.AssetAccessibility = map[string]AssetAccessibility{"cover": {Description: "A quiet village at dusk"}}
	data, err := WriteBundle(manifest, definition, map[string][]byte{"cover": tinyPNG(t)})
	if err != nil {
		t.Fatal(err)
	}
	imported, err := ReadBundle(data)
	if err != nil {
		t.Fatal(err)
	}
	if imported.Definition.Metadata.Name != definition.Metadata.Name {
		t.Fatalf("unexpected definition: %#v", imported.Definition.Metadata)
	}
	if imported.Manifest.FormatVersion != 2 || imported.Manifest.Assets[0].DisplayName != "Village cover" || imported.Manifest.Assets[0].AccessibilityText != "A quiet village at dusk" {
		t.Fatalf("media metadata was not preserved: %#v", imported.Manifest.Assets)
	}
	if len(imported.Definition.Chat.Channels) != 1 ||
		imported.Definition.Chat.Channels[0].MessageRestriction != ChatEmojiOnly {
		t.Fatalf("custom chat channel was not preserved: %#v", imported.Definition.Chat.Channels)
	}
	if len(imported.Definition.Abilities) != 1 ||
		!imported.Definition.Abilities[0].CanCombineWithOtherAbilities ||
		len(imported.Definition.Abilities[0].ActivationPhaseIDs) != 1 {
		t.Fatalf("playable ability was not preserved: %#v", imported.Definition.Abilities)
	}
}

func TestReadBundleV1DerivesAssetDisplayName(t *testing.T) {
	definition := testDefinition()
	content := tinyPNG(t)
	definition.Metadata.CoverAssetKey = "cover"
	rulesetJSON, _ := json.Marshal(definition)
	manifest := BundleManifest{
		FormatVersion: 1, SourceApplicationVersion: "legacy", MinimumApplicationVersion: "legacy",
		LogicalSourceRulesetID: "legacy", SourceVersionNumber: 1, Name: "Legacy",
		RulesetChecksum: checksum(rulesetJSON),
		Assets:          []BundleAssetManifest{{Path: "assets/legacy-cover.png", AssetKey: "cover", Kind: "image", MIMEType: "image/png", ByteSize: int64(len(content)), Checksum: checksum(content)}},
	}
	manifestJSON, _ := json.Marshal(manifest)
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for name, data := range map[string][]byte{"manifest.json": manifestJSON, "ruleset.json": rulesetJSON, "assets/legacy-cover.png": content} {
		if err := writeZipFile(writer, name, data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	imported, err := ReadBundle(output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if imported.Manifest.Assets[0].DisplayName != "legacy-cover.png" {
		t.Fatalf("legacy display name was not derived: %#v", imported.Manifest.Assets[0])
	}
}

func TestBundleRoundTripsInvalidSavedWork(t *testing.T) {
	definition := testDefinition()
	definition.Teams = nil
	data, err := WriteBundle(BundleManifest{Name: "Incomplete"}, definition, map[string][]byte{})
	if err != nil {
		t.Fatal(err)
	}
	imported, err := ReadBundle(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(imported.Definition.Teams) != 0 {
		t.Fatalf("incomplete definition changed during round trip: %#v", imported.Definition.Teams)
	}
}

func TestBundleRejectsTraversal(t *testing.T) {
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	entry, err := writer.Create("../ruleset.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("{}")); err != nil {
		t.Fatal(err)
	}
	manifest, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manifest.Write([]byte("{}")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadBundle(output.Bytes()); err == nil || !strings.Contains(err.Error(), "unsafe path") {
		t.Fatalf("expected unsafe path rejection, got %v", err)
	}
}
