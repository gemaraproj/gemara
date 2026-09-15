// SPDX-License-Identifier: Apache-2.0

package cmd

import "testing"

func TestGenerateRawProducesSchemas(t *testing.T) {
	v, _, _, err := loadPrepared("../../..")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := generateRaw(v, "Gemara", "test")
	if err != nil {
		t.Fatal(err)
	}
	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	// 90 public definitions plus the two #_…Strict helpers Task 5 collapses.
	if len(schemas) != 92 {
		t.Errorf("got %d schemas, want 92", len(schemas))
	}
	for _, want := range []string{"Metadata", "ControlCatalog", "Evidence", "_MappingStrict"} {
		if _, ok := schemas[want]; !ok {
			t.Errorf("missing schema %q", want)
		}
	}
}

// The encoder's non-concrete `default` blocks leak raw CUE and @go()
// attributes into the document; FieldFilter must suppress them.
func TestGenerateRawHasNoDefaultsOrAdditionalItems(t *testing.T) {
	v, _, _, err := loadPrepared("../../..")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := generateRaw(v, "Gemara", "test")
	if err != nil {
		t.Fatal(err)
	}
	var walk func(any)
	walk = func(n any) {
		switch x := n.(type) {
		case map[string]any:
			for _, k := range []string{"default", "additionalItems"} {
				if _, ok := x[k]; ok {
					t.Errorf("document still contains %q", k)
				}
			}
			for _, v := range x {
				walk(v)
			}
		case []any:
			for _, v := range x {
				walk(v)
			}
		}
	}
	walk(doc)
}
