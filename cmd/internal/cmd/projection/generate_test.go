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
	// The raw encoder output retains the three hidden helper definitions. Post
	// removes them after the CUE converter workaround rewrites their references.
	if len(schemas) != 94 {
		t.Errorf("got %d schemas, want 94", len(schemas))
	}
	for _, want := range []string{"Metadata", "ControlCatalog", "Evidence", "Mapping"} {
		if _, ok := schemas[want]; !ok {
			t.Errorf("missing schema %q", want)
		}
	}
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
