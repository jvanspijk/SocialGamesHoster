package abilities

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type PlayerChoice struct {
	ID          string `json:"id"`
	AbilityID   string `json:"abilityId"`
	AbilityName string `json:"abilityName"`
	Status      string `json:"status"`
	ActivatedAt any    `json:"activatedAt"`
	FinalizedAt any    `json:"finalizedAt,omitempty"`
}

func Activate(app core.App, gameID, profileID, abilityID string, now time.Time) ([]PlayerChoice, error) {
	var choices []PlayerChoice
	err := app.RunInTransaction(func(tx core.App) error {
		game, participant, definition, err := activationContext(tx, gameID, profileID)
		if err != nil {
			return err
		}
		if !game.GetDateTime("ability_phase_locked_at").IsZero() {
			return result.Conflict("ability.phase_locked", "Ability choices for this phase are finalized.")
		}
		ability, owned := ownedAbility(definition, participant.GetString("role_key"), abilityID)
		if !owned {
			return result.Forbidden("ability.not_owned", "That ability does not belong to your assigned role.")
		}
		if !contains(ability.ActivationPhaseIDs, game.GetString("phase_key")) {
			return result.Conflict("ability.phase_not_allowed", "That ability cannot be activated in this phase.")
		}
		existing, err := phaseChoices(tx, game, participant.Id)
		if err != nil {
			return err
		}
		for _, record := range existing {
			if record.GetString("ability_key") == ability.ID {
				choices, err = ProjectPlayerChoices(tx, game, participant, definition)
				return err
			}
			existingAbility := findAbility(definition, record.GetString("ability_key"))
			if !ability.CanCombineWithOtherAbilities || existingAbility == nil || !existingAbility.CanCombineWithOtherAbilities {
				return result.Conflict("ability.combination_conflict", "This ability cannot be combined with another active ability.")
			}
		}
		collection, err := tx.FindCollectionByNameOrId("ability_choices")
		if err != nil {
			return err
		}
		record := core.NewRecord(collection)
		record.Set("game", game.Id)
		record.Set("participant", participant.Id)
		record.Set("phase_key", game.GetString("phase_key"))
		record.Set("round_number", game.GetInt("round_number"))
		record.Set("phase_instance", game.GetInt("ability_phase_instance"))
		record.Set("ability_key", ability.ID)
		record.Set("status", "activated")
		record.Set("activated_at", now.UTC())
		if err := tx.Save(record); err != nil {
			return err
		}
		game.Set("revision", game.GetInt("revision")+1)
		if err := tx.Save(game); err != nil {
			return err
		}
		choices, err = ProjectPlayerChoices(tx, game, participant, definition)
		return err
	})
	return choices, err
}

func Undo(app core.App, gameID, profileID, abilityID string) ([]PlayerChoice, error) {
	var choices []PlayerChoice
	err := app.RunInTransaction(func(tx core.App) error {
		game, participant, definition, err := activationContext(tx, gameID, profileID)
		if err != nil {
			return err
		}
		if !game.GetDateTime("ability_phase_locked_at").IsZero() {
			return result.Conflict("ability.phase_locked", "Ability choices for this phase are finalized.")
		}
		if _, owned := ownedAbility(definition, participant.GetString("role_key"), abilityID); !owned {
			return result.Forbidden("ability.not_owned", "That ability does not belong to your assigned role.")
		}
		records, err := tx.FindRecordsByFilter(
			"ability_choices",
			"game = {:game} && participant = {:participant} && phase_instance = {:instance} && ability_key = {:ability}",
			"",
			1,
			0,
			dbx.Params{
				"game": game.Id, "participant": participant.Id,
				"instance": game.GetInt("ability_phase_instance"), "ability": abilityID,
			},
		)
		if err != nil {
			return err
		}
		if len(records) == 1 {
			if records[0].GetString("status") != "activated" {
				return result.Conflict("ability.phase_locked", "Ability choices for this phase are finalized.")
			}
			if err := tx.Delete(records[0]); err != nil {
				return err
			}
			game.Set("revision", game.GetInt("revision")+1)
			if err := tx.Save(game); err != nil {
				return err
			}
		}
		choices, err = ProjectPlayerChoices(tx, game, participant, definition)
		return err
	})
	return choices, err
}

