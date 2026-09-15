// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"reflect"
	"testing"
)

func TestUnwrapTupleItems(t *testing.T) {
	doc := map[string]any{
		"components": map[string]any{"schemas": map[string]any{
			"A": map[string]any{"properties": map[string]any{
				"xs": map[string]any{
					"type":     "array",
					"minItems": float64(1),
					"items":    []any{map[string]any{"$ref": "#/components/schemas/B"}},
				},
			}},
		}},
	}
	if err := unwrapTupleItems(doc); err != nil {
		t.Fatal(err)
	}
	xs := doc["components"].(map[string]any)["schemas"].(map[string]any)["A"].(map[string]any)["properties"].(map[string]any)["xs"].(map[string]any)
	want := map[string]any{"$ref": "#/components/schemas/B"}
	if !reflect.DeepEqual(xs["items"], want) {
		t.Errorf("items = %#v, want %#v", xs["items"], want)
	}
	if xs["minItems"] != float64(1) {
		t.Errorf("minItems lost: %#v", xs["minItems"])
	}
}

// A multi-element tuple is a per-item schema, which OpenAPI cannot express.
// Fail rather than silently keeping the first entry.
func TestUnwrapTupleItemsRejectsMultiElement(t *testing.T) {
	doc := map[string]any{
		"components": map[string]any{"schemas": map[string]any{
			"A": map[string]any{"properties": map[string]any{
				"xs": map[string]any{"items": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "integer"},
				}},
			}},
		}},
	}
	if err := unwrapTupleItems(doc); err == nil {
		t.Error("expected an error for a multi-element tuple, got nil")
	}
}

func helperDoc(helperProps map[string]any) map[string]any {
	return map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Base": map[string]any{
			"type":       "object",
			"required":   []any{"f"},
			"properties": map[string]any{"f": map[string]any{"type": "string"}},
		},
		"_BaseStrict": map[string]any{
			"type":       "object",
			"properties": helperProps,
			"allOf": []any{
				map[string]any{"$ref": "#/components/schemas/Base"},
				map[string]any{"required": []any{"f"}},
			},
		},
		"User": map[string]any{"properties": map[string]any{
			"x": map[string]any{"$ref": "#/components/schemas/_BaseStrict"},
		}},
	}}}
}

func TestCollapseHelperSchemas(t *testing.T) {
	doc := helperDoc(map[string]any{"f": map[string]any{"type": "string"}})
	if err := collapseHelperSchemas(doc); err != nil {
		t.Fatal(err)
	}
	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	if _, ok := schemas["_BaseStrict"]; ok {
		t.Error("_BaseStrict was not dropped")
	}
	x := schemas["User"].(map[string]any)["properties"].(map[string]any)["x"].(map[string]any)
	if x["$ref"] != "#/components/schemas/Base" {
		t.Errorf("ref = %v, want the base", x["$ref"])
	}
}

// A helper that adds a constraint the base does not already carry must not be
// silently discarded.
func TestCollapseHelperSchemasRejectsNonRedundant(t *testing.T) {
	doc := helperDoc(map[string]any{"f": map[string]any{"type": "string", "minLength": float64(3)}})
	if err := collapseHelperSchemas(doc); err == nil {
		t.Error("expected an error for a non-redundant helper, got nil")
	}
}

// ControlEvaluation.assessment-logs refs both AssessmentLog and
// _AssessmentLogStrict in one allOf; collapsing must not leave a duplicate.
func TestCollapseHelperSchemasDeduplicates(t *testing.T) {
	doc := helperDoc(map[string]any{"f": map[string]any{"type": "string"}})
	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	schemas["User"].(map[string]any)["properties"].(map[string]any)["x"] = map[string]any{
		"allOf": []any{
			map[string]any{"$ref": "#/components/schemas/Base"},
			map[string]any{"$ref": "#/components/schemas/_BaseStrict"},
			map[string]any{"required": []any{"f"}},
		},
	}
	if err := collapseHelperSchemas(doc); err != nil {
		t.Fatal(err)
	}
	all := schemas["User"].(map[string]any)["properties"].(map[string]any)["x"].(map[string]any)["allOf"].([]any)
	refs := 0
	for _, e := range all {
		if e.(map[string]any)["$ref"] == "#/components/schemas/Base" {
			refs++
		}
	}
	if refs != 1 {
		t.Errorf("Base referenced %d times in allOf, want 1", refs)
	}
}

