package rulesets

import (
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestBuildRulesetPreviewCoversEveryMode(t *testing.T) {
	visible, sendable := true, false
	definition := DefinitionV1{
		SchemaVersion: 1,
		Metadata:      Metadata{Name: "Preview game", MinPlayers: 3, MaxPlayers: 3},
		Teams:         []Team{{ID: "team-village", Name: "Village"}},
		Abilities:     []Ability{{ID: "ability-see", Name: "Inspect", Description: "Inspect one player."}},
		Roles: []Role{{
			ID: "role-villager", Name: "Villager", Description: "Find the threat.", TeamID: "team-village",
			AbilityIDs: []string{"ability-see"}, WinCondition: "Keep the village safe.", MaxCopies: 3, ImageAssetKey: "portrait",
		}},
		Phases: []Phase{{ID: "phase-night", Name: "Night", Description: "Close your eyes.", Order: 1, SuggestedDurationSeconds: 60, AudioCueID: "cue-bell"}},
		CompositionBands: []CompositionBand{{
			ID: "band-three", MinPlayers: 3, MaxPlayers: 3,
			Slots: []CompositionSlot{{ID: "slot-village", Label: "Village roles", Count: 3, Selector: Selector{TeamIDs: []string{"team-village"}}}},
		}},
		Chat: ChatPolicy{
			DefaultPolicy:  ChatPolicyDefaults{General: &RoomPermission{Visible: true, Readable: true, Sendable: true}, Teams: map[string]RoomPermission{}},
			PhaseOverrides: map[string]ChatPolicyOverride{"phase-night": {General: &PartialRoomPermission{Visible: &visible, Sendable: &sendable}}},
			Channels:       []ChatChannel{{ID: "channel-dead", Name: "Quiet room", Visible: true, Sendable: true, MessageRestriction: ChatEmojiOnly}},
		},
		AudioCues: []AudioCue{{ID: "cue-bell", Name: "Bell", AssetKey: "bell"}},
	}
	imageRecord := core.NewRecord(&core.Collection{})
	imageRecord.Id = "staged-image"
	imageRecord.Set("file", "portrait.png")
	audioRecord := core.NewRecord(&core.Collection{})
	audioRecord.Id = "saved-audio"
	assets := map[string]effectiveAsset{
		"portrait": {key: "portrait", kind: "image", displayName: "Village portrait", accessibilityText: "A villager", record: imageRecord, staged: true},
		"bell":     {key: "bell", kind: "audio", displayName: "Night bell", accessibilityText: "One bell", record: audioRecord},
	}

	role, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "role", RoleID: "role-villager"}, assets)
	if err != nil || role["role"].(map[string]any)["name"] != "Villager" {
		t.Fatalf("role preview = %#v, %v", role, err)
	}
	if media := role["media"].(map[string]any); !strings.Contains(media["preview"].(string), "ruleset-edit-assets") {
		t.Fatalf("role preview did not use staged media: %#v", media)
	}

	phases, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "phases", PhaseID: "phase-night"}, assets)
	if err != nil || len(phases["phases"].([]map[string]any)) != 1 {
		t.Fatalf("phase preview = %#v, %v", phases, err)
	}

	composition, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "composition", PlayerCount: 3}, assets)
	if err != nil || composition["feasible"] != true || len(composition["roles"].([]map[string]any)) != 1 {
		t.Fatalf("composition preview = %#v, %v", composition, err)
	}

	chat, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "chat", RoleID: "role-villager", PhaseID: "phase-night"}, assets)
	if err != nil {
		t.Fatal(err)
	}
	rooms := chat["rooms"].([]map[string]any)
	if rooms[0]["sendable"] != false || rooms[1]["messageRestriction"] != ChatEmojiOnly {
		t.Fatalf("chat preview did not apply live phase/channel policy: %#v", rooms)
	}

	media, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "media", AssetKey: "bell"}, assets)
	contexts := media["contexts"].([]map[string]any)
	if err != nil || media["media"].(map[string]any)["displayName"] != "Night bell" || len(contexts) != 1 || contexts[0]["kind"] != "phase" || contexts[0]["title"] != "Night" {
		t.Fatalf("media preview = %#v, %v", media, err)
	}
}

func TestMediaPreviewProjectsEveryConsumingContext(t *testing.T) {
	definition := DefinitionV1{
		Metadata:     Metadata{Name: "Context game", Description: "A contextual game.", MinPlayers: 4, MaxPlayers: 9, CoverAssetKey: "shared"},
		Teams:        []Team{{ID: "team", Name: "Village", Description: "The village side.", ImageAssetKey: "shared"}},
		Abilities:    []Ability{{ID: "ability", Name: "Inspect", Description: "Inspect a player.", ImageAssetKey: "shared"}},
		Roles:        []Role{{ID: "role", Name: "Seer", Description: "Find the threat.", TeamID: "team", WinCondition: "Keep the village safe.", ImageAssetKey: "shared"}},
		Achievements: []Achievement{{ID: "achievement", Name: "First sight", Description: "Inspect once.", Points: 2, ImageAssetKey: "shared"}},
	}
	record := core.NewRecord(&core.Collection{})
	record.Id = "image"
	record.Set("file", "shared.png")
	preview, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "media", AssetKey: "shared"}, map[string]effectiveAsset{
		"shared": {key: "shared", kind: "image", displayName: "Shared art", accessibilityText: "Shared artwork", record: record},
	})
	if err != nil {
		t.Fatal(err)
	}
	contexts := preview["contexts"].([]map[string]any)
	wantKinds := []string{"cover", "team", "role", "ability", "achievement"}
	if len(contexts) != len(wantKinds) {
		t.Fatalf("contexts = %#v", contexts)
	}
	for index, want := range wantKinds {
		if contexts[index]["kind"] != want {
			t.Fatalf("context %d kind = %v, want %s", index, contexts[index]["kind"], want)
		}
	}
}

func TestCompositionAndPhaseLessPreviewsExplainUnavailableConfiguration(t *testing.T) {
	definition := DefinitionV1{Metadata: Metadata{MinPlayers: 4, MaxPlayers: 8}}
	phases, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "phases"}, nil)
	if err != nil || phases["empty"] != true {
		t.Fatalf("phase-less preview = %#v, %v", phases, err)
	}
	composition, err := buildRulesetPreview(previewRequest{Definition: definition, Mode: "composition", PlayerCount: 6}, nil)
	if err != nil || composition["feasible"] != false || composition["message"] != "No player setup covers 6 players." {
		t.Fatalf("infeasible preview = %#v, %v", composition, err)
	}
}