func FinalizePhase(app core.App, game *core.Record, now time.Time) (bool, error) {
	if game.GetString("phase_key") == "" || !game.GetDateTime("ability_phase_locked_at").IsZero() {
		return false, nil
	}
	records, err := phaseChoices(app, game, "")
	if err != nil {
		return false, err
	}
	for _, record := range records {
		if record.GetString("status") == "activated" {
			record.Set("status", "finalized")
			record.Set("finalized_at", now.UTC())
			if err := app.Save(record); err != nil {
				return false, err
			}
		}
	}
	game.Set("ability_phase_locked_at", now.UTC())
	if err := app.Save(game); err != nil {
		return false, err
	}
	return true, nil
}

func ResetPhaseLock(game *core.Record) {
	game.Set("ability_phase_locked_at", nil)
}

func ProjectPlayer(app core.App, game, participant *core.Record, definition rulesets.DefinitionV1) ([]PlayerChoice, error) {
	return ProjectPlayerChoices(app, game, participant, definition)
}

func ProjectPlayerChoices(app core.App, game, participant *core.Record, definition rulesets.DefinitionV1) ([]PlayerChoice, error) {
	if game.GetString("phase_key") == "" {
		return []PlayerChoice{}, nil
	}
	records, err := phaseChoices(app, game, participant.Id)
	if err != nil {
		return nil, err
	}
	projected := make([]PlayerChoice, 0, len(records))
	for _, record := range records {
		ability := findAbility(definition, record.GetString("ability_key"))
		if ability == nil {
			continue
		}
		status := "Activated"
		if record.GetString("status") == "finalized" {
			status = "Finalized"
		}
		projected = append(projected, PlayerChoice{
			ID: record.Id, AbilityID: ability.ID, AbilityName: ability.Name, Status: status,
			ActivatedAt: dateValue(record, "activated_at"), FinalizedAt: dateValue(record, "finalized_at"),
		})
	}
	sort.Slice(projected, func(i, j int) bool { return projected[i].AbilityName < projected[j].AbilityName })
	return projected, nil
}

func ProjectAdmin(app core.App, game *core.Record, definition rulesets.DefinitionV1, participants []*core.Record) (map[string]any, []map[string]any, error) {
	eligible := 0
	for _, participant := range participants {
		if !gamepolicy.IsActivePlayer(gamepolicy.ParticipantStatus(participant.GetString("status"))) {
			continue
		}
		role := findRole(definition, participant.GetString("role_key"))
		if role == nil {
			continue
		}
		for _, abilityID := range role.AbilityIDs {
			ability := findAbility(definition, abilityID)
			if ability != nil && contains(ability.ActivationPhaseIDs, game.GetString("phase_key")) {
				eligible++
				break
			}
		}
	}
	records, err := phaseChoices(app, game, "")
	if err != nil {
		return nil, nil, err
	}
	activePlayers := map[string]bool{}
	finalizedPlayers := map[string]bool{}
	for _, record := range records {
		activePlayers[record.GetString("participant")] = true
		if record.GetString("status") == "finalized" {
			finalizedPlayers[record.GetString("participant")] = true
		}
	}
	lockedAt := dateValue(game, "ability_phase_locked_at")
	progress := map[string]any{
		"phaseKey": game.GetString("phase_key"), "roundNumber": game.GetInt("round_number"),
		"locked": lockedAt != nil, "lockedAt": lockedAt, "eligiblePlayerCount": eligible,
		"activatedPlayerCount": len(activePlayers), "finalizedPlayerCount": len(finalizedPlayers),
	}
	results := []map[string]any{}
	finalized, err := app.FindRecordsByFilter(
		"ability_choices",
		"game = {:game} && status = 'finalized'",
		"round_number,phase_key,participant,activated_at",
		10000,
		0,
		dbx.Params{"game": game.Id},
	)
	if err != nil {
		return nil, nil, err
	}
	type resultGroup struct {
		participantID string
		phaseKey      string
		round         int
		abilities     []map[string]any
	}
	byParticipantPhase := map[string]*resultGroup{}
	for _, record := range finalized {
		ability := findAbility(definition, record.GetString("ability_key"))
		if ability == nil {
			continue
		}
		key := fmt.Sprintf("%s\x00%s\x00%d", record.GetString("participant"), record.GetString("phase_key"), record.GetInt("round_number"))
		group := byParticipantPhase[key]
		if group == nil {
			group = &resultGroup{
				participantID: record.GetString("participant"),
				phaseKey:      record.GetString("phase_key"),
				round:         record.GetInt("round_number"),
			}
			byParticipantPhase[key] = group
		}
		group.abilities = append(group.abilities, map[string]any{"id": ability.ID, "name": ability.Name})
	}
	for _, participant := range participants {
		for _, group := range byParticipantPhase {
			if group.participantID != participant.Id {
				continue
			}
			results = append(results, map[string]any{
				"participantId": participant.Id, "displayName": participant.GetString("display_name_snapshot"),
				"seatNumber": participant.GetInt("seat_number"), "phaseKey": group.phaseKey,
				"roundNumber": group.round, "abilities": group.abilities,
			})
		}
	}
	return progress, results, nil
}

