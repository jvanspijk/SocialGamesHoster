package rulesets

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestBundleRoundTrip(t *testing.T) {
	definition := testDefinition()
	manifest := BundleManifest{
		SourceApplicationVersion:  "test",
		MinimumApplicationVersion: "1.0.0",
		LogicalSourceRulesetID:    "test",
		SourceVersionNumber:       1,
		Name:                      "Test",
	}
	data, err := WriteBundle(manifest, definition, map[string][]byte{})
	if err != nil {
		t.Fatal(err)
	}
	imported, err := ReadBundle(data)
	if err != nil {
		t.Fatal(err)
	}
	if imported.Definition.Metadata.Name != definition.Metadata.Name {
		t.Fatalf("unexpected definition: %#v", imported.Definition.Metadata)
	}
}

func TestBundleRejectsTraversal(t *testing.T) {
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	entry, err := writer.Create("../ruleset.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("{}")); err != nil {
		t.Fatal(err)
	}
	manifest, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manifest.Write([]byte("{}")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadBundle(output.Bytes()); err == nil || !strings.Contains(err.Error(), "unsafe path") {
		t.Fatalf("expected unsafe path rejection, got %v", err)
	}
}
