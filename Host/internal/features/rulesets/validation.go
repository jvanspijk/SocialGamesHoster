package rulesets

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
)

var stableIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

type ValidationIssue struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationReport struct {
	Errors   []ValidationIssue `json:"errors"`
	Warnings []ValidationIssue `json:"warnings"`
}

func (r ValidationReport) Valid() bool {
	return len(r.Errors) == 0
}

func Validate(def DefinitionV1, assetKeys map[string]struct{}) ValidationReport {
	report := ValidationReport{}
	addError := func(path, code, message string) {
		report.Errors = append(report.Errors, ValidationIssue{Path: path, Code: code, Message: message})
	}
	addWarning := func(path, code, message string) {
		report.Warnings = append(report.Warnings, ValidationIssue{Path: path, Code: code, Message: message})
	}

	if def.SchemaVersion != 1 {
		addError("schemaVersion", "schema.unsupported", "Schema version must be 1.")
	}
	if strings.TrimSpace(def.Metadata.Name) == "" {
		addError("metadata.name", "required", "Name is required.")
	}
	if def.Metadata.MinPlayers < 1 || def.Metadata.MaxPlayers > 30 || def.Metadata.MinPlayers > def.Metadata.MaxPlayers {
		addError("metadata", "players.invalid_range", "Player range must be between 1 and 30.")
	}

	teamIDs := validateIDs("teams", teamIDValues(def.Teams), addError)
	categoryIDs := validateIDs("categories", categoryIDValues(def.Categories), addError)
	abilityIDs := validateIDs("abilities", abilityIDValues(def.Abilities), addError)
	roleIDs := validateIDs("roles", roleIDValues(def.Roles), addError)
	if len(def.Teams) == 0 {
		addError("teams", "teams.required", "Add at least one team.")
	}
	if len(def.Roles) == 0 {
		addError("roles", "roles.required", "Add at least one role.")
	}
	phaseIDs := validateIDs("phases", phaseIDValues(def.Phases), addError)
	achievementIDs := validateIDs("achievements", achievementIDValues(def.Achievements), addError)
	audioCueIDs := validateIDs("audioCues", audioCueIDValues(def.AudioCues), addError)
	audioCueAudiences := make(map[string]string, len(def.AudioCues))
	for _, cue := range def.AudioCues {
		audioCueAudiences[cue.ID] = cue.DefaultAudience
	}
	_ = achievementIDs

	checkAsset := func(path, key string) {
		if key == "" {
			return
		}
		if _, ok := assetKeys[key]; !ok {
			addError(path, "asset.unknown", "Choose an existing asset.")
		}
	}
	checkAsset("metadata.coverAssetKey", def.Metadata.CoverAssetKey)
	for i, team := range def.Teams {
		checkAsset(fmt.Sprintf("teams[%d].imageAssetKey", i), team.ImageAssetKey)
	}
	for i, ability := range def.Abilities {
		checkAsset(fmt.Sprintf("abilities[%d].imageAssetKey", i), ability.ImageAssetKey)
		for phaseIndex, phaseID := range ability.ActivationPhaseIDs {
			if _, ok := phaseIDs[phaseID]; !ok {
				addError(
					fmt.Sprintf("abilities[%d].activationPhaseIds[%d]", i, phaseIndex),
					"reference.unknown",
					"Choose an existing phase.",
				)
			}
		}
	}
	for assetKey, accessibility := range def.AssetAccessibility {
		if _, ok := assetKeys[assetKey]; !ok {
			addError("assetAccessibility."+assetKey, "reference.unknown", "Choose an existing asset.")
		}
		description := strings.TrimSpace(accessibility.Description)
		if description == "" || len([]rune(description)) > 1000 {
			addError("assetAccessibility."+assetKey+".description", "asset.invalid_accessibility_description", "Enter an accessibility description of at most 1000 characters.")
		}
	}
	for i, role := range def.Roles {
		path := fmt.Sprintf("roles[%d]", i)
		if _, ok := teamIDs[role.TeamID]; !ok {
			addError(path+".teamId", "reference.unknown", "Role references an unknown team.")
		}
		for _, id := range role.CategoryIDs {
			if _, ok := categoryIDs[id]; !ok {
				addError(path+".categoryIds", "reference.unknown", "Choose an existing category.")
			}
		}
		for _, id := range role.AbilityIDs {
			if _, ok := abilityIDs[id]; !ok {
				addError(path+".abilityIds", "reference.unknown", "Choose an existing ability.")
			}
		}
		if role.MaxCopies < 1 || role.MaxCopies > 30 {
			addError(path+".maxCopies", "role.invalid_max_copies", "Maximum copies must be between 1 and 30.")
		}
		checkAsset(path+".imageAssetKey", role.ImageAssetKey)
	}

	phaseOrders := map[int]struct{}{}
	for i, phase := range def.Phases {
		path := fmt.Sprintf("phases[%d]", i)
		if _, exists := phaseOrders[phase.Order]; exists {
			addError(path+".order", "phase.duplicate_order", "Phase order values must be unique.")
		}
		phaseOrders[phase.Order] = struct{}{}
		if phase.SuggestedDurationSeconds < 0 {
			addError(path+".suggestedDurationSeconds", "phase.invalid_duration", "Suggested duration cannot be negative.")
		}
		if phase.SuggestedDurationSeconds == 0 {
			addWarning(path+".suggestedDurationSeconds", "phase.no_duration", "This phase has no suggested duration.")
		}
		if phase.AudioCueID != "" {
			if _, ok := audioCueIDs[phase.AudioCueID]; !ok {
				addError(path+".audioCueId", "reference.unknown", "Phase references an unknown audio cue.")
			} else if audience := audioCueAudiences[phase.AudioCueID]; audience == "team" || audience == "player" {
				addError(path+".audioCueId", "audio.target_required", "Automatic phase sounds must target all players or game masters.")
			}
		}
	}

	for i, rule := range def.KnowledgeRules {
		validateSelector(fmt.Sprintf("knowledgeRules[%d].viewer", i), rule.Viewer, roleIDs, teamIDs, categoryIDs, def.Roles, addError)
		validateSelector(fmt.Sprintf("knowledgeRules[%d].target", i), rule.Target, roleIDs, teamIDs, categoryIDs, def.Roles, addError)
		for _, reveal := range rule.Reveal {
			if !slices.Contains([]string{"identity", "role", "team", "elimination_state"}, reveal) {
				addError(fmt.Sprintf("knowledgeRules[%d].reveal", i), "knowledge.invalid_reveal", fmt.Sprintf("Unknown reveal field %q.", reveal))
			}
		}
	}

	sortedBands := slices.Clone(def.CompositionBands)
	sort.Slice(sortedBands, func(i, j int) bool { return sortedBands[i].MinPlayers < sortedBands[j].MinPlayers })
	for i, band := range sortedBands {
		path := fmt.Sprintf("compositionBands[%d]", i)
		if band.MinPlayers < def.Metadata.MinPlayers || band.MaxPlayers > def.Metadata.MaxPlayers || band.MinPlayers > band.MaxPlayers {
			addError(path, "band.invalid_range", "Composition band is outside the ruleset player range.")
		}
		if i > 0 && band.MinPlayers <= sortedBands[i-1].MaxPlayers {
			addError(path, "band.overlap", "Composition bands cannot overlap.")
		}
		slotIDs := validateIDs(path+".slots", slotIDValues(band.Slots), addError)
		_ = slotIDs
		for slotIndex, slot := range band.Slots {
			slotPath := fmt.Sprintf("%s.slots[%d]", path, slotIndex)
			if slot.Count < 0 {
				addError(slotPath+".count", "slot.negative", "Slot count cannot be negative.")
			}
			validateSelector(slotPath+".selector", slot.Selector, roleIDs, teamIDs, categoryIDs, def.Roles, addError)
			if len(MatchingRoles(def.Roles, slot.Selector)) == 0 {
				addError(slotPath+".selector", "selector.empty", "Slot selector does not match any role.")
			}
		}
	}

	if !bandsCoverRange(sortedBands, def.Metadata.MinPlayers, def.Metadata.MaxPlayers) {
		addError("compositionBands", "band.missing_coverage", "Composition bands must cover every supported player count.")
	}

	allSlotIDs := map[string]struct{}{}
	for _, band := range def.CompositionBands {
		for _, slot := range band.Slots {
			allSlotIDs[slot.ID] = struct{}{}
		}
	}
	validateIDs("compositionModifiers", modifierIDValues(def.CompositionModifiers), addError)
	for i, modifier := range def.CompositionModifiers {
		path := fmt.Sprintf("compositionModifiers[%d]", i)
		if _, ok := roleIDs[modifier.WhenRolePresent]; !ok {
			addError(path+".whenRolePresent", "reference.unknown", "Modifier trigger role does not exist.")
		}
		for _, id := range append(slices.Clone(modifier.RequiresRoleIDs), modifier.ExcludesRoleIDs...) {
			if _, ok := roleIDs[id]; !ok {
				addError(path, "reference.unknown", "Choose an existing role.")
			}
		}
		for adjustmentIndex, adjustment := range modifier.SlotAdjustments {
			if _, ok := allSlotIDs[adjustment.SlotID]; !ok {
				addError(fmt.Sprintf("%s.slotAdjustments[%d]", path, adjustmentIndex), "reference.unknown", "Modifier references an unknown slot.")
			}
		}
	}

	for teamID := range def.Chat.DefaultPolicy.Teams {
		if _, ok := teamIDs[teamID]; !ok {
			addError("chat.defaultPolicy.teams", "reference.unknown", "Choose an existing team.")
		}
	}
	for phaseID, override := range def.Chat.PhaseOverrides {
		if _, ok := phaseIDs[phaseID]; !ok {
			addError("chat.phaseOverrides", "reference.unknown", "Choose an existing phase.")
		}
		for teamID := range override.Teams {
			if _, ok := teamIDs[teamID]; !ok {
				addError("chat.phaseOverrides."+phaseID, "reference.unknown", "Choose an existing team.")
			}
		}
	}
	validateIDs("chat.channels", chatChannelIDValues(def.Chat.Channels), addError)
	for index, channel := range def.Chat.Channels {
		path := fmt.Sprintf("chat.channels[%d]", index)
		if strings.TrimSpace(channel.Name) == "" {
			addError(path+".name", "chat.channel_name_required", "Chat channel name is required.")
		}
		if channel.MessageRestriction != ChatNormalText && channel.MessageRestriction != ChatEmojiOnly {
			addError(path+".messageRestriction", "chat.invalid_message_restriction", "Choose normal text or emoji-only messages.")
		}
		if !slices.Contains([]SenderDisplay{
			SenderProfileName, SenderGameAlias, SenderSeatNumber, SenderRoleLabel, SenderTeamLabel,
		}, channel.SenderDisplay) {
			addError(path+".senderDisplay", "chat.invalid_sender_display", "Choose how player senders are shown.")
		}
		validateChatAudienceReferences(path+".readers", channel.ReaderRoleIDs, channel.ReaderTeamIDs, roleIDs, teamIDs, addError)
		validateChatAudienceReferences(path+".senders", channel.SenderRoleIDs, channel.SenderTeamIDs, roleIDs, teamIDs, addError)
		for phaseID := range channel.PhaseOverrides {
			if _, ok := phaseIDs[phaseID]; !ok {
				addError(path+".phaseOverrides", "reference.unknown", "Choose an existing phase.")
			}
		}
		for _, role := range def.Roles {
			if ChatChannelAudienceMatches(channel, role, true) && !ChatChannelAudienceMatches(channel, role, false) {
				addError(path+".senders", "chat.sender_cannot_read", fmt.Sprintf("%s can send but cannot read this channel.", role.Name))
				break
			}
		}
	}
	for i, cue := range def.AudioCues {
		checkAsset(fmt.Sprintf("audioCues[%d].assetKey", i), cue.AssetKey)
		if !slices.Contains([]string{"all", "team", "player", "game_masters"}, cue.DefaultAudience) {
			addError(fmt.Sprintf("audioCues[%d].defaultAudience", i), "audio.invalid_audience", "Audio cue audience is invalid.")
		}
	}
	for i, achievement := range def.Achievements {
		checkAsset(fmt.Sprintf("achievements[%d].imageAssetKey", i), achievement.ImageAssetKey)
		if achievement.Points < 0 || achievement.Points > 10000 {
			addError(fmt.Sprintf("achievements[%d].points", i), "achievement.invalid_points", "Achievement points must be between 0 and 10,000.")
		}
	}

	return report
}

