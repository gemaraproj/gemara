// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"cuelang.org/go/cue"
)

// effective resolves a schema's properties and required set through allOf.
func effective(schemas map[string]any, name string, seen map[string]bool) (map[string]any, map[string]bool) {
	props, req := map[string]any{}, map[string]bool{}
	if seen[name] {
		return props, req
	}
	seen[name] = true
	s, _ := schemas[name].(map[string]any)
	for k, v := range mapOf(s["properties"]) {
		props[k] = v
	}
	for _, r := range sliceOf(s["required"]) {
		req[r.(string)] = true
	}
	for _, e := range sliceOf(s["allOf"]) {
		em, _ := e.(map[string]any)
		if ref, ok := em["$ref"].(string); ok {
			p, r := effective(schemas, strings.TrimPrefix(ref, refPrefix), seen)
			for k, v := range p {
				if _, ok := props[k]; !ok {
					props[k] = v
				}
			}
			for k := range r {
				req[k] = true
			}
			continue
		}
		for k, v := range mapOf(em["properties"]) {
			if _, ok := props[k]; !ok {
				props[k] = v
			}
		}
		for _, r := range sliceOf(em["required"]) {
			req[r.(string)] = true
		}
	}
	return props, req
}

// TestEveryCUEFieldReachesTheDocument is the invariant gemara#466, #468, #473
// and #474 all violate: a field the CUE declares must appear in the generated
// document, and a non-optional one must appear in an effective required set.
// The ground truth is the UNMODIFIED CUE, deliberately. Using loadPrepared here
// would compare post-rewrite CUE against a document generated from post-rewrite
// CUE: a pre-pass rewrite that dropped a real field would drop it from both
// sides at once and this gate would stay green. The pre-pass is the riskiest
// part of the change, so it does not get to supply its own ground truth.
func TestEveryCUEFieldReachesTheDocument(t *testing.T) {
	schemas, _ := loadGenerated(t)
	v, err := loadRaw("../../..")
	if err != nil {
		t.Fatal(err)
	}

	defsChecked, fieldsChecked := 0, 0
	defs, err := v.Fields(cue.Definitions(true))
	if err != nil {
		t.Fatalf("cannot enumerate the definitions of the unmodified CUE: %v", err)
	}
	for defs.Next() {
		sel := defs.Selector()
		if !sel.IsDefinition() {
			continue
		}
		name := strings.TrimPrefix(sel.String(), "#")
		if strings.HasPrefix(name, "_") {
			// Collapsed helper. These are also the only definitions that
			// evaluate to bottom in the unmodified CUE — their cross-field
			// conditionals leave a disjunction unresolved — which is why the
			// value below is only ever touched for a non-'_' name.
			continue
		}
		if _, ok := schemas[name]; !ok {
			t.Errorf("definition #%s is missing from the document", name)
			continue
		}
		if err := defs.Value().Err(); err != nil {
			t.Errorf("definition #%s does not evaluate in the unmodified CUE, so this gate cannot "+
				"check it: %v", name, err)
			continue
		}
		defsChecked++

		// Only struct definitions have fields to compare; enums and scalar
		// aliases are covered by the presence check above.
		if defs.Value().IncompleteKind() != cue.StructKind {
			continue
		}
		props, req := effective(schemas, name, map[string]bool{})

		fields, err := defs.Value().Fields(cue.Optional(true))
		if err != nil {
			// Skipping here would drop an entire definition from the gate
			// without a word, which is the failure mode this whole gate exists
			// to catch. Fail instead.
			t.Errorf("cannot enumerate the fields of #%s in the unmodified CUE, so none of them "+
				"is checked against the document: %v", name, err)
			continue
		}
		for fields.Next() {
			field := fields.Selector().Unquoted()
			if strings.HasPrefix(field, "_") {
				continue
			}
			fieldsChecked++
			spec, ok := props[field]
			if !ok {
				t.Errorf("%s.%s is declared in CUE but missing from the document", name, field)
				continue
			}
			// A field with a default always has a value in CUE, so it reads as
			// non-optional here — but on the wire the producer may omit it and
			// the consumer applies the default. Requiring it rejects documents
			// the CUE accepts, so the projection drops it from `required` and
			// carries the default instead. Both halves are gated: dropping it
			// from `required` without publishing the default would leave
			// consumers with no way to know the fallback.
			if d, has := fields.Value().Default(); has {
				if _, scalar, err := scalarDefault(d); err == nil && scalar {
					if req[field] {
						t.Errorf("%s.%s has a default in CUE but is required in the document; a "+
							"producer that omits it would be rejected", name, field)
					}
					if _, ok := mapOf(spec)["default"]; !ok {
						t.Errorf("%s.%s has a default in CUE that did not reach the document", name, field)
					}
					continue
				}
			}
			if !fields.IsOptional() && !req[field] {
				t.Errorf("%s.%s is required in CUE but not in the document", name, field)
			}
		}
	}

	// Observed on the real schema: 90 definitions, 347 fields. Floors are
	// roughly half that, rounded down to a tidy number — enough to catch a
	// collapse to zero or near-zero without pinning an exact count.
	t.Logf("checked %d definitions, %d fields", defsChecked, fieldsChecked)
	const minDefs, minFields = 40, 150
	if defsChecked < minDefs {
		t.Fatalf("gate checked only %d definitions; it is not exercising the schema and would pass vacuously", defsChecked)
	}
	if fieldsChecked < minFields {
		t.Fatalf("gate checked only %d fields; it is not exercising the schema and would pass vacuously", fieldsChecked)
	}
}