func TestRestoreRefDescriptions(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Metadata": map[string]any{"properties": map[string]any{
			"date":  map[string]any{"$ref": "#/components/schemas/Datetime"},
			"draft": map[string]any{"type": "boolean"},
		}},
	}}}
	docs := map[string]string{
		"Metadata.date":  "date is the publication or effective date of this artifact",
		"Metadata.draft": "draft indicates whether this artifact is a pre-release version",
	}
	if err := restoreRefDescriptions(doc, docs); err != nil {
		t.Fatal(err)
	}
	props := doc["components"].(map[string]any)["schemas"].(map[string]any)["Metadata"].(map[string]any)["properties"].(map[string]any)

	// A $ref property is wrapped, because OpenAPI 3.0 forbids $ref siblings.
	date := props["date"].(map[string]any)
	if _, ok := date["$ref"]; ok {
		t.Error("date still carries a sibling $ref")
	}
	if date["description"] != docs["Metadata.date"] {
		t.Errorf("date description = %v", date["description"])
	}
	all := date["allOf"].([]any)
	if len(all) != 1 || all[0].(map[string]any)["$ref"] != "#/components/schemas/Datetime" {
		t.Errorf("date allOf = %#v", all)
	}

	// A property that already has a description is left alone.
	if _, ok := props["draft"].(map[string]any)["allOf"]; ok {
		t.Error("non-ref property was wrapped")
	}
}

// Every $ref sibling except the wrapper form must be gone afterwards.
func TestRestoreRefDescriptionsLeavesNoSiblings(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"A": map[string]any{"properties": map[string]any{
			"x": map[string]any{"$ref": "#/components/schemas/B", "description": "stale"},
		}},
	}}}
	if err := restoreRefDescriptions(doc, map[string]string{"A.x": "real"}); err != nil {
		t.Fatal(err)
	}
	x := doc["components"].(map[string]any)["schemas"].(map[string]any)["A"].(map[string]any)["properties"].(map[string]any)["x"].(map[string]any)
	if _, ok := x["$ref"]; ok {
		t.Errorf("$ref sibling survived: %#v", x)
	}
	if x["description"] != "real" {
		t.Errorf("description = %v, want the CUE doc comment", x["description"])
	}
}

// A $ref property with no CUE doc comment and no existing description must be
// left exactly as it is: the allOf wrapper exists only to carry a description,
// so wrapping without one is churn in a released document.
func TestRestoreRefDescriptionsLeavesUndocumentedRefsBare(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"A": map[string]any{"properties": map[string]any{
			"x": map[string]any{"$ref": "#/components/schemas/B"},
		}},
	}}}
	if err := restoreRefDescriptions(doc, map[string]string{}); err != nil {
		t.Fatal(err)
	}
	x := doc["components"].(map[string]any)["schemas"].(map[string]any)["A"].(map[string]any)["properties"].(map[string]any)["x"].(map[string]any)
	if x["$ref"] != "#/components/schemas/B" {
		t.Errorf("bare $ref was disturbed: %#v", x)
	}
	if _, ok := x["allOf"]; ok {
		t.Errorf("undocumented $ref was wrapped in allOf: %#v", x)
	}
	if _, ok := x["description"]; ok {
		t.Errorf("empty description was added: %#v", x)
	}
}

// A property that is an inline narrowing rather than a $ref still gets its CUE
// doc comment, and an existing description is never overwritten.
func TestRestoreRefDescriptionsFillsNonRefProperties(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"A": map[string]any{"properties": map[string]any{
			"narrowed": map[string]any{"type": "object"},
			"kept":     map[string]any{"type": "string", "description": "already here"},
		}},
	}}}
	docs := map[string]string{"A.narrowed": "from CUE", "A.kept": "should not win"}
	if err := restoreRefDescriptions(doc, docs); err != nil {
		t.Fatal(err)
	}
	props := doc["components"].(map[string]any)["schemas"].(map[string]any)["A"].(map[string]any)["properties"].(map[string]any)
	if got := props["narrowed"].(map[string]any)["description"]; got != "from CUE" {
		t.Errorf("narrowed description = %v, want the CUE doc comment", got)
	}
	if _, ok := props["narrowed"].(map[string]any)["allOf"]; ok {
		t.Error("a non-$ref property must not be wrapped in allOf")
	}
	if got := props["kept"].(map[string]any)["description"]; got != "already here" {
		t.Errorf("existing description was overwritten: %v", got)
	}
}

// A helper whose base is itself a helper must be rejected: collapsing both with a
// single-level rewrite would leave a $ref pointing at a schema that was deleted.
func TestCollapseHelperSchemasRejectsChainedHelpers(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Base": map[string]any{
			"type":       "object",
			"required":   []any{"f"},
			"properties": map[string]any{"f": map[string]any{"type": "string"}},
		},
		"_Inner": map[string]any{
			"type":  "object",
			"allOf": []any{map[string]any{"$ref": "#/components/schemas/Base"}},
		},
		"_Outer": map[string]any{
			"type":  "object",
			"allOf": []any{map[string]any{"$ref": "#/components/schemas/_Inner"}},
		},
		"User": map[string]any{"properties": map[string]any{
			"x": map[string]any{"$ref": "#/components/schemas/_Outer"},
		}},
	}}}
	if err := collapseHelperSchemas(doc); err == nil {
		t.Error("expected an error for a chained helper, got nil")
	}
}

