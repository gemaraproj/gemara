// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"testing"

	"cuelang.org/go/cue"
)

func TestCrossLayerMappingFields(t *testing.T) {
	for _, linkage := range []struct{ definition, field string }{
		{"Control", "guidelines"}, {"Control", "threats"},
		{"Threat", "capabilities"}, {"Threat", "vectors"},
		{"Guideline", "principles"}, {"Guideline", "vectors"},
		{"Risk", "threats"},
	} {
		t.Run(linkage.definition+"."+linkage.field, func(t *testing.T) {
			fields, err := schemaValue.LookupPath(cue.ParsePath("#" + linkage.definition)).Fields(cue.Optional(true))
			if err != nil {
				t.Fatal(err)
			}
			var field cue.Value
			found := false
			for fields.Next() {
				if fields.Selector().Unquoted() == linkage.field {
					field, found = fields.Value(), true
				}
			}
			if !found {
				t.Fatal("linkage field not found")
			}
			for _, tc := range []struct {
				name, input string
				bad         bool
			}{
				{"legacy", `[{"reference-id":"A",entries:[{"reference-id":"T"}]}]`, false},
				{"evidence", `[{"reference-id":"A",relationship:"supports",entries:[{"reference-id":"T",strength:7,"confidence-level":"High",rationale:"Example assessment",applicability:["demo"],remarks:"Example evidence"}]}]`, false},
				{"gap", `[{"reference-id":"A",relationship:"no-match"}]`, false},
				{"missing targets", `[{"reference-id":"A"}]`, true},
				{"invalid strength", `[{"reference-id":"A",entries:[{"reference-id":"T",strength:11}]}]`, true},
				{"invalid relationship", `[{"reference-id":"A",relationship:"invented",entries:[{"reference-id":"T"}]}]`, true},
				{"invalid confidence", `[{"reference-id":"A",entries:[{"reference-id":"T","confidence-level":"Certain"}]}]`, true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					v := field.Unify(schemaCtx.CompileString(tc.input))
					if err := v.Validate(cue.Concrete(true)); (err != nil) != tc.bad {
						t.Fatalf("validation = %v, want error %v", err, tc.bad)
					}
					if tc.name == "legacy" {
						var links []map[string]any
						if err := v.Decode(&links); err != nil {
							t.Fatal(err)
						}
						if links[0]["relationship"] != "relates-to" {
							t.Fatalf("relationship = %v, want relates-to", links[0]["relationship"])
						}
						for _, key := range []string{"strength", "confidence-level", "rationale", "applicability"} {
							if _, exists := links[0]["entries"].([]any)[0].(map[string]any)[key]; exists {
								t.Errorf("legacy link gained %s", key)
							}
						}
					}
				})
			}
		})
	}
}
