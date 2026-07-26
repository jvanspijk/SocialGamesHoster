package rulesets

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"sort"
)

type Assignment struct {
	ParticipantID string `json:"participantId"`
	RoleID        string `json:"roleId"`
	Locked        bool   `json:"locked,omitempty"`
}

type AssignmentReport struct {
	Errors []string `json:"errors"`
}

func (r AssignmentReport) Valid() bool {
	return len(r.Errors) == 0
}

func ValidateAssignments(def DefinitionV1, participantCount int, assignments []Assignment) AssignmentReport {
	report := AssignmentReport{}
	if len(assignments) != participantCount {
		report.Errors = append(report.Errors, fmt.Sprintf("expected %d assignments, found %d", participantCount, len(assignments)))
		return report
	}
	band, ok := bandForPlayerCount(def.CompositionBands, participantCount)
	if !ok {
		report.Errors = append(report.Errors, fmt.Sprintf("no composition band covers %d players", participantCount))
		return report
	}

	roleByID := make(map[string]Role, len(def.Roles))
	counts := map[string]int{}
	for _, role := range def.Roles {
		roleByID[role.ID] = role
	}
	for _, assignment := range assignments {
		role, exists := roleByID[assignment.RoleID]
		if !exists {
			report.Errors = append(report.Errors, fmt.Sprintf("unknown role %q", assignment.RoleID))
			continue
		}
		counts[role.ID]++
		if counts[role.ID] > role.MaxCopies {
			report.Errors = append(report.Errors, fmt.Sprintf("%s exceeds its maximum of %d copies", role.Name, role.MaxCopies))
		}
	}
	if len(report.Errors) > 0 {
		return report
	}

	effectiveSlots := make([]CompositionSlot, len(band.Slots))
	copy(effectiveSlots, band.Slots)
	slotIndex := map[string]int{}
	for i, slot := range effectiveSlots {
		slotIndex[slot.ID] = i
	}
	for _, modifier := range def.CompositionModifiers {
		if counts[modifier.WhenRolePresent] == 0 {
			continue
		}
		for _, required := range modifier.RequiresRoleIDs {
			if counts[required] == 0 {
				report.Errors = append(report.Errors, fmt.Sprintf("role %q requires role %q", modifier.WhenRolePresent, required))
			}
		}
		for _, excluded := range modifier.ExcludesRoleIDs {
			if counts[excluded] > 0 {
				report.Errors = append(report.Errors, fmt.Sprintf("role %q excludes role %q", modifier.WhenRolePresent, excluded))
			}
		}
		for _, adjustment := range modifier.SlotAdjustments {
			index, exists := slotIndex[adjustment.SlotID]
			if !exists {
				report.Errors = append(report.Errors, fmt.Sprintf("modifier %q references unknown slot %q", modifier.ID, adjustment.SlotID))
				continue
			}
			effectiveSlots[index].Count += adjustment.Delta
			if effectiveSlots[index].Count < 0 {
				report.Errors = append(report.Errors, fmt.Sprintf("modifier %q makes slot %q negative", modifier.ID, adjustment.SlotID))
			}
		}
	}
	if len(report.Errors) > 0 {
		return report
	}

	total := 0
	for _, slot := range effectiveSlots {
		total += slot.Count
	}
	if total != participantCount {
		report.Errors = append(report.Errors, fmt.Sprintf("effective slot total is %d, expected %d", total, participantCount))
		return report
	}

	roleIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		roleIDs = append(roleIDs, assignment.RoleID)
	}
	if !rolesFitSlots(roleIDs, effectiveSlots, roleByID) {
		report.Errors = append(report.Errors, "assigned roles do not satisfy the composition slots")
	}
	return report
}