func validateChatAudienceReferences(
	path string,
	roleValues, teamValues []string,
	roleIDs, teamIDs map[string]struct{},
	addError func(string, string, string),
) {
	for _, id := range roleValues {
		if _, ok := roleIDs[id]; !ok {
			addError(path+".roleIds", "reference.unknown", "Choose an existing role.")
		}
	}
	for _, id := range teamValues {
		if _, ok := teamIDs[id]; !ok {
			addError(path+".teamIds", "reference.unknown", "Choose an existing team.")
		}
	}
}

func validateIDs(path string, values []string, addError func(string, string, string)) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for i, value := range values {
		itemPath := fmt.Sprintf("%s[%d].id", path, i)
		if !stableIDPattern.MatchString(value) {
			addError(itemPath, "id.invalid", "This item cannot be used. Create it again.")
		}
		if _, exists := result[value]; exists {
			addError(itemPath, "id.duplicate", "Two items conflict. Create one of them again.")
		}
		result[value] = struct{}{}
	}
	return result
}

func validateSelector(path string, selector Selector, roleIDs, teamIDs, categoryIDs map[string]struct{}, roles []Role, addError func(string, string, string)) {
	for _, id := range selector.RoleIDs {
		if _, ok := roleIDs[id]; !ok {
			addError(path+".roleIds", "reference.unknown", "Choose an existing role.")
		}
	}
	for _, id := range selector.TeamIDs {
		if _, ok := teamIDs[id]; !ok {
			addError(path+".teamIds", "reference.unknown", "Choose an existing team.")
		}
	}
	for _, id := range selector.CategoryIDs {
		if _, ok := categoryIDs[id]; !ok {
			addError(path+".categoryIds", "reference.unknown", "Choose an existing category.")
		}
	}
	if len(MatchingRoles(roles, selector)) == 0 {
		addError(path, "selector.empty", "Selector does not match any role.")
	}
}

