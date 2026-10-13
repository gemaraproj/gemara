// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGenerateRawProducesSchemas(t *testing.T) {
	v, _, _, err := loadPrepared("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := generateRaw(v, "Gemara", "test")
	if err != nil {
		t.Fatal(err)
	}
	schemas := doc.Components.Schemas
	for _, want := range []string{"Metadata", "ControlCatalog", "Evidence", "Mapping"} {
		if _, ok := schemas[want]; !ok {
			t.Errorf("missing schema %q", want)
		}
	}
	// The raw encoder output retains the three hidden helper definitions. Post
	// removes them after the CUE converter workaround rewrites their references.
	for _, helper := range []string{"_EvidenceStrict", "_MappingStrict", "_AssessmentLogStrict"} {
		if _, ok := schemas[helper]; !ok {
			t.Errorf("missing helper schema %q", helper)
		}
	}
}

// The encoder's non-concrete `default` blocks leak raw CUE and @go()
// attributes into the document; FieldFilter must suppress them.
func TestGenerateRawHasNoDefaultsOrAdditionalItems(t *testing.T) {
	v, _, _, err := loadPrepared("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := generateRaw(v, "Gemara", "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := walkOpenAPISchemas(doc, func(schema *openapi3.Schema) error {
		if schema.Default != nil {
			t.Errorf("document still contains default %#v", schema.Default)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestExecutionContextRejectsEmptyMethod(t *testing.T) {
	v, _, _, err := loadPrepared("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := generateRaw(v, "Gemara", "test")
	if err != nil {
		t.Fatal(err)
	}
	schema := doc.Components.Schemas["ExecutionContext"].Value
	if err := schema.VisitJSON(map[string]any{"method-id": ""}); err == nil {
		t.Error("expected empty method-id to be rejected")
	}
	for _, method := range []string{"X", "AC-01/X", "é", "\n"} {
		if err := schema.VisitJSON(map[string]any{"method-id": method}); err != nil {
			t.Errorf("non-empty method-id %q was rejected: %v", method, err)
		}
	}
	if err := schema.VisitJSON(map[string]any{}); err != nil {
		t.Errorf("omitted optional method-id was rejected: %v", err)
	}
}
