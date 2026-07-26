package profiles

import "testing"

func TestNormalizeName(t *testing.T) {
	display, normalized, err := NormalizeName("  JASon\t van   Spijk  ")
	if err != nil {
		t.Fatal(err)
	}
	if display != "JASon van Spijk" || normalized != "jason van spijk" {
		t.Fatalf("unexpected names: %q %q", display, normalized)
	}
}

func TestNormalizeNameRejectsBidiControls(t *testing.T) {
	if _, _, err := NormalizeName("Jane\u202eDoe"); err == nil {
		t.Fatal("expected bidi control to be rejected")
	}
}
