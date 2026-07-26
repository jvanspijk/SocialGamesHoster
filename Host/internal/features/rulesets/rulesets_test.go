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

func TestRandomAssignmentIsDeterministicAndRespectsLocks(t *testing.T) {
	def := testDefinition()
	participants := []string{"a", "b", "c", "d"}
	locked := []Assignment{{ParticipantID: "a", RoleID: "mafioso", Locked: true}}

	first, err := RandomizeAssignments(def, participants, locked, 42)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RandomizeAssignments(def, participants, locked, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("seed was not deterministic:\n%#v\n%#v", first, second)
	}
	if first[0].RoleID != "mafioso" || !first[0].Locked {
		t.Fatalf("lock was not preserved: %#v", first[0])
	}
	if report := ValidateAssignments(def, len(participants), first); !report.Valid() {
		t.Fatalf("invalid assignment: %#v", report.Errors)
	}
}

func TestModifierRejectsExcludedRole(t *testing.T) {
	def := testDefinition()
	def.CompositionModifiers = []CompositionModifier{{
		ID:              "boss_rule",
		WhenRolePresent: "mafioso",
		ExcludesRoleIDs: []string{"doctor"},
		SlotAdjustments: []SlotAdjustment{},
	}}
	assignments := []Assignment{
		{ParticipantID: "a", RoleID: "mafioso"},
		{ParticipantID: "b", RoleID: "doctor"},
		{ParticipantID: "c", RoleID: "villager"},
		{ParticipantID: "d", RoleID: "villager"},
	}
	if ValidateAssignments(def, 4, assignments).Valid() {
		t.Fatal("expected excluded role combination to fail")
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
		CompositionBands: []CompositionBand{{
			ID:         "four_players",
			MinPlayers: 4,
			MaxPlayers: 4,
			Slots: []CompositionSlot{
				{ID: "mafia_slot", Label: "Mafia", Count: 1, Selector: Selector{TeamIDs: []string{"mafia"}}},
				{ID: "town_slot", Label: "Town", Count: 3, Selector: Selector{TeamIDs: []string{"town"}}},
			},
		}},
		Chat: ChatPolicy{
			DefaultPolicy:  ChatPolicyDefaults{Teams: map[string]RoomPermission{}},
			PhaseOverrides: map[string]ChatPolicyOverride{},
		},
	}
}