// TestEveryDocumentedFieldKeepsItsDescription gates the one place this
// migration can regress: the encoder cannot hang a description off a $ref, so
// restoreRefDescriptions has to put it back.
func TestEveryDocumentedFieldKeepsItsDescription(t *testing.T) {
	schemas, _ := loadGenerated(t)
	// Unmodified CUE for the same reason as the gate above: the doc comments
	// the document must carry are the ones the source declares, not the ones
	// that survived the pre-pass.
	v, err := loadRaw("../../..")
	if err != nil {
		t.Fatal(err)
	}
	docs, err := fieldDocs(v)
	if err != nil {
		t.Fatal(err)
	}
	docsChecked := 0
	for key, want := range docs {
		schema, field, _ := strings.Cut(key, ".")
		if strings.HasPrefix(schema, "_") {
			continue
		}
		s, ok := schemas[schema].(map[string]any)
		if !ok {
			continue
		}
		p, ok := mapOf(s["properties"])[field].(map[string]any)
		if !ok {
			continue
		}
		docsChecked++
		if got, _ := p["description"].(string); got == "" {
			t.Errorf("%s lost its description (CUE has %q)", key, want)
		}
	}

	// Observed on the real schema: 241 documented fields checked. The floor
	// is roughly half that, rounded down to a tidy number.
	t.Logf("checked %d documented fields", docsChecked)
	const minDocs = 100
	if docsChecked < minDocs {
		t.Fatalf("gate checked only %d documented fields; it is not exercising the schema and would pass vacuously", docsChecked)
	}
}

// OpenAPI 3.0 requires these keywords to be integers. JSON decoding turns every
// number into a float64, which YAML then renders as `1.0` — valid-looking and
// off-spec. This is the shape of bug the whole migration exists to prevent, so
// it gets a gate rather than a comment.
func TestIntegerKeywordsAreIntegers(t *testing.T) {
	schemas, data := loadGenerated(t)
	_ = schemas
	for _, kw := range []string{"minItems", "maxItems", "minLength", "maxLength", "minProperties", "maxProperties"} {
		if bytes.Contains(data, []byte(kw+": ")) {
			for _, line := range strings.Split(string(data), "\n") {
				trimmed := strings.TrimSpace(line)
				if !strings.HasPrefix(trimmed, kw+": ") {
					continue
				}
				val := strings.TrimPrefix(trimmed, kw+": ")
				if strings.Contains(val, ".") {
					t.Errorf("%s must be an integer, got %q", kw, val)
				}
			}
		}
	}
}

// knownRefSiblings is the explicit allowlist of $ref-with-siblings that are
// known and accepted today. OpenAPI 3.0 forbids siblings of $ref; the
// document this replaces had 55. One remains, tracked as a separate
// follow-up: the encoder emits a redundant `type: object` beside a $ref. Any
// OTHER sibling is a regression and must fail the build.
var knownRefSiblings = map[string]bool{
	"EnforcementLog.properties.actions.items": true,
}

func TestNoRefSiblings(t *testing.T) {
	schemas, _ := loadGenerated(t)
	found := map[string]bool{}
	var walk func(string, any)
	walk = func(path string, n any) {
		switch x := n.(type) {
		case map[string]any:
			if _, ok := x["$ref"]; ok && len(x) > 1 {
				found[path] = true
				if !knownRefSiblings[path] {
					t.Errorf("$ref carries siblings at %s: %#v", path, x)
				}
			}
			for k, v := range x {
				walk(path+"."+k, v)
			}
		case []any:
			for i, v := range x {
				walk(fmt.Sprintf("%s[%d]", path, i), v)
			}
		}
	}
	for name, s := range schemas {
		walk(name, s)
	}
	for path := range knownRefSiblings {
		if !found[path] {
			t.Errorf("allowlisted $ref sibling %s was not found; remove it from knownRefSiblings", path)
		}
	}
}