func TestApplyTimeFormats(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Datetime": map[string]any{"type": "string", "description": "d"},
	}}}
	err := applyTimeFormats(doc, map[string]string{"Datetime": "2006-01-02T15:04:05Z07:00"})
	if err != nil {
		t.Fatal(err)
	}
	got := doc["components"].(map[string]any)["schemas"].(map[string]any)["Datetime"].(map[string]any)
	if got["format"] != "date-time" {
		t.Errorf("format = %v, want date-time", got["format"])
	}
	if _, ok := got["pattern"]; ok {
		t.Errorf("unexpected pattern: %v", got["pattern"])
	}
}

// An unmapped layout must be an error, not a narrower fallback. Emitting a
// date-only pattern for a date-time layout is gemara#466.
func TestApplyTimeFormatsRejectsUnknownLayout(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Weird": map[string]any{"type": "string"},
	}}}
	if err := applyTimeFormats(doc, map[string]string{"Weird": "Mon Jan _2 15:04:05 2006"}); err == nil {
		t.Error("expected an error for an unmapped layout, got nil")
	}
}

// CUE's `|` is not an exclusive or: `#ArtifactType | string` means "one of
// these recommended values, or any string". The encoder renders it as `oneOf`,
// which in JSON Schema means EXACTLY one branch may match — so every
// recommended value matches both the enum branch and the unconstrained one and
// is rejected, while arbitrary strings pass. Exactly backwards.
func TestOpenDisjunctionsBecomeAnyOf(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"EvidenceType": map[string]any{
			"type": "string",
			"oneOf": []any{
				map[string]any{"enum": []any{"EvaluationLog", "AuditLog"}},
				map[string]any{},
			},
		},
	}}}
	if err := openDisjunctions(doc); err != nil {
		t.Fatal(err)
	}
	s := doc["components"].(map[string]any)["schemas"].(map[string]any)["EvidenceType"].(map[string]any)
	if _, ok := s["oneOf"]; ok {
		t.Error("oneOf survived; the recommended values are still unsatisfiable")
	}
	all, ok := s["anyOf"].([]any)
	if !ok {
		t.Fatalf("anyOf = %#v, want the rewritten branches", s["anyOf"])
	}
	if len(all) != 2 {
		t.Errorf("anyOf has %d branches, want the original 2", len(all))
	}
	enum, _ := all[0].(map[string]any)["enum"].([]any)
	if len(enum) != 2 {
		t.Errorf("the enum branch lost its values: %#v", all[0])
	}
}

// A closed disjunction has no unconstrained branch, so `oneOf` is not
// self-defeating there. Widening it would weaken a real constraint, so it is
// left exactly as the encoder produced it.
func TestClosedDisjunctionsKeepOneOf(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Either": map[string]any{
			"oneOf": []any{
				map[string]any{"type": "string"},
				map[string]any{"type": "integer"},
			},
		},
	}}}
	if err := openDisjunctions(doc); err != nil {
		t.Fatal(err)
	}
	s := doc["components"].(map[string]any)["schemas"].(map[string]any)["Either"].(map[string]any)
	if _, ok := s["oneOf"]; !ok {
		t.Errorf("oneOf was rewritten on a closed disjunction: %#v", s)
	}
}

// A CUE field with a default always has a value, so the encoder emits it as
// required. On the wire it is the opposite: the producer may omit it and the
// consumer applies the default. Requiring it rejects documents the CUE accepts.
func TestApplyDefaultsMakesDefaultedFieldsOptional(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Control": map[string]any{
			"properties": map[string]any{
				"id":    map[string]any{"type": "string"},
				"state": map[string]any{"type": "string"},
			},
			"required": []any{"id", "state"},
		},
	}}}
	if err := applyDefaults(doc, map[string]any{"Control.state": "Active"}); err != nil {
		t.Fatal(err)
	}
	s := doc["components"].(map[string]any)["schemas"].(map[string]any)["Control"].(map[string]any)
	state := s["properties"].(map[string]any)["state"].(map[string]any)
	if state["default"] != "Active" {
		t.Errorf("default = %#v, want Active", state["default"])
	}
	var req []string
	for _, r := range sliceOf(s["required"]) {
		req = append(req, r.(string))
	}
	if len(req) != 1 || req[0] != "id" {
		t.Errorf("required = %v, want only [id]", req)
	}
}

// A default recorded for a field the document does not have means the name was
// misattributed; the default would be silently lost.
func TestApplyDefaultsRejectsUnknownProperty(t *testing.T) {
	doc := map[string]any{"components": map[string]any{"schemas": map[string]any{
		"Control": map[string]any{"properties": map[string]any{"id": map[string]any{"type": "string"}}},
	}}}
	if err := applyDefaults(doc, map[string]any{"Control.nope": "x"}); err == nil {
		t.Error("expected an error for a default on an unknown property, got nil")
	}
}
