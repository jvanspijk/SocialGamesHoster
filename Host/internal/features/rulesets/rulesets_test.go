package rulesets

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

func TestSelectorUsesOrWithinAndAcrossFields(t *testing.T) {
	roles := []Role{
		{ID: "sheriff", TeamID: "town", CategoryIDs: []string{"investigative"}, Tags: []string{"night_action"}},
		{ID: "doctor", TeamID: "town", CategoryIDs: []string{"protective"}, Tags: []string{"night_action"}},
		{ID: "mafioso", TeamID: "mafia", CategoryIDs: []string{"killing"}, Tags: []string{"night_action"}},
	}
	matches := MatchingRoles(roles, Selector{TeamIDs: []string{"town"}, CategoryIDs: []string{"investigative", "support"}})
	if len(matches) != 1 || matches[0].ID != "sheriff" {
		t.Fatalf("unexpected matches: %#v", matches)
	}
}

func TestDecodeSnapshot(t *testing.T) {
	want := testDefinition()
	got, err := DecodeSnapshot(want)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded definition = %#v, error = %v", got, err)
	}
	got, err = DecodeSnapshot(map[string]any{"schemaVersion": 1})
	if err != nil || got.SchemaVersion != 1 {
		t.Fatalf("decoded stored snapshot = %#v, error = %v", got, err)
	}
	for _, snapshot := range []any{make(chan int), map[string]any{"schemaVersion": "one"}} {
		if _, err := DecodeSnapshot(snapshot); err == nil {
			t.Fatal("expected decoding error")
		}
	}
}

func TestAchievementContractPreservesPointsAndSpoilerVisibility(t *testing.T) {
	definition := testDefinition()
	definition.Achievements = []Achievement{{
		ID: "secret_win", Name: "Secret win", Description: "Revealed after the game.",
		Points: 75, HiddenUntilGameCompleted: true,
	}}
	encoded, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DefinitionV1
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	achievement := decoded.Achievements[0]
	if achievement.Points != 75 || !achievement.HiddenUntilGameCompleted {
		t.Fatalf("achievement contract did not round-trip: %+v", achievement)
	}
	if report := Validate(decoded, map[string]struct{}{}); !report.Valid() {
		t.Fatalf("valid achievement rejected: %+v", report.Errors)
	}
}

func TestAbilityActivationAndAssetAccessibilityRoundTrip(t *testing.T) {
	definition := testDefinition()
	definition.Phases = []Phase{{ID: "night", Name: "Night", Order: 1}}
	definition.Abilities = []Ability{{ID: "inspect", Name: "Inspect", Description: "Inspect one player."}}
	definition.Abilities[0].ActivationPhaseIDs = []string{"night"}
	definition.Abilities[0].CanCombineWithOtherAbilities = true
	definition.AssetAccessibility = map[string]AssetAccessibility{
		"ability_art": {Description: "A silver eye against a dark sky."},
	}
	definition.Abilities[0].ImageAssetKey = "ability_art"
	encoded, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DefinitionV1
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Abilities[0].CanCombineWithOtherAbilities ||
		!reflect.DeepEqual(decoded.Abilities[0].ActivationPhaseIDs, []string{"night"}) ||
		decoded.AssetAccessibility["ability_art"].Description == "" {
		t.Fatalf("activation/accessibility contract did not round-trip: %#v", decoded)
	}
	if report := Validate(decoded, map[string]struct{}{"ability_art": {}}); !report.Valid() {
		t.Fatalf("valid activation/accessibility rejected: %+v", report.Errors)
	}
}

func TestAchievementPointsAndAutomaticAudioAudienceValidation(t *testing.T) {
	definition := testDefinition()
	definition.Achievements = []Achievement{{
		ID: "too_many", Name: "Too many", Points: 10001,
	}}
	definition.AudioCues = []AudioCue{{
		ID: "private_cue", Name: "Private cue", AssetKey: "cue_audio", DefaultAudience: "player",
	}}
	definition.Phases = []Phase{{
		ID: "night", Name: "Night", Order: 1, AudioCueID: "private_cue",
	}}
	report := Validate(definition, map[string]struct{}{"cue_audio": {}})
	codes := make([]string, len(report.Errors))
	for index, issue := range report.Errors {
		codes[index] = issue.Code
	}
	if !slices.Contains(codes, "achievement.invalid_points") {
		t.Fatalf("missing achievement point validation: %+v", report.Errors)
	}
	if !slices.Contains(codes, "audio.target_required") {
		t.Fatalf("missing automatic audio audience validation: %+v", report.Errors)
	}
}

