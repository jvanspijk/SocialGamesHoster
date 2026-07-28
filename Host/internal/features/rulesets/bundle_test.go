package rulesets

import (
	"archive/zip"
	"bytes"
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
	}
	data, err := WriteBundle(manifest, definition, map[string][]byte{})
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
