// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// ref builds an unresolved $ref SchemaRef.
func ref(name string) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Ref: refPrefix + name}
}

// val wraps a schema value in a SchemaRef, as the encoder does for inline branches.
func val(s *openapi3.Schema) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: s}
}

func docWithSchema(name string, s *openapi3.Schema) *openapi3.T {
	return &openapi3.T{Components: &openapi3.Components{Schemas: openapi3.Schemas{name: val(s)}}}
}

// evidenceShapeOneOf reproduces the encoder's rendering of `Evidence & (A | B)`
// where A requires "payload" and B requires "source": each branch is wrapped
// in an allOf with a `not: {anyOf: [<other branch>]}` exclusivity clause.
func evidenceShapeOneOf() *openapi3.Schema {
	sourceBranch := &openapi3.Schema{
		Properties: openapi3.Schemas{"source": ref("EvidenceMapping")},
		Required:   []string{"source"},
	}
	payloadBranch := &openapi3.Schema{Required: []string{"payload"}}
	return &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			val(&openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					val(payloadBranch),
					val(&openapi3.Schema{Not: val(&openapi3.Schema{AnyOf: openapi3.SchemaRefs{val(sourceBranch)}})}),
				},
			}),
			val(&openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					val(sourceBranch),
					val(&openapi3.Schema{Not: val(&openapi3.Schema{AnyOf: openapi3.SchemaRefs{val(payloadBranch)}})}),
				},
			}),
		},
	}
}

// mappingShapeOneOf reproduces the encoder's rendering for a rule where one
// branch is plain (no `not`) and the other carries the exclusivity `not`.
func mappingShapeOneOf() *openapi3.Schema {
	noMatchBranch := &openapi3.Schema{
		Properties: openapi3.Schemas{"relationship": val(&openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"no-match"}})},
		Required:   []string{"relationship"},
	}
	targetsBranch := &openapi3.Schema{
		Properties: openapi3.Schemas{"targets": val(&openapi3.Schema{Type: &openapi3.Types{"array"}, Items: ref("MappingTarget")})},
		Required:   []string{"targets"},
	}
	return &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			val(noMatchBranch),
			val(&openapi3.Schema{
				AllOf: openapi3.SchemaRefs{
					val(targetsBranch),
					val(&openapi3.Schema{Not: val(&openapi3.Schema{AnyOf: openapi3.SchemaRefs{val(noMatchBranch)}})}),
				},
			}),
		},
	}
}

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

func TestInclusiveRequiredDisjunctionsTypedEvidenceShape(t *testing.T) {
	doc := docWithSchema("Item", evidenceShapeOneOf())
	if err := inclusiveRequiredDisjunctionsTyped(doc); err != nil {
		t.Fatal(err)
	}
	s := doc.Components.Schemas["Item"].Value
	if len(s.OneOf) != 0 {
		t.Errorf("OneOf survived: %#v", s.OneOf)
	}
	if len(s.AnyOf) != 2 {
		t.Fatalf("AnyOf = %d branches, want 2: %#v", len(s.AnyOf), s.AnyOf)
	}
	for _, branch := range s.AnyOf {
		if len(branch.Value.AllOf) != 0 {
			t.Errorf("branch retained its exclusivity allOf wrapper: %#v", branch.Value)
		}
		if branch.Value.Not != nil {
			t.Errorf("branch retained its exclusivity not clause: %#v", branch.Value)
		}
	}
	if !slicesContainString(s.AnyOf[0].Value.Required, "payload") {
		t.Errorf("first branch lost its required:[payload]: %#v", s.AnyOf[0].Value)
	}
	if !slicesContainString(s.AnyOf[1].Value.Required, "source") {
		t.Errorf("second branch lost its required:[source]: %#v", s.AnyOf[1].Value)
	}
}

