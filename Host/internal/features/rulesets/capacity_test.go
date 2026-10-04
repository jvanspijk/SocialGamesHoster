package rulesets

import "testing"

func TestPlayerCapacityBoundary(t *testing.T) {
	for _, maximum := range []int{31, 50, 100, 255, 256} {
		definition := testDefinition()
		definition.Metadata.MaxPlayers = maximum
		report := Validate(definition, map[string]struct{}{})
		if report.Valid() != (maximum <= 255) {
			t.Fatalf("maxPlayers=%d: errors=%v", maximum, report.Errors)
		}
	}
}
