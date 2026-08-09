package rulesets

import "encoding/json"

// DecodeSnapshot decodes an untyped stored game ruleset snapshot.
func DecodeSnapshot(snapshot any) (DefinitionV1, error) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return DefinitionV1{}, err
	}
	var definition DefinitionV1
	return definition, json.Unmarshal(data, &definition)
}