func bandsCoverRange(bands []CompositionBand, minPlayers, maxPlayers int) bool {
	for playerCount := minPlayers; playerCount <= maxPlayers; playerCount++ {
		found := false
		for _, band := range bands {
			if playerCount >= band.MinPlayers && playerCount <= band.MaxPlayers {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func MatchingRoles(roles []Role, selector Selector) []Role {
	matches := make([]Role, 0)
	for _, role := range roles {
		if selectorMatches(role, selector) {
			matches = append(matches, role)
		}
	}
	return matches
}

func selectorMatches(role Role, selector Selector) bool {
	if len(selector.RoleIDs) > 0 && !slices.Contains(selector.RoleIDs, role.ID) {
		return false
	}
	if len(selector.TeamIDs) > 0 && !slices.Contains(selector.TeamIDs, role.TeamID) {
		return false
	}
	if len(selector.CategoryIDs) > 0 && !intersects(selector.CategoryIDs, role.CategoryIDs) {
		return false
	}
	if len(selector.Tags) > 0 && !intersects(selector.Tags, role.Tags) {
		return false
	}
	return true
}

func intersects(left, right []string) bool {
	for _, value := range left {
		if slices.Contains(right, value) {
			return true
		}
	}
	return false
}

func teamIDValues(values []Team) []string {
	return mapIDs(values, func(value Team) string { return value.ID })
}
func categoryIDValues(values []Category) []string {
	return mapIDs(values, func(value Category) string { return value.ID })
}
func abilityIDValues(values []Ability) []string {
	return mapIDs(values, func(value Ability) string { return value.ID })
}
func roleIDValues(values []Role) []string {
	return mapIDs(values, func(value Role) string { return value.ID })
}
func phaseIDValues(values []Phase) []string {
	return mapIDs(values, func(value Phase) string { return value.ID })
}

func chatChannelIDValues(values []ChatChannel) []string {
	result := make([]string, len(values))
	for index := range values {
		result[index] = values[index].ID
	}
	return result
}
func achievementIDValues(values []Achievement) []string {
	return mapIDs(values, func(value Achievement) string { return value.ID })
}
func audioCueIDValues(values []AudioCue) []string {
	return mapIDs(values, func(value AudioCue) string { return value.ID })
}
func slotIDValues(values []CompositionSlot) []string {
	return mapIDs(values, func(value CompositionSlot) string { return value.ID })
}
func modifierIDValues(values []CompositionModifier) []string {
	return mapIDs(values, func(value CompositionModifier) string { return value.ID })
}

func mapIDs[T any](values []T, getID func(T) string) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = getID(value)
	}
	return result
}
