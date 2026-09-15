// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"reflect"
	"strings"
)

// walkSchemas calls fn on every map node in doc, depth first. fn may modify
// the map in place; returning an error stops the walk.
func walkSchemas(doc map[string]any, fn func(map[string]any) error) error {
	var walk func(any) error
	walk = func(n any) error {
		switch x := n.(type) {
		case map[string]any:
			for _, v := range x {
				if err := walk(v); err != nil {
					return err
				}
			}
			return fn(x)
		case []any:
			for _, v := range x {
				if err := walk(v); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(doc)
}

// unwrapTupleItems rewrites `items: [T]` to `items: T`. CUE's open list form
// [#X, ...#X] encodes as a one-element tuple plus minItems: 1, and OpenAPI 3.0
// has no tuple form. A longer tuple is a per-item schema OpenAPI cannot
// express at all, so it is an error rather than a silent truncation.
func unwrapTupleItems(doc map[string]any) error {
	return walkSchemas(doc, func(m map[string]any) error {
		items, ok := m["items"].([]any)
		if !ok {
			return nil
		}
		if len(items) != 1 {
			return fmt.Errorf("cannot represent a %d-element tuple schema in OpenAPI 3.0: %v", len(items), items)
		}
		m["items"] = items[0]
		return nil
	})
}

const refPrefix = "#/components/schemas/"

// collapseHelperSchemas removes hidden `_…` schemas that add nothing over the
// single schema they extend, rewriting refs to point at that base instead.
//
// These come from CUE's #_… definitions, which exist only to carry the
// cross-field conditionals prepare() strips. Once those are gone the helper is
// a pass-through, and today's document already refs the base directly.
//
// A helper that is NOT provably redundant is an error: dropping it would
// silently weaken the schema, and keeping it would leak an internal name.
func collapseHelperSchemas(doc map[string]any) error {
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	if schemas == nil {
		return nil
	}

	base := map[string]string{} // helper name → base name
	for name, raw := range schemas {
		if !strings.HasPrefix(name, "_") {
			continue
		}
		s, _ := raw.(map[string]any)
		b, err := passthroughBase(name, s, schemas)
		if err != nil {
			return err
		}
		base[name] = b
	}
	if len(base) == 0 {
		return nil
	}

	for name := range base {
		delete(schemas, name)
	}
	if err := walkSchemas(doc, func(m map[string]any) error {
		ref, ok := m["$ref"].(string)
		if !ok {
			return nil
		}
		if b, ok := base[strings.TrimPrefix(ref, refPrefix)]; ok {
			m["$ref"] = refPrefix + b
		}
		return nil
	}); err != nil {
		return err
	}
	return walkSchemas(doc, func(m map[string]any) error {
		all, ok := m["allOf"].([]any)
		if !ok {
			return nil
		}
		seen := map[string]bool{}
		out := all[:0:0]
		for _, e := range all {
			em, _ := e.(map[string]any)
			if ref, ok := em["$ref"].(string); ok {
				if seen[ref] {
					continue
				}
				seen[ref] = true
			}
			out = append(out, e)
		}
		m["allOf"] = out
		return nil
	})
}

// passthroughBase returns the name of the single schema `helper` extends,
// provided every property and requirement it declares is already satisfied
// there identically.
func passthroughBase(name string, helper, schemas map[string]any) (string, error) {
	all, _ := helper["allOf"].([]any)
	var baseName string
	required := map[string]bool{}
	for _, e := range all {
		em, _ := e.(map[string]any)
		if ref, ok := em["$ref"].(string); ok {
			if baseName != "" {
				return "", fmt.Errorf("helper schema %s extends more than one schema; cannot collapse it", name)
			}
			baseName = strings.TrimPrefix(ref, refPrefix)
			continue
		}
		for _, r := range sliceOf(em["required"]) {
			required[r.(string)] = true
		}
	}
	if baseName == "" {
		return "", fmt.Errorf("helper schema %s extends no named schema; cannot collapse it", name)
	}
	if strings.HasPrefix(baseName, "_") {
		return "", fmt.Errorf("helper schema %s extends another helper schema %s; "+
			"chained helpers are not supported because collapsing them would leave a dangling $ref",
			name, baseName)
	}
	b, _ := schemas[baseName].(map[string]any)
	if b == nil {
		return "", fmt.Errorf("helper schema %s extends unknown schema %s", name, baseName)
	}

	baseProps, _ := b["properties"].(map[string]any)
	for prop, spec := range mapOf(helper["properties"]) {
		got, ok := baseProps[prop]
		if !ok || !reflect.DeepEqual(got, spec) {
			return "", fmt.Errorf("helper schema %s constrains %s.%s beyond %s; collapsing it would drop that constraint",
				name, name, prop, baseName)
		}
	}
	baseRequired := map[string]bool{}
	for _, r := range sliceOf(b["required"]) {
		baseRequired[r.(string)] = true
	}
	for r := range required {
		if !baseRequired[r] {
			return "", fmt.Errorf("helper schema %s requires %q, which %s does not; collapsing it would drop that constraint",
				name, r, baseName)
		}
	}
	return baseName, nil
}

func mapOf(v any) map[string]any { m, _ := v.(map[string]any); return m }
func sliceOf(v any) []any        { s, _ := v.([]any); return s }

// restoreRefDescriptions puts field doc comments back onto properties that
// lost them.
//
// For a bare $ref: OpenAPI 3.0 forbids siblings of $ref, so the encoder drops
// the description rather than emit an invalid document — 55 fields in the
// Gemara schema. The document this replaces kept them by emitting the
// sibling anyway, which does not validate. `{description, allOf: [{$ref}]}`
// is the form 3.0 permits and is what the website's reference pages read.
//
// For a non-$ref property with no description: this is an inline narrowing
// of a field inherited from an embedded definition (e.g.
// `metadata: type: "ControlCatalog"` over #Catalog's `metadata: #Metadata`),
// which the encoder renders with no description at all even though the CUE
// doc comment on the base field still applies. The description is filled in
// directly; an existing description is never overwritten.
func restoreRefDescriptions(doc map[string]any, docs map[string]string) error {
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	for schema, raw := range schemas {
		s, _ := raw.(map[string]any)
		for prop, praw := range mapOf(s["properties"]) {
			p, _ := praw.(map[string]any)
			if p == nil {
				continue
			}
			ref, isRef := p["$ref"].(string)
			if !isRef {
				// Not a $ref, so a description can sit directly on the schema.
				// The encoder omits one when the property is an inline
				// narrowing of a field inherited from an embedded definition
				// (e.g. `metadata: type: "ControlCatalog"` over #Catalog's
				// `metadata: #Metadata`). Fill it from the CUE doc comment.
				// An empty description is as absent as a missing key — and the
				// gate that checks this tests emptiness, so testing key
				// presence here would let an empty string through the fill and
				// fail the gate.
				if s, _ := p["description"].(string); s == "" {
					if text := docs[schema+"."+prop]; text != "" {
						p["description"] = text
					}
				}
				continue
			}
			// Prefer the CUE doc comment; fall back to any description already
			// present, which OpenAPI 3.0 does not permit beside a $ref.
			text := docs[schema+"."+prop]
			if text == "" {
				text, _ = p["description"].(string)
			}
			if text == "" {
				// Nothing to carry, and a bare $ref is already valid. Leave it.
				continue
			}
			// The $ref moves into allOf, so an allOf already sitting beside it
			// would be overwritten and its constraints lost. Never seen on the
			// Gemara schema; an error rather than a silent discard.
			if existing, has := p["allOf"]; has {
				return fmt.Errorf("%s.%s has both a $ref and an allOf (%v); moving the $ref into "+
					"allOf to carry its description would discard the existing one",
					schema, prop, existing)
			}
			delete(p, "$ref")
			delete(p, "description")
			p["description"] = text
			p["allOf"] = []any{map[string]any{"$ref": ref}}
		}
	}
	return nil
}

// timeLayoutFormats maps a Go time layout to the OpenAPI `format` describing
// the same value space. Add entries as the schema needs them; an unmapped
// layout is an error, never a narrower guess.
var timeLayoutFormats = map[string]string{
	"2006-01-02T15:04:05Z07:00": "date-time", // RFC 3339
	"2006-01-02":                "date",
}

// applyTimeFormats restores the `format` for each definition whose
// time.Format() call prepare() rewrote away. See the WORKAROUND note in
// prepass.go for why the call cannot survive to this point.
func applyTimeFormats(doc map[string]any, layouts map[string]string) error {
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	for name, layout := range layouts {
		format, ok := timeLayoutFormats[layout]
		if !ok {
			return fmt.Errorf("no OpenAPI format for time layout %q on %s; add it to timeLayoutFormats", layout, name)
		}
		s, _ := schemas[name].(map[string]any)
		if s == nil {
			// A layout was recorded against #name during the pre-pass, so the
			// document must have a schema to restore it onto. Its absence means
			// the layout was attributed to the wrong definition, or the schema
			// was dropped — either way the format is silently lost and the
			// value degrades to a bare string. This is the one check that can
			// catch such a misattribution, so it fails rather than skips.
			return fmt.Errorf("time layout %q was recorded for definition #%s, but no schema %s "+
				"exists in the generated document; its format would be silently dropped",
				layout, name, name)
		}
		s["format"] = format
	}
	return nil
}

// openDisjunctions rewrites `oneOf` to `anyOf` on a disjunction that has an
// unconstrained branch.
//
// CUE's `|` is not an exclusive or. `#EvidenceType: #ArtifactType | string`
// means "one of these recommended values, or any string", and the encoder
// renders it as `oneOf: [{enum: […]}, {}]`. In JSON Schema `oneOf` means
// EXACTLY one branch may match, and `{}` matches everything — so each
// recommended value matches both branches and is rejected, while an arbitrary
// string matches only `{}` and passes. The document ends up accepting
// precisely what the CUE meant to single out and rejecting the rest, which is
// how the repo's own good-audit-log.yaml (evidence[0].type: EvaluationLog)
// stopped validating.
//
// `anyOf` is the faithful reading of `|`, and keeps the recommended values in
// the document where the website's reference pages can show them. A closed
// disjunction — no unconstrained branch — is left alone: `oneOf` is not
// self-defeating there, and widening it would drop a real constraint.
func openDisjunctions(doc map[string]any) error {
	return walkSchemas(doc, func(m map[string]any) error {
		branches, ok := m["oneOf"].([]any)
		if !ok {
			return nil
		}
		open := false
		for _, b := range branches {
			if bm, ok := b.(map[string]any); ok && len(bm) == 0 {
				open = true
				break
			}
		}
		if !open {
			return nil
		}
		// Renaming the key would silently discard an existing anyOf. Never
		// seen on the Gemara schema; an error rather than a quiet drop.
		if existing, has := m["anyOf"]; has {
			return fmt.Errorf("schema has both an open oneOf and an anyOf (%v); rewriting the "+
				"oneOf would discard the anyOf", existing)
		}
		delete(m, "oneOf")
		m["anyOf"] = branches
		return nil
	})
}

// applyDefaults restores the defaults fieldDefaults recovered, and drops those
// fields from their schema's `required` list.
//
// A CUE field with a default always has a value, so the encoder emits it as
// required. On the wire the opposite is true: the producer may omit it and the
// consumer applies the default — which is exactly what
// `state: #Lifecycle @yaml("state,omitempty")` says. Leaving it required
// rejects documents the CUE accepts, as the repo's own good-ccc.yaml and
// good-policy.yaml fixtures do.
func applyDefaults(doc map[string]any, defaults map[string]any) error {
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	for key, val := range defaults {
		schema, prop, _ := strings.Cut(key, ".")
		// A hidden #_… definition is expected to be absent: collapseHelperSchemas
		// folded it into its base, which carries the same field and its default.
		if strings.HasPrefix(schema, "_") {
			continue
		}
		s, _ := schemas[schema].(map[string]any)
		if s == nil {
			return fmt.Errorf("a default was recovered for %s, but no schema %s exists in the "+
				"generated document; the default would be silently lost", key, schema)
		}
		p, _ := mapOf(s["properties"])[prop].(map[string]any)
		if p == nil {
			return fmt.Errorf("a default was recovered for %s, but that property is not in the "+
				"generated schema; the default would be silently lost", key)
		}
		p["default"] = val

		req, ok := s["required"].([]any)
		if !ok {
			continue
		}
		out := req[:0:0]
		for _, r := range req {
			if r != prop {
				out = append(out, r)
			}
		}
		if len(out) == 0 {
			delete(s, "required")
			continue
		}
		s["required"] = out
	}
	return nil
}