func RandomizeAssignments(def DefinitionV1, participantIDs []string, locked []Assignment, seed uint64) ([]Assignment, error) {
	if len(participantIDs) == 0 || len(participantIDs) > 30 {
		return nil, fmt.Errorf("participant count must be between 1 and 30")
	}
	band, ok := bandForPlayerCount(def.CompositionBands, len(participantIDs))
	if !ok {
		return nil, fmt.Errorf("no composition band covers %d players", len(participantIDs))
	}
	roleByID := make(map[string]Role, len(def.Roles))
	for _, role := range def.Roles {
		roleByID[role.ID] = role
	}
	lockedByParticipant := map[string]string{}
	lockedRoleCounts := map[string]int{}
	participantSet := map[string]struct{}{}
	for _, id := range participantIDs {
		participantSet[id] = struct{}{}
	}
	for _, assignment := range locked {
		if _, ok := participantSet[assignment.ParticipantID]; !ok {
			return nil, fmt.Errorf("locked participant %q is not in the roster", assignment.ParticipantID)
		}
		role, ok := roleByID[assignment.RoleID]
		if !ok {
			return nil, fmt.Errorf("locked role %q does not exist", assignment.RoleID)
		}
		if _, duplicate := lockedByParticipant[assignment.ParticipantID]; duplicate {
			return nil, fmt.Errorf("participant %q is locked more than once", assignment.ParticipantID)
		}
		lockedByParticipant[assignment.ParticipantID] = assignment.RoleID
		lockedRoleCounts[assignment.RoleID]++
		if lockedRoleCounts[assignment.RoleID] > role.MaxCopies {
			return nil, fmt.Errorf("locked assignments exceed %s's maximum copies", role.Name)
		}
	}

	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	slotPositions := make([]CompositionSlot, 0, len(participantIDs))
	for _, slot := range band.Slots {
		for range slot.Count {
			slotPositions = append(slotPositions, slot)
		}
	}
	if len(slotPositions) != len(participantIDs) {
		return nil, fmt.Errorf("base composition slots total %d, expected %d", len(slotPositions), len(participantIDs))
	}
	rng.Shuffle(len(slotPositions), func(i, j int) {
		slotPositions[i], slotPositions[j] = slotPositions[j], slotPositions[i]
	})

	roleIDs := make([]string, len(slotPositions))
	counts := map[string]int{}
	var solve func(int) bool
	solve = func(index int) bool {
		if index == len(slotPositions) {
			candidate := make([]Assignment, len(roleIDs))
			for i, roleID := range roleIDs {
				candidate[i] = Assignment{ParticipantID: fmt.Sprintf("seat-%d", i+1), RoleID: roleID}
			}
			if !ValidateAssignments(def, len(roleIDs), candidate).Valid() {
				return false
			}
			for roleID, required := range lockedRoleCounts {
				if counts[roleID] < required {
					return false
				}
			}
			return true
		}
		candidates := MatchingRoles(def.Roles, slotPositions[index].Selector)
		rng.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
		for _, role := range candidates {
			if counts[role.ID] >= role.MaxCopies {
				continue
			}
			roleIDs[index] = role.ID
			counts[role.ID]++
			if solve(index + 1) {
				return true
			}
			counts[role.ID]--
		}
		return false
	}
	if !solve(0) {
		return nil, fmt.Errorf("no valid role assignment exists for %d players with the selected locks", len(participantIDs))
	}

	available := slices.Clone(roleIDs)
	result := make([]Assignment, 0, len(participantIDs))
	for _, participantID := range participantIDs {
		if roleID, ok := lockedByParticipant[participantID]; ok {
			index := slices.Index(available, roleID)
			if index < 0 {
				return nil, fmt.Errorf("solver could not satisfy locked role %q", roleID)
			}
			available = slices.Delete(available, index, index+1)
			result = append(result, Assignment{ParticipantID: participantID, RoleID: roleID, Locked: true})
		}
	}
	rng.Shuffle(len(available), func(i, j int) { available[i], available[j] = available[j], available[i] })
	for _, participantID := range participantIDs {
		if _, locked := lockedByParticipant[participantID]; locked {
			continue
		}
		result = append(result, Assignment{ParticipantID: participantID, RoleID: available[0]})
		available = available[1:]
	}
	sort.Slice(result, func(i, j int) bool {
		return slices.Index(participantIDs, result[i].ParticipantID) < slices.Index(participantIDs, result[j].ParticipantID)
	})
	return result, nil
}

func rolesFitSlots(roleIDs []string, slots []CompositionSlot, roleByID map[string]Role) bool {
	positions := make([]CompositionSlot, 0, len(roleIDs))
	for _, slot := range slots {
		for range slot.Count {
			positions = append(positions, slot)
		}
	}
	used := make([]bool, len(roleIDs))
	sort.Slice(positions, func(i, j int) bool {
		return len(MatchingRoles(mapRoles(roleByID), positions[i].Selector)) < len(MatchingRoles(mapRoles(roleByID), positions[j].Selector))
	})
	var match func(int) bool
	match = func(position int) bool {
		if position == len(positions) {
			return true
		}
		for i, roleID := range roleIDs {
			if used[i] || !selectorMatches(roleByID[roleID], positions[position].Selector) {
				continue
			}
			used[i] = true
			if match(position + 1) {
				return true
			}
			used[i] = false
		}
		return false
	}
	return len(positions) == len(roleIDs) && match(0)
}

func mapRoles(roleByID map[string]Role) []Role {
	roles := make([]Role, 0, len(roleByID))
	for _, role := range roleByID {
		roles = append(roles, role)
	}
	return roles
}

func bandForPlayerCount(bands []CompositionBand, count int) (CompositionBand, bool) {
	for _, band := range bands {
		if count >= band.MinPlayers && count <= band.MaxPlayers {
			return band, true
		}
	}
	return CompositionBand{}, false
}
