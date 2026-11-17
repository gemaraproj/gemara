// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cuelang.org/go/cue"
	cueyaml "cuelang.org/go/encoding/yaml"
)

// This spike catches loss of author-supplied edge evidence, accidental defaults
// for evidence, and the inability to publish an explicit no-match assertion.
func TestMappingResolutionPrototype(t *testing.T) {
	catalog := prototypeFixture(t, "../examples/inline-mapping-prototype/source-catalog.yaml", "#ControlCatalog")
	mapping := prototypeFixture(t, "../examples/inline-mapping-prototype/mapping-document.yaml", "#MappingDocument")
	before, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := resolvePrototype(catalog, mapping)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(resolved)
	if err != nil {
		t.Fatal(err)
	}
	value := schemaValue.LookupPath(cue.ParsePath("#ControlCatalog")).Unify(schemaCtx.CompileBytes(data))
	if err := value.Validate(cue.Concrete(true)); err != nil {
		t.Fatalf("resolved catalog is not distributable: %v", err)
	}
	var exported map[string]any
	if err := value.Decode(&exported); err != nil {
		t.Fatal(err)
	}
	want := prototypeFixture(t, "../examples/inline-mapping-prototype/resolved-catalog.yaml", "#ControlCatalog")
	if !reflect.DeepEqual(exported, want) {
		t.Fatal("resolved example differs from the resolver output")
	}
	first := exported["controls"].([]any)[0].(map[string]any)["guidelines"].([]any)
	legacy := first[0].(map[string]any)
	if legacy["relationship"] != "relates-to" {
		t.Errorf("legacy relationship = %v, want relates-to", legacy["relationship"])
	}
	legacyTarget := legacy["entries"].([]any)[0].(map[string]any)
	for _, key := range []string{"strength", "confidence-level", "rationale"} {
		if _, exists := legacyTarget[key]; exists {
			t.Errorf("invented %s for a legacy link", key)
		}
	}
	for _, raw := range mapping["mappings"].([]any) {
		m := raw.(map[string]any)
		var control map[string]any
		for _, c := range exported["controls"].([]any) {
			if c.(map[string]any)["id"] == m["source"] {
				control = c.(map[string]any)
			}
		}
		var found map[string]any
		for _, g := range control["guidelines"].([]any) {
			if g.(map[string]any)["id"] == m["id"] {
				found = g.(map[string]any)
			}
		}
		if found == nil {
			t.Fatalf("mapping %s lost", m["id"])
		}
		if found["relationship"] != m["relationship"] || found["remarks"] != m["remarks"] {
			t.Errorf("mapping semantics lost: %v", found)
		}
		if m["relationship"] == "no-match" {
			if _, exists := found["entries"]; exists {
				t.Error("no-match gained a fabricated target")
			}
			continue
		}
		entries := found["entries"].([]any)
		targets := m["targets"].([]any)
		if len(entries) != len(targets) {
			t.Fatalf("lost targets: %v", found)
		}
		for j, target := range targets {
			want := map[string]any{}
			for key, val := range target.(map[string]any) {
				if key == "entry-id" {
					key = "reference-id" // Preserve the existing grouped wire shape.
				}
				want[key] = val
			}
			if !reflect.DeepEqual(entries[j], want) {
				t.Errorf("target evidence changed: got %v, want %v", entries[j], want)
			}
		}
	}
	after, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("source catalog was mutated")
	}
	if path := os.Getenv("GEMARA_PROTOTYPE_OUTPUT"); path != "" {
		output, err := cueyaml.Encode(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, output, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPrototypeInlineConstraints(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		bad         bool
	}{
		{"legacy", `{"reference-id": "G", entries: [{"reference-id": "G1"}]}`, false},
		{"explicit gap", `{"reference-id": "G", relationship: "no-match"}`, false},
		{"missing target", `{"reference-id": "G", relationship: "implements"}`, true},
		{"missing legacy target", `{"reference-id": "G"}`, true},
		{"invalid relationship", `{"reference-id": "G", relationship: "invented", entries: [{"reference-id": "G1"}]}`, true},
		{"strength zero", `{"reference-id": "G", entries: [{"reference-id": "G1", strength: 0}]}`, true},
		{"strength eleven", `{"reference-id": "G", entries: [{"reference-id": "G1", strength: 11}]}`, true},
		{"invalid confidence", `{"reference-id": "G", entries: [{"reference-id": "G1", "confidence-level": "Certain"}]}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := schemaValue.LookupPath(cue.ParsePath("#CatalogMultiEntryMapping")).Unify(schemaCtx.CompileString(tc.input))
			if err := v.Validate(cue.Concrete(true)); (err != nil) != tc.bad {
				t.Fatalf("validation = %v, want error %v", err, tc.bad)
			}
		})
	}
}

func TestPrototypeResolutionBoundaries(t *testing.T) {
	for _, name := range []string{"undeclared source", "optional applicability"} {
		t.Run(name, func(t *testing.T) {
			catalog := prototypeFixture(t, "../examples/inline-mapping-prototype/source-catalog.yaml", "#ControlCatalog")
			mapping := prototypeFixture(t, "../examples/inline-mapping-prototype/mapping-document.yaml", "#MappingDocument")
			mappingMetadata := mapping["metadata"].(map[string]any)
			refs := mappingMetadata["mapping-references"].([]any)
			if name == "undeclared source" {
				mappingMetadata["mapping-references"] = refs[1:]
			} else {
				delete(mappingMetadata, "applicability-groups")
				for _, raw := range mapping["mappings"].([]any) {
					targets, _ := raw.(map[string]any)["targets"].([]any)
					for _, target := range targets {
						delete(target.(map[string]any), "applicability")
					}
				}
			}
			_, err := resolvePrototype(catalog, mapping)
			if (err != nil) != (name == "undeclared source") {
				t.Fatalf("resolution error = %v", err)
			}
		})
	}
}

func prototypeFixture(t *testing.T, path, definition string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, err := cueyaml.Extract(path, data)
	if err != nil {
		t.Fatal(err)
	}
	value := schemaValue.LookupPath(cue.ParsePath(definition)).Unify(schemaCtx.BuildFile(file))
	if err := value.Validate(cue.Concrete(true)); err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := value.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	return doc
}