func TestCustomChatChannelValidatesAndRoundTrips(t *testing.T) {
	definition := testDefinition()
	hidden := false
	definition.Phases = []Phase{{ID: "night", Name: "Night", Order: 1}}
	definition.Chat.Channels = []ChatChannel{{
		ID: "mafia_signals", Name: "Mafia signals",
		ReaderTeamIDs: []string{"mafia"}, SenderRoleIDs: []string{"mafioso"},
		MessageRestriction: ChatEmojiOnly, Visible: true, Sendable: true,
		GameMasterMaySend: true, SenderDisplay: SenderRoleLabel,
		PhaseOverrides: map[string]ChatChannelPhaseOverride{"night": {Visible: &hidden}},
	}}
	if report := Validate(definition, map[string]struct{}{}); !report.Valid() {
		t.Fatalf("valid custom channel rejected: %+v", report.Errors)
	}
	encoded, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DefinitionV1
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	channel := decoded.Chat.Channels[0]
	if channel.MessageRestriction != ChatEmojiOnly || channel.PhaseOverrides["night"].Visible == nil {
		t.Fatalf("custom channel did not round-trip: %+v", channel)
	}
}

func TestCustomChatChannelRejectsUnknownAndWriteOnlyAudiences(t *testing.T) {
	definition := testDefinition()
	definition.Chat.Channels = []ChatChannel{{
		ID: "town_only", Name: "Town only",
		ReaderTeamIDs: []string{"town"}, SenderTeamIDs: []string{"mafia", "unknown"},
		MessageRestriction: ChatNormalText, Visible: true, Sendable: true,
		SenderDisplay: SenderGameAlias, PhaseOverrides: map[string]ChatChannelPhaseOverride{},
	}}
	report := Validate(definition, map[string]struct{}{})
	codes := make([]string, len(report.Errors))
	for index, issue := range report.Errors {
		codes[index] = issue.Code
	}
	if !slices.Contains(codes, "reference.unknown") || !slices.Contains(codes, "chat.sender_cannot_read") {
		t.Fatalf("missing custom channel validation errors: %+v", report.Errors)
	}
}

func TestCustomChatChannelEmptySenderAudienceMeansEveryReader(t *testing.T) {
	channel := ChatChannel{ReaderTeamIDs: []string{"town"}}
	if !ChatChannelAudienceMatches(channel, Role{ID: "doctor", TeamID: "town"}, true) {
		t.Fatal("reader was not allowed to send with the default sender audience")
	}
	if ChatChannelAudienceMatches(channel, Role{ID: "mafioso", TeamID: "mafia"}, true) {
		t.Fatal("non-reader was allowed to send with the default sender audience")
	}
}

func TestNormalizeDefinitionIdentifiersPreservesImportedValuesAndFillsMissingValues(t *testing.T) {
	definition := DefinitionV1{
		Teams: []Team{{ID: "existing_team"}, {}},
		Roles: []Role{{}},
		Chat:  ChatPolicy{Channels: []ChatChannel{{}}},
	}
	normalized, err := normalizeDefinitionIdentifiers(definition)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Teams[0].ID != "existing_team" {
		t.Fatalf("imported identifier changed: %q", normalized.Teams[0].ID)
	}
	for _, value := range []string{
		normalized.Teams[1].ID,
		normalized.Roles[0].ID,
		normalized.Chat.Channels[0].ID,
	} {
		if !stableIDPattern.MatchString(value) {
			t.Fatalf("generated identifier is not valid: %q", value)
		}
	}
}

func testDefinition() DefinitionV1 {
	return DefinitionV1{
		SchemaVersion: 1,
		Metadata:      Metadata{Name: "Test", MinPlayers: 4, MaxPlayers: 4},
		Teams: []Team{
			{ID: "town", Name: "Town"},
			{ID: "mafia", Name: "Mafia"},
		},
		Categories: []Category{
			{ID: "town_any", Name: "Town"},
			{ID: "mafia_any", Name: "Mafia"},
		},
		Roles: []Role{
			{ID: "villager", Name: "Villager", TeamID: "town", CategoryIDs: []string{"town_any"}, MaxCopies: 3},
			{ID: "doctor", Name: "Doctor", TeamID: "town", CategoryIDs: []string{"town_any"}, MaxCopies: 1},
			{ID: "mafioso", Name: "Mafioso", TeamID: "mafia", CategoryIDs: []string{"mafia_any"}, MaxCopies: 1},
		},
		Chat: ChatPolicy{
			DefaultPolicy:  ChatPolicyDefaults{Teams: map[string]RoomPermission{}},
			PhaseOverrides: map[string]ChatPolicyOverride{},
		},
	}
}
