// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestCatalogMappingWireConstraints(t *testing.T) {
	out := filepath.Join(t.TempDir(), "openapi.yaml")
	writeGenerated(t, out, "", "")
	doc, err := openapi3.NewLoader().LoadFromFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, input string
		bad         bool
	}{
		{"legacy", `{"reference-id":"G","entries":[{"reference-id":"G1"}]}`, false},
		{"evidence", `{"reference-id":"G","relationship":"implements","entries":[{"reference-id":"G1","strength":6,"confidence-level":"Medium","rationale":"Partial coverage"}]}`, false},
		{"explicit gap", `{"reference-id":"G","relationship":"no-match"}`, false},
		{"missing legacy target", `{"reference-id":"G"}`, true},
		{"missing target", `{"reference-id":"G","relationship":"implements"}`, true},
		{"empty targets", `{"reference-id":"G","entries":[]}`, true},
		{"invalid relationship", `{"reference-id":"G","relationship":"invented","entries":[{"reference-id":"G1"}]}`, true},
		{"strength zero", `{"reference-id":"G","entries":[{"reference-id":"G1","strength":0}]}`, true},
		{"strength eleven", `{"reference-id":"G","entries":[{"reference-id":"G1","strength":11}]}`, true},
		{"invalid confidence", `{"reference-id":"G","entries":[{"reference-id":"G1","confidence-level":"Certain"}]}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var value any
			if err := json.Unmarshal([]byte(tc.input), &value); err != nil {
				t.Fatal(err)
			}
			for _, linkage := range []struct{ definition, field string }{
				{"Control", "guidelines"}, {"Control", "threats"},
				{"Threat", "capabilities"}, {"Threat", "vectors"},
				{"Guideline", "principles"}, {"Guideline", "vectors"},
				{"Risk", "threats"},
			} {
				t.Run(linkage.definition+"."+linkage.field, func(t *testing.T) {
					holder := declaringSchema(doc.Components.Schemas, linkage.definition, linkage.field)
					if holder == nil {
						t.Fatal("linkage field not found")
					}
					err := holder.Properties[linkage.field].Value.VisitJSON([]any{value})
					if (err != nil) != tc.bad {
						t.Fatalf("validation = %v, want error %v", err, tc.bad)
					}
				})
			}
		})
	}
}
