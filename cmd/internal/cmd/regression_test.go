// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func propOf(t *testing.T, schemas map[string]any, schema, field string) map[string]any {
	t.Helper()
	s, _ := schemas[schema].(map[string]any)
	p, ok := mapOf(s["properties"])[field].(map[string]any)
	if !ok {
		t.Fatalf("%s.%s missing", schema, field)
	}
	return p
}

// gemara#474: the regex must reach the document unescaped, whether written as
// a named definition or directly on a field.
func TestRegexConstraintsReachTheDocument(t *testing.T) {
	schemas, _ := loadGenerated(t)

	email, _ := schemas["Email"].(map[string]any)
	pattern, _ := email["pattern"].(string)
	rx, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("Email pattern %q does not compile: %v", pattern, err)
	}
	for _, addr := range []string{"maintainer@example.com", "a.b+c@sub.example.co.uk"} {
		if !rx.MatchString(addr) {
			t.Errorf("Email pattern %q rejects %s", pattern, addr)
		}
	}

	for _, tc := range []struct{ schema, field string }{
		{"Entity", "uri"},
		{"LexiconReference", "url"},
		{"MappingReference", "url"},
		{"EvidenceMapping", "digest"},
	} {
		if p, _ := propOf(t, schemas, tc.schema, tc.field)["pattern"].(string); p == "" {
			t.Errorf("%s.%s has no pattern", tc.schema, tc.field)
		}
	}
}

// gemara#468: payload is the top type and must constrain nothing.
func TestEvidencePayloadIsUnconstrained(t *testing.T) {
	schemas, _ := loadGenerated(t)
	p := propOf(t, schemas, "Evidence", "payload")
	if _, ok := p["type"]; ok {
		t.Errorf("Evidence.payload has type %v, want none", p["type"])
	}
}

// gemara#466: the generated Datetime must accept the repo's own fixtures.
func TestDatetimeIsDateTime(t *testing.T) {
	schemas, _ := loadGenerated(t)
	dt, _ := schemas["Datetime"].(map[string]any)
	if dt["format"] != "date-time" {
		t.Errorf("Datetime format = %v, want date-time", dt["format"])
	}
	pattern, _ := dt["pattern"].(string)
	if pattern == "" {
		return
	}
	rx := regexp.MustCompile(pattern)
	for _, ts := range []string{"2025-08-22T16:02:00Z", "2026-02-10T15:05:00Z"} {
		if !rx.MatchString(ts) {
			t.Errorf("Datetime pattern %q rejects %s", pattern, ts)
		}
	}
}

// gemara#473: aliased labels and embedded definitions must both reach the
// document. Embeddings arrive through allOf, which is what the spec chose.
func TestAliasedAndEmbeddedFieldsReachTheDocument(t *testing.T) {
	schemas, _ := loadGenerated(t)
	for _, field := range []string{"mapping-references", "applicability-groups"} {
		propOf(t, schemas, "Metadata", field)
	}
	props, req := effective(schemas, "ControlCatalog", map[string]bool{})
	for _, field := range []string{"title", "groups", "extends", "imports"} {
		if _, ok := props[field]; !ok {
			t.Errorf("ControlCatalog.%s missing (embedded from #Catalog)", field)
		}
	}
	if !req["title"] {
		t.Error("ControlCatalog.title is required in CUE but not in the document")
	}
}

// The golden document makes any drift in encoder behaviour across CUE
// upgrades show up here rather than in a release asset.
// Regenerate with: UPDATE_GOLDEN=1 go test ./internal/cmd/ -run TestGolden
func TestGoldenDocument(t *testing.T) {
	out := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := convertCUEToOpenAPI("../../..", out, ConvertOpts{Version: "golden"}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "openapi.golden.yaml")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Log("golden file updated")
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run UPDATE_GOLDEN=1 go test ./internal/cmd/ -run TestGolden)", err)
	}
	if string(got) != string(want) {
		t.Error("generated document differs from testdata/openapi.golden.yaml; inspect the diff, then regenerate with UPDATE_GOLDEN=1 if the change is intended")
	}
}
