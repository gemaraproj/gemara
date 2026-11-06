// SPDX-License-Identifier: Apache-2.0

package jsonschema

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestDiscoverDefinitions(t *testing.T) {
	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)

	names, err := gen.DiscoverDefinitions()
	if err != nil {
		t.Fatalf("DiscoverDefinitions failed: %v", err)
	}

	if len(names) == 0 {
		t.Fatalf("expected non-empty definition list, got 0")
	}

	foundControlCatalog := false
	foundMetadata := false
	for _, n := range names {
		if n == "#ControlCatalog" {
			foundControlCatalog = true
		}
		if n == "#Metadata" {
			foundMetadata = true
		}
		if len(n) >= 2 && n[:2] == "#_" {
			t.Errorf("discovered definition should not be private (#_): %s", n)
		}
	}

	if !foundControlCatalog {
		t.Errorf("expected #ControlCatalog in discovered definitions")
	}
	if !foundMetadata {
		t.Errorf("expected #Metadata in discovered definitions")
	}
}

func TestGenerateAndDeterminism(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue binary not found on PATH; skipping generation test")
	}

	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)

	schema1, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate run 1 failed: %v", err)
	}

	if schema1["$schema"] != schemaDialect {
		t.Errorf("expected $schema to be %s, got %v", schemaDialect, schema1["$schema"])
	}

	defs, ok := schema1["$defs"].(map[string]any)
	if !ok || len(defs) == 0 {
		t.Fatalf("expected non-empty $defs map, got %v", schema1["$defs"])
	}

	if _, exists := defs["ControlCatalog"]; !exists {
		t.Errorf("expected defs to contain ControlCatalog")
	}
	if _, exists := defs["Metadata"]; !exists {
		t.Errorf("expected defs to contain Metadata")
	}

	b1, err := json.MarshalIndent(schema1, "", "  ")
	if err != nil {
		t.Fatalf("marshalling schema 1: %v", err)
	}

	// Verify no escaped references leak into the output
	s1 := string(b1)
	for _, forbidden := range []string{"%23", "%22", "#/$defs/#"} {
		if strings.Contains(s1, forbidden) {
			t.Errorf("found forbidden reference substring in generated output: %s", forbidden)
		}
	}

	// Second run for determinism
	schema2, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate run 2 failed: %v", err)
	}
	b2, err := json.MarshalIndent(schema2, "", "  ")
	if err != nil {
		t.Fatalf("marshalling schema 2: %v", err)
	}

	hash1 := sha256.Sum256(b1)
	hash2 := sha256.Sum256(b2)

	if hash1 != hash2 {
		t.Errorf("generation is not deterministic: hash1=%s, hash2=%s",
			hex.EncodeToString(hash1[:]), hex.EncodeToString(hash2[:]))
	}
}

func TestWrite(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue binary not found on PATH; skipping write test")
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "nested", "gemara-v1.schema.json")

	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)

	if err := gen.Write(outPath); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading generated schema: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("generated schema file is empty")
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshalling generated file: %v", err)
	}

	if parsed["$schema"] != schemaDialect {
		t.Errorf("expected $schema %q, got %q", schemaDialect, parsed["$schema"])
	}
}

func TestInvalidSchemaDir(t *testing.T) {
	gen := NewGenerator(filepath.Join(t.TempDir(), "nonexistent"))

	if _, err := gen.DiscoverDefinitions(); err == nil {
		t.Errorf("expected error for nonexistent schema dir, got nil")
	}

	if _, err := gen.Generate(); err == nil {
		t.Errorf("expected error generating from nonexistent schema dir, got nil")
	}
}

func TestMissingCueBin(t *testing.T) {
	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir).WithCueBin("nonexistent-cue-bin-12345")

	if _, err := gen.Generate(); err == nil {
		t.Errorf("expected error with missing cue binary, got nil")
	}
}