func activationContext(app core.App, gameID, profileID string) (*core.Record, *core.Record, rulesets.DefinitionV1, error) {
	game, err := app.FindRecordById("games", gameID)
	if err != nil {
		return nil, nil, rulesets.DefinitionV1{}, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound}
	}
	if game.GetString("status") != "running" && game.GetString("status") != "paused" {
		return nil, nil, rulesets.DefinitionV1{}, result.Conflict("ability.not_available", "Abilities are only available during a live phase.")
	}
	if game.GetString("phase_key") == "" {
		return nil, nil, rulesets.DefinitionV1{}, result.Conflict("ability.no_phase", "Choose a game phase before activating an ability.")
	}
	participants, err := app.FindRecordsByFilter(
		"participants",
		"game = {:game} && profile = {:profile} && status = 'active'",
		"",
		1,
		0,
		dbx.Params{"game": game.Id, "profile": profileID},
	)
	if err != nil {
		return nil, nil, rulesets.DefinitionV1{}, err
	}
	if len(participants) != 1 {
		return nil, nil, rulesets.DefinitionV1{}, result.Forbidden("ability.forbidden", "Abilities are not available to this player.")
	}
	var definition rulesets.DefinitionV1
	data, err := json.Marshal(game.Get("ruleset_snapshot"))
	if err != nil {
		return nil, nil, rulesets.DefinitionV1{}, result.Internal(err)
	}
	if err := json.Unmarshal(data, &definition); err != nil {
		return nil, nil, rulesets.DefinitionV1{}, result.Internal(err)
	}
	return game, participants[0], definition, nil
}

func phaseChoices(app core.App, game *core.Record, participantID string) ([]*core.Record, error) {
	filter := "game = {:game} && phase_instance = {:instance}"
	params := dbx.Params{"game": game.Id, "instance": game.GetInt("ability_phase_instance")}
	if participantID != "" {
		filter += " && participant = {:participant}"
		params["participant"] = participantID
	}
	return app.FindRecordsByFilter("ability_choices", filter, "activated_at,id", 1000, 0, params)
}

func ownedAbility(definition rulesets.DefinitionV1, roleID, abilityID string) (*rulesets.Ability, bool) {
	role := findRole(definition, roleID)
	if role == nil || !contains(role.AbilityIDs, abilityID) {
		return nil, false
	}
	ability := findAbility(definition, abilityID)
	return ability, ability != nil
}

func findRole(definition rulesets.DefinitionV1, id string) *rulesets.Role {
	for index := range definition.Roles {
		if definition.Roles[index].ID == id {
			return &definition.Roles[index]
		}
	}
	return nil
}

func findAbility(definition rulesets.DefinitionV1, id string) *rulesets.Ability {
	for index := range definition.Abilities {
		if definition.Abilities[index].ID == id {
			return &definition.Abilities[index]
		}
	}
	return nil
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func dateValue(record *core.Record, field string) any {
	value := record.GetDateTime(field)
	if value.IsZero() {
		return nil
	}
	return value.Time().UTC()
}