func TestInclusiveRequiredDisjunctionsTypedMappingShape(t *testing.T) {
	doc := docWithSchema("Item", mappingShapeOneOf())
	if err := inclusiveRequiredDisjunctionsTyped(doc); err != nil {
		t.Fatal(err)
	}
	s := doc.Components.Schemas["Item"].Value
	if len(s.OneOf) != 0 {
		t.Errorf("OneOf survived: %#v", s.OneOf)
	}
	if len(s.AnyOf) != 2 {
		t.Fatalf("AnyOf = %d branches, want 2: %#v", len(s.AnyOf), s.AnyOf)
	}
	if !slicesContainString(s.AnyOf[0].Value.Required, "relationship") {
		t.Errorf("first branch lost its required:[relationship]: %#v", s.AnyOf[0].Value)
	}
	if s.AnyOf[1].Value.Not != nil || len(s.AnyOf[1].Value.AllOf) != 0 {
		t.Errorf("second branch retained its exclusivity wrapper: %#v", s.AnyOf[1].Value)
	}
	if !slicesContainString(s.AnyOf[1].Value.Required, "targets") {
		t.Errorf("second branch lost its required:[targets]: %#v", s.AnyOf[1].Value)
	}
}

// A oneOf without the encoder's not:{anyOf} exclusivity signature is a
// genuine closed disjunction (e.g. an open enum with untyped extension
// values) and must be left untouched.
func TestInclusiveRequiredDisjunctionsTypedLeavesPlainOneOfAlone(t *testing.T) {
	plain := &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			val(&openapi3.Schema{Required: []string{"a"}}),
			val(&openapi3.Schema{Required: []string{"b"}}),
		},
	}
	doc := docWithSchema("Item", plain)
	if err := inclusiveRequiredDisjunctionsTyped(doc); err != nil {
		t.Fatal(err)
	}
	s := doc.Components.Schemas["Item"].Value
	if len(s.OneOf) != 2 {
		t.Fatalf("OneOf was rewritten: %#v", s.OneOf)
	}
	if len(s.AnyOf) != 0 {
		t.Errorf("AnyOf unexpectedly populated: %#v", s.AnyOf)
	}
}

// caveatFixture builds a document with one annotated field and one ordinary
// field to prove only CUE-discovered unprojectable fields receive a caveat.
func caveatFixture() *openapi3.T {
	schemas := openapi3.Schemas{
		"Widget": val(&openapi3.Schema{
			Properties: openapi3.Schemas{
				"complete": val(&openapi3.Schema{Description: "fully projected"}),
				"partial":  val(&openapi3.Schema{Description: "partially projected"}),
			},
		}),
	}
	return &openapi3.T{Components: &openapi3.Components{Schemas: schemas}}
}

func TestAppendUnenforcedCaveatsTypedAppendsToUnprojectableField(t *testing.T) {
	doc := caveatFixture()
	if err := appendUnenforcedCaveatsTyped(doc, map[string]bool{"Widget.partial": true}); err != nil {
		t.Fatal(err)
	}
	got := doc.Components.Schemas["Widget"].Value.Properties["partial"].Value.Description
	if !strings.Contains(got, unenforcedCaveat) {
		t.Errorf("description = %q, want it to contain %q", got, unenforcedCaveat)
	}
}

func TestAppendUnenforcedCaveatsTypedLeavesProjectableFieldsAlone(t *testing.T) {
	doc := caveatFixture()
	want := doc.Components.Schemas["Widget"].Value.Properties["complete"].Value.Description
	if err := appendUnenforcedCaveatsTyped(doc, map[string]bool{"Widget.partial": true}); err != nil {
		t.Fatal(err)
	}
	got := doc.Components.Schemas["Widget"].Value.Properties["complete"].Value.Description
	if got != want {
		t.Errorf("description = %q, want unchanged %q", got, want)
	}
}