func TestHelperRulesPreservedInJSONSchema(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue binary not found on PATH; skipping test")
	}

	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)

	schema, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("missing $defs map")
	}

	// 1. Verify AssessmentLog carries the start rule from #_AssessmentLogStrict
	al, ok := defs["AssessmentLog"].(map[string]any)
	if !ok {
		t.Fatalf("missing AssessmentLog in $defs")
	}
	alAllOf, ok := al["allOf"].([]any)
	if !ok || len(alAllOf) == 0 {
		t.Fatalf("expected AssessmentLog to have allOf containing helper rules, got %v", al["allOf"])
	}
	alBytes, _ := json.Marshal(alAllOf)
	if !jsonContains(string(alBytes), "start") || !jsonContains(string(alBytes), "Not Run") {
		t.Errorf("AssessmentLog allOf missing expected start conditional: %s", string(alBytes))
	}

	// 2. Verify Evidence carries the payload-or-source rule from #_EvidenceStrict
	ev, ok := defs["Evidence"].(map[string]any)
	if !ok {
		t.Fatalf("missing Evidence in $defs")
	}
	evAllOf, ok := ev["allOf"].([]any)
	if !ok || len(evAllOf) == 0 {
		t.Fatalf("expected Evidence to have allOf containing helper rules, got %v", ev["allOf"])
	}
	evBytes, _ := json.Marshal(evAllOf)
	if !jsonContains(string(evBytes), "payload") || !jsonContains(string(evBytes), "source") {
		t.Errorf("Evidence allOf missing expected payload-or-source conditional: %s", string(evBytes))
	}

	// 3. Verify Mapping carries the targets-required rule from #_MappingStrict
	m, ok := defs["Mapping"].(map[string]any)
	if !ok {
		t.Fatalf("missing Mapping in $defs")
	}
	mAllOf, ok := m["allOf"].([]any)
	if !ok || len(mAllOf) == 0 {
		t.Fatalf("expected Mapping to have allOf containing helper rules, got %v", m["allOf"])
	}
	mBytes, _ := json.Marshal(mAllOf)
	if !jsonContains(string(mBytes), "targets") || !jsonContains(string(mBytes), "no-match") {
		t.Errorf("Mapping allOf missing expected targets conditional: %s", string(mBytes))
	}
}

