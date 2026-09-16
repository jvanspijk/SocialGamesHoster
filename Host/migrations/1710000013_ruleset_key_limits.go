package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
)

// Ruleset editor identifiers are prefix + UUID values, which can be up to 48
// characters. Runtime records retain several of those identifiers, so their
// fields must accommodate the same contract.
func init() {
	pbmigrations.Register(func(app core.App) error {
		return setRulesetKeyLimits(app, 64)
	}, func(app core.App) error {
		return setRulesetKeyLimits(app, 32)
	}, "1710000013_ruleset_key_limits.go")
}

func setRulesetKeyLimits(app core.App, max int) error {
	fields := map[string][]string{
		"games":              {"phase_key"},
		"participants":       {"role_key"},
		"chat_rooms":         {"team_key"},
		"chat_messages":      {"cue_key"},
		"attention_items":    {"cue_key"},
		"achievement_awards": {"achievement_key"},
		"ability_choices":    {"phase_key", "ability_key"},
	}
	for collectionName, names := range fields {
		collection, err := app.FindCollectionByNameOrId(collectionName)
		if err != nil {
			return err
		}
		for _, name := range names {
			field, ok := collection.Fields.GetByName(name).(*core.TextField)
			if !ok {
				return fmt.Errorf("%s field %q is not text", collectionName, name)
			}
			field.Max = max
		}
		if err := app.Save(collection); err != nil {
			return err
		}
	}
	return nil
}