func TestAppendUnenforcedCaveatsTypedIsIdempotent(t *testing.T) {
	doc := caveatFixture()
	unprojectable := map[string]bool{"Widget.partial": true}
	if err := appendUnenforcedCaveatsTyped(doc, unprojectable); err != nil {
		t.Fatal(err)
	}
	if err := appendUnenforcedCaveatsTyped(doc, unprojectable); err != nil {
		t.Fatal(err)
	}
	got := doc.Components.Schemas["Widget"].Value.Properties["partial"].Value.Description
	if n := strings.Count(got, unenforcedCaveat); n != 1 {
		t.Errorf("caveat appears %d times, want 1: %q", n, got)
	}
}

// An annotation naming a field absent from the generated document must fail
// loudly rather than silently caveating nothing.
func TestAppendUnenforcedCaveatsTypedErrorsOnMissingSchema(t *testing.T) {
	doc := docWithSchema("SomethingElse", &openapi3.Schema{})
	if err := appendUnenforcedCaveatsTyped(doc, map[string]bool{"Widget.partial": true}); err == nil {
		t.Error("expected an error for an unprojectable field with no matching schema, got nil")
	}
}

func slicesContainString(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

// The encoder places the extra constraints of `#Base & {...}` in an inline
// allOf branch, not on the helper itself. A requirement there that the base
// lacks must block the collapse rather than vanish with the helper.
func TestTypedPassthroughBaseRejectsInlineAllOfRequired(t *testing.T) {
	schemas := openapi3.Schemas{
		"Base": val(&openapi3.Schema{Properties: openapi3.Schemas{"id": val(&openapi3.Schema{Type: &openapi3.Types{"string"}})}}),
	}
	helper := &openapi3.Schema{AllOf: openapi3.SchemaRefs{
		ref("Base"),
		val(&openapi3.Schema{Required: []string{"foo"}}),
	}}
	_, err := typedPassthroughBase("_X", helper, schemas)
	if err == nil || !strings.Contains(err.Error(), `"foo"`) {
		t.Fatalf("err = %v, want an error naming the dropped requirement foo", err)
	}
}

func TestTypedPassthroughBaseRejectsInlineAllOfProperty(t *testing.T) {
	schemas := openapi3.Schemas{
		"Base": val(&openapi3.Schema{Properties: openapi3.Schemas{"id": val(&openapi3.Schema{Type: &openapi3.Types{"string"}})}}),
	}
	helper := &openapi3.Schema{AllOf: openapi3.SchemaRefs{
		ref("Base"),
		val(&openapi3.Schema{Properties: openapi3.Schemas{"id": val(&openapi3.Schema{Type: &openapi3.Types{"string"}, MinLength: 3})}}),
	}}
	_, err := typedPassthroughBase("_X", helper, schemas)
	if err == nil || !strings.Contains(err.Error(), "_X.id") {
		t.Fatalf("err = %v, want an error naming the narrowed property _X.id", err)
	}
}

// A requirement the base itself carries in an inline branch is not new, so it
// must not block the collapse.
func TestTypedPassthroughBaseAcceptsRequirementFromBaseInlineAllOf(t *testing.T) {
	schemas := openapi3.Schemas{
		"Base": val(&openapi3.Schema{AllOf: openapi3.SchemaRefs{
			val(&openapi3.Schema{Required: []string{"id"}}),
		}}),
	}
	helper := &openapi3.Schema{AllOf: openapi3.SchemaRefs{
		ref("Base"),
		val(&openapi3.Schema{Required: []string{"id"}}),
	}}
	got, err := typedPassthroughBase("_X", helper, schemas)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Base" {
		t.Errorf("base = %q, want Base", got)
	}
}