func jsonContains(jsonStr, substr string) bool {
	return filepath.Clean(jsonStr) != "" && (len(jsonStr) > 0 && (len(substr) > 0 && true)) && (len(jsonStr) >= len(substr) && containsSubstr(jsonStr, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestStripRefEncoding(t *testing.T) {
	input := `{"$ref": "#/$defs/#Control", "other": "#/$defs/%23Actor%22", "ev": "#/$defs/_EvidenceStrict"}`
	got := stripRefEncoding(input)
	want := `{"$ref": "#/$defs/Control", "other": "#/$defs/Actor", "ev": "#/$defs/Evidence"}`
	if got != want {
		t.Errorf("stripRefEncoding got %q, want %q", got, want)
	}
}

func TestHelperRulesCoverage(t *testing.T) {
	schemaDir := filepath.Join("..", "..", "..", "..")
	absDir, err := filepath.Abs(schemaDir)
	if err != nil {
		t.Fatalf("resolving schema directory: %v", err)
	}

	insts := load.Instances([]string{"."}, &load.Config{Dir: absDir})
	if len(insts) == 0 || insts[0].Err != nil {
		t.Fatalf("loading CUE instances in %s: %v", absDir, insts[0].Err)
	}

	ctx := cuecontext.New()
	v := ctx.BuildInstance(insts[0])
	if err := v.Err(); err != nil {
		t.Fatalf("building CUE instance: %v", err)
	}

	it, err := v.Fields(cue.Definitions(true))
	if err != nil {
		t.Fatalf("enumerating CUE definitions: %v", err)
	}

	knownStrictHelpers := map[string]bool{
		"#_AssessmentLogStrict": true,
		"#_EvidenceStrict":      true,
		"#_MappingStrict":       true,
	}

	var discoveredHelpers []string
	for it.Next() {
		sel := it.Selector().String()
		if strings.HasPrefix(sel, "#_") {
			discoveredHelpers = append(discoveredHelpers, sel)
			if !knownStrictHelpers[sel] {
				t.Errorf("unaccounted #_ helper in CUE schemas: %s (must be projected and dropped in jsonschema)", sel)
			}
		}
	}

	if len(discoveredHelpers) != len(knownStrictHelpers) {
		t.Errorf("expected %d strict helpers in CUE schemas, found %d: %v", len(knownStrictHelpers), len(discoveredHelpers), discoveredHelpers)
	}

	gen := NewGenerator(schemaDir)
	schema, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("missing $defs map")
	}

	// Assert no helper or private definitions starting with '_' exist in $defs
	for k := range defs {
		if strings.HasPrefix(k, "_") {
			t.Errorf("helper definition leaked into $defs: %s", k)
		}
	}

	// Assert no $ref in the entire schema points to any helper
	var checkRefs func(v any)
	checkRefs = func(v any) {
		switch node := v.(type) {
		case map[string]any:
			if ref, ok := node["$ref"].(string); ok {
				for h := range knownStrictHelpers {
					unprefixed := strings.TrimPrefix(h, "#")
					if strings.Contains(ref, unprefixed) {
						t.Errorf("found dangling reference to helper %s in schema: %s", h, ref)
					}
				}
			}
			for _, child := range node {
				checkRefs(child)
			}
		case []any:
			for _, item := range node {
				checkRefs(item)
			}
		}
	}
	checkRefs(defs)
}

func TestGeneratedSchemaValidatesFixtures(t *testing.T) {
	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)
	schema, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		t.Fatalf("unmarshal schema doc: %v", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", doc); err != nil {
		t.Fatalf("AddResource: %v", err)
	}

	fixtures := []struct {
		file    string
		defName string
	}{
		{"good-ccc.yaml", "ControlCatalog"},
		{"good-threat-catalog.yaml", "ThreatCatalog"},
		{"good-capability-catalog.yaml", "CapabilityCatalog"},
		{"good-audit-log.yaml", "AuditLog"},
		{"good-policy.yaml", "Policy"},
		{"good-mapping-document.yaml", "MappingDocument"},
		{"good-lexicon.yaml", "Lexicon"},
		{"good-risk-catalog.yaml", "RiskCatalog"},
	}

	for _, tc := range fixtures {
		t.Run(tc.file, func(t *testing.T) {
			sch, err := c.Compile("schema.json#/$defs/" + tc.defName)
			if err != nil {
				t.Fatalf("Compile %s: %v", tc.defName, err)
			}

			fixturePath := filepath.Join(schemaDir, "test", "test-data", tc.file)
			fixtureData, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			var raw any
			if err := yaml.Unmarshal(fixtureData, &raw); err != nil {
				t.Fatalf("unmarshal yaml: %v", err)
			}
			rawJson, err := json.Marshal(raw)
			if err != nil {
				t.Fatalf("marshal json: %v", err)
			}
			valDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(rawJson))
			if err != nil {
				t.Fatalf("unmarshal fixture val: %v", err)
			}

			if err := sch.Validate(valDoc); err != nil {
				t.Errorf("fixture %s failed validation: %v", tc.file, err)
			}
		})
	}

	// Verify Evidence rule: an AssessmentLog with only payload MUST pass
	t.Run("EvidenceWithOnlyPayloadInAssessmentLog", func(t *testing.T) {
		sch, err := c.Compile("schema.json#/$defs/AssessmentLog")
		if err != nil {
			t.Fatalf("Compile AssessmentLog: %v", err)
		}

		alData := map[string]any{
			"id": "AL-1",
			"requirement": map[string]any{
				"reference-id": "REF-1",
				"entry-id":     "REQ-1",
			},
			"result":        "Not Run",
			"description":   "Verify procedure",
			"message":       "Assessment not run",
			"applicability": []any{"all"},
			"steps":         []any{"step 1"},
			"evidence": []any{
				map[string]any{
					"id":           "EV-1",
					"type":         "document",
					"collected-at": "2024-01-01T00:00:00Z",
					"payload": map[string]any{
						"info": "raw evidence data only, no source",
					},
				},
			},
		}

		rawJson, err := json.Marshal(alData)
		if err != nil {
			t.Fatalf("marshal json: %v", err)
		}
		valDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(rawJson))
		if err != nil {
			t.Fatalf("unmarshal val: %v", err)
		}

		if err := sch.Validate(valDoc); err != nil {
			t.Errorf("AssessmentLog with only payload failed validation: %v", err)
		}
	})

	// Negative check: an AssessmentLog with neither payload nor source MUST fail
	t.Run("EvidenceWithNeitherPayloadNorSourceFails", func(t *testing.T) {
		sch, err := c.Compile("schema.json#/$defs/AssessmentLog")
		if err != nil {
			t.Fatalf("Compile AssessmentLog: %v", err)
		}

		alData := map[string]any{
			"id": "AL-2",
			"requirement": map[string]any{
				"reference-id": "REF-1",
				"entry-id":     "REQ-1",
			},
			"result":        "Not Run",
			"description":   "Verify procedure",
			"message":       "Assessment not run",
			"applicability": []any{"all"},
			"steps":         []any{"step 1"},
			"evidence": []any{
				map[string]any{
					"id":           "EV-2",
					"type":         "document",
					"collected-at": "2024-01-01T00:00:00Z",
				},
			},
		}

		rawJson, _ := json.Marshal(alData)
		valDoc, _ := jsonschema.UnmarshalJSON(bytes.NewReader(rawJson))
		if err := sch.Validate(valDoc); err == nil {
			t.Errorf("expected validation failure for evidence with neither payload nor source, but passed")
		}
	})

	// Negative check: unevaluatedProperties: false rejects unknown properties on extended types
	t.Run("UnevaluatedPropertiesRejectsUnknownFields", func(t *testing.T) {
		sch, err := c.Compile("schema.json#/$defs/ControlCatalog")
		if err != nil {
			t.Fatalf("Compile ControlCatalog: %v", err)
		}

		fixturePath := filepath.Join(schemaDir, "test", "test-data", "good-ccc.yaml")
		fixtureData, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}

		var raw map[string]any
		if err := yaml.Unmarshal(fixtureData, &raw); err != nil {
			t.Fatalf("unmarshal yaml: %v", err)
		}
		raw["unexpected_extra_property"] = "bad"

		rawJson, _ := json.Marshal(raw)
		valDoc, _ := jsonschema.UnmarshalJSON(bytes.NewReader(rawJson))
		if err := sch.Validate(valDoc); err == nil {
			t.Errorf("expected validation failure for unknown property under unevaluatedProperties: false, but passed")
		}
	})
}
