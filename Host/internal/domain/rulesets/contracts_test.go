package rulesets

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRoleDiscardRetiredMaxCopiesWhenReadingSavedRulesets(t *testing.T) {
	var role Role
	if err := json.Unmarshal([]byte(`{"id":"role","name":"Role","maxCopies":1}`), &role); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(role)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "maxCopies") {
		t.Fatalf("retired maxCopies was retained: %s", encoded)
	}
}