// TestRestoreRefDescriptionsAppliesDeprecations verifies that @deprecated attributes
// (both bare and with reasons) reach the projected OpenAPI schema properties,
// for both inline properties and wrapped $ref properties.
func TestRestoreRefDescriptionsAppliesDeprecations(t *testing.T) {
	doc := &openapi3.T{
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"Widget": val(&openapi3.Schema{
					Properties: openapi3.Schemas{
						"inlineBare": val(&openapi3.Schema{
							Type:        &openapi3.Types{"string"},
							Description: "Original description",
						}),
						"inlineWithReason": val(&openapi3.Schema{
							Type:        &openapi3.Types{"string"},
							Description: "Original description",
						}),
						"refBare":       ref("OtherSchema"),
						"refWithReason": ref("OtherSchema"),
						"notDeprecated": val(&openapi3.Schema{
							Type:        &openapi3.Types{"string"},
							Description: "Active field",
						}),
					},
				}),
			},
		},
	}

	docs := map[string]string{
		"Widget.inlineBare":       "Original description",
		"Widget.inlineWithReason": "Original description",
		"Widget.refBare":          "Doc comment for refBare",
		"Widget.refWithReason":    "Doc comment for refWithReason",
	}

	deprecations := map[string]string{
		"Widget.inlineBare":       "",
		"Widget.inlineWithReason": "use newField instead",
		"Widget.refBare":          "",
		"Widget.refWithReason":    "use otherRef instead",
	}

	if err := restoreRefDescriptionsTyped(doc, docs, deprecations); err != nil {
		t.Fatalf("restoreRefDescriptionsTyped failed: %v", err)
	}

	widgetProps := doc.Components.Schemas["Widget"].Value.Properties

	// 1. Inline bare deprecation
	inlineBare := widgetProps["inlineBare"]
	if !inlineBare.Value.Deprecated {
		t.Errorf("inlineBare.Deprecated = false, want true")
	}
	if inlineBare.Value.Description != "Original description" {
		t.Errorf("inlineBare.Description = %q, want %q", inlineBare.Value.Description, "Original description")
	}

	// 2. Inline deprecation with reason
	inlineReason := widgetProps["inlineWithReason"]
	if !inlineReason.Value.Deprecated {
		t.Errorf("inlineWithReason.Deprecated = false, want true")
	}
	wantInlineDesc := "Original description\n\nDeprecated: use newField instead"
	if inlineReason.Value.Description != wantInlineDesc {
		t.Errorf("inlineWithReason.Description = %q, want %q", inlineReason.Value.Description, wantInlineDesc)
	}

	// 3. Wrapped $ref bare deprecation
	refBare := widgetProps["refBare"]
	if !refBare.Value.Deprecated {
		t.Errorf("refBare.Deprecated = false, want true")
	}
	if refBare.Value.Description != "Doc comment for refBare" {
		t.Errorf("refBare.Description = %q, want %q", refBare.Value.Description, "Doc comment for refBare")
	}
	if len(refBare.Value.AllOf) != 1 || refBare.Value.AllOf[0].Ref != refPrefix+"OtherSchema" {
		t.Errorf("refBare.AllOf = %v, want allOf with ref OtherSchema", refBare.Value.AllOf)
	}

	// 4. Wrapped $ref deprecation with reason
	refReason := widgetProps["refWithReason"]
	if !refReason.Value.Deprecated {
		t.Errorf("refWithReason.Deprecated = false, want true")
	}
	wantRefDesc := "Doc comment for refWithReason\n\nDeprecated: use otherRef instead"
	if refReason.Value.Description != wantRefDesc {
		t.Errorf("refWithReason.Description = %q, want %q", refReason.Value.Description, wantRefDesc)
	}
	if len(refReason.Value.AllOf) != 1 || refReason.Value.AllOf[0].Ref != refPrefix+"OtherSchema" {
		t.Errorf("refWithReason.AllOf = %v, want allOf with ref OtherSchema", refReason.Value.AllOf)
	}

	// 5. Not deprecated field
	notDep := widgetProps["notDeprecated"]
	if notDep.Value.Deprecated {
		t.Errorf("notDeprecated.Deprecated = true, want false")
	}
	if notDep.Value.Description != "Active field" {
		t.Errorf("notDeprecated.Description = %q, want %q", notDep.Value.Description, "Active field")
	}
}
