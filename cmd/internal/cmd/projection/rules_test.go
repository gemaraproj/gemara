// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// A dropped conditional is recorded with the definition it belongs to and the
// comment that describes it: its own, or that of the hidden field it is the
// value of, but never the definition's, which describes the whole schema.
func TestPrepareRecordsDroppedRuleOwnerAndDoc(t *testing.T) {
	_, info := mustPrepare(t, `package p
// Lexicon is a glossary
#Lexicon: {
	terms: [...{id: string}]

	// term ids must be unique
	_uniqueTermIds: {for i, t in terms {(t.id): i}}
}
// Requirement is one requirement
#Requirement: {
	state: string

	// retired requirements must not have a recommendation
	if state == "Retired" {
		recommendation?: _|_
	}
}
// Log is a log
#Log: {
	results?: [...{id: string}]
	if results != _|_ {
		_uniqueResultIds: {for i, r in results {(r.id): i}}
	}
}
`)
	var got []droppedConditional
	for _, d := range info.DroppedConditionals {
		got = append(got, droppedConditional{owner: d.owner, doc: d.doc})
	}
	want := []droppedConditional{
		{owner: "Lexicon", doc: "term ids must be unique"},
		{owner: "Requirement", doc: "retired requirements must not have a recommendation"},
		{owner: "Log", doc: ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dropped = %+v, want %+v", got, want)
	}
}

func TestApplyRuleList(t *testing.T) {
	newDoc := func() *openapi3.T {
		return &openapi3.T{Components: &openapi3.Components{Schemas: openapi3.Schemas{
			"Log":  {Value: &openapi3.Schema{}},
			"Base": {Value: &openapi3.Schema{}},
			"_BaseStrict": {Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{
				{Ref: "#/components/schemas/Base"},
			}}},
		}}}
	}

	t.Run("lists each rule on its owner's schema, in order", func(t *testing.T) {
		doc := newDoc()
		err := applyRuleListTyped(doc, []droppedConditional{
			{pos: "log.cue:3:2", owner: "Log", doc: "result ids must be unique"},
			{pos: "log.cue:9:2", owner: "Log", doc: "each result names a declared criterion"},
		})
		if err != nil {
			t.Fatal(err)
		}
		got := doc.Components.Schemas["Log"].Value.Extensions[rulesExtension]
		want := []string{"result ids must be unique", "each result names a declared criterion"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %v, want %v", rulesExtension, got, want)
		}
	})

	t.Run("a hidden helper's rule goes on its base", func(t *testing.T) {
		doc := newDoc()
		err := applyRuleListTyped(doc, []droppedConditional{
			{pos: "base.cue:5:2", owner: "_BaseStrict", doc: "a rule the helper cannot project"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := doc.Components.Schemas["Base"].Value.Extensions[rulesExtension]; !reflect.DeepEqual(got, []string{"a rule the helper cannot project"}) {
			t.Errorf("Base %s = %v", rulesExtension, got)
		}
	})

	for _, tc := range []struct {
		name    string
		dropped droppedConditional
		wantErr string
	}{
		{"no comment", droppedConditional{pos: "log.cue:3:2", owner: "Log"}, "describe it in a // comment"},
		{"outside a definition", droppedConditional{pos: "log.cue:3:2", doc: "x"}, "outside any definition"},
		{"owner not emitted", droppedConditional{pos: "log.cue:3:2", owner: "Missing", doc: "x"}, "not in the generated document"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := applyRuleListTyped(newDoc(), []droppedConditional{tc.dropped})
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.dropped.pos) {
				t.Errorf("err %q does not name the source position", err)
			}
		})
	}
}

// The comment a rule is listed with: its own, joined into one line, even when
// the enclosing hidden field has one too; never a trailing comment or the
// comment of a regular field, which describes that field.
func TestPrepareRuleDocSources(t *testing.T) {
	_, info := mustPrepare(t, `package p
#A: {
	x?: string
	// inner holds a nested value
	inner: {
		if x != _|_ {y: string}
	}
	// the first line
	// and the second
	if x != _|_ {z: string}
	if x == _|_ {w: string} // a trailing note
	// the field's comment
	_h: {
		// the rule's own comment
		for i, v in [1] {(v): i}
	}
}
`)
	var got []string
	for _, d := range info.DroppedConditionals {
		got = append(got, d.doc)
	}
	// A struct's own rules are recorded before those of structs nested in it.
	want := []string{"the first line and the second", "", "", "the rule's own comment"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("docs = %q, want %q", got, want)
	}
}

// A new rule without a comment must stop generation of a real document, not
// only make the unit fail.
func TestUndescribedRuleFailsGeneration(t *testing.T) {
	_, err := convertSource(t, `package p

// A is a thing
#A: {
	x?: string
	if x != _|_ {
		y: string
	}
}
`)
	if err == nil || !strings.Contains(err.Error(), "schema.cue:6:2") || !strings.Contains(err.Error(), "describe it in a // comment") {
		t.Fatalf("err = %v, want it to name schema.cue:6:2 and ask for a comment", err)
	}
}
