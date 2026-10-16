// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Post applies Gemara projection metadata and output normalizations.
func (c *Converter) Post() error {
	if c.doc == nil {
		return fmt.Errorf("convert a CUE package before postprocessing")
	}
	doc := c.doc
	if err := openDisjunctionsTyped(doc); err != nil {
		return err
	}
	if err := inclusiveRequiredDisjunctionsTyped(doc); err != nil {
		return err
	}
	// Order matters from here: collapseHelperSchemas rewrites `$ref` in place,
	// and restoreRefDescriptions moves a `$ref` inside an `allOf` where that
	// rewrite no longer reaches it. Collapsing second would leave dangling
	// refs to the deleted helper schemas.
	if err := collapseHelperSchemasTyped(doc); err != nil {
		return err
	}
	fieldDocumentation, err := fieldDocs(c.value)
	if err != nil {
		return err
	}
	definitionDocumentation, err := definitionDocs(c.value)
	if err != nil {
		return err
	}
	deprecations, err := fieldDeprecations(c.value)
	if err != nil {
		return err
	}
	restoreDefinitionDescriptionsTyped(doc, definitionDocumentation)
	if err := restoreRefDescriptionsTyped(doc, fieldDocumentation, deprecations); err != nil {
		return err
	}
	unprojectable, err := unprojectableFields(c.value)
	if err != nil {
		return err
	}
	// Must run after descriptions are in their final form: it appends to
	// text restoreRefDescriptionsTyped may have just set.
	if err := appendUnenforcedCaveatsTyped(doc, unprojectable); err != nil {
		return err
	}
	formats, err := definitionFormats(c.value)
	if err != nil {
		return err
	}
	if err := applyFormatsTyped(doc, formats); err != nil {
		return err
	}
	defaults, err := fieldDefaults(c.value)
	if err != nil {
		return err
	}
	if err := applyDefaultsTyped(doc, defaults); err != nil {
		return err
	}
	statuses, err := fileStatuses(c.instance)
	if err != nil {
		return err
	}
	applyStatusTyped(doc, schemaAbsFiles(c.value), statuses)
	return nil
}

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

// CUE CONVERTER WORKAROUND: unwrapTupleItems rewrites `items: [T]` to `items: T`. CUE's open list form
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

func mapOf(v any) map[string]any { m, _ := v.(map[string]any); return m }
func sliceOf(v any) []any        { s, _ := v.([]any); return s }

// appendDeprecationReason adds a "Deprecated: <reason>" note to text,
// separated by a blank line if text is non-empty, or standing alone if not.
func appendDeprecationReason(text, reason string) string {
	note := "Deprecated: " + reason
	if text == "" {
		return note
	}
	return text + "\n\n" + note
}

func walkOpenAPISchemas(doc *openapi3.T, fn func(*openapi3.Schema) error) error {
	seen := map[*openapi3.Schema]bool{}
	var walkRef func(*openapi3.SchemaRef) error
	walkRef = func(ref *openapi3.SchemaRef) error {
		if ref == nil || ref.Value == nil || seen[ref.Value] {
			return nil
		}
		s := ref.Value
		seen[s] = true
		for _, p := range s.Properties {
			if err := walkRef(p); err != nil {
				return err
			}
		}
		for _, branch := range append(append(s.AllOf, s.AnyOf...), s.OneOf...) {
			if err := walkRef(branch); err != nil {
				return err
			}
		}
		if err := walkRef(s.Items); err != nil {
			return err
		}
		return fn(s)
	}
	for _, schema := range doc.Components.Schemas {
		if err := walkRef(schema); err != nil {
			return err
		}
	}
	return nil
}

// CUE CONVERTER WORKAROUND: openDisjunctionsTyped converts an open CUE
// disjunction that the encoder renders as a oneOf containing an empty branch
// into OpenAPI's equivalent anyOf.
func openDisjunctionsTyped(doc *openapi3.T) error {
	return walkOpenAPISchemas(doc, func(s *openapi3.Schema) error {
		if len(s.OneOf) == 0 {
			return nil
		}
		for _, branch := range s.OneOf {
			if branch.Ref == "" && branch.Value != nil && reflect.DeepEqual(branch.Value, &openapi3.Schema{}) {
				if len(s.AnyOf) != 0 {
					return fmt.Errorf("schema has both an open oneOf and an anyOf")
				}
				s.AnyOf, s.OneOf = s.OneOf, nil
				return nil
			}
		}
		return nil
	})
}

// CUE CONVERTER WORKAROUND: inclusiveRequiredDisjunctionsTyped rewrites the CUE encoder's rendering of
// `X & (A | B)` "required-unless / at-least-one" rules. The encoder treats
// struct disjunction as exclusive: it emits `oneOf` with each branch wrapped
// in an allOf carrying a `not: {anyOf: [<other branches>]}` clause that
// forbids satisfying more than one branch. Our rules are inclusive (e.g. an
// Evidence entry may legitimately carry both payload and source), so any
// oneOf exhibiting that not:{anyOf} signature is rewritten to anyOf, with the
// exclusivity clause stripped from each branch. A oneOf with no such not
// clause is a genuine closed disjunction and is left untouched.
func inclusiveRequiredDisjunctionsTyped(doc *openapi3.T) error {
	return walkOpenAPISchemas(doc, func(s *openapi3.Schema) error {
		if len(s.OneOf) == 0 || !hasExclusivitySignature(s.OneOf) {
			return nil
		}
		for i, branch := range s.OneOf {
			s.OneOf[i] = stripExclusivity(branch)
		}
		s.AnyOf, s.OneOf = s.OneOf, nil
		return nil
	})
}

// hasExclusivitySignature reports whether at least one branch carries the
// encoder's mutual-exclusion clause: an allOf of [positive-schema,
// {not: {anyOf: [...]}}].
func hasExclusivitySignature(branches openapi3.SchemaRefs) bool {
	for _, branch := range branches {
		if exclusivityNotIndex(branch) >= 0 {
			return true
		}
	}
	return false
}

// exclusivityNotIndex returns the index within branch.Value.AllOf of the
// `not: {anyOf: [...]}` exclusivity clause, or -1 if branch does not have
// that shape.
func exclusivityNotIndex(branch *openapi3.SchemaRef) int {
	if branch == nil || branch.Ref != "" || branch.Value == nil {
		return -1
	}
	for i, item := range branch.Value.AllOf {
		if item.Ref == "" && item.Value != nil && item.Value.Not != nil && item.Value.Not.Value != nil &&
			len(item.Value.Not.Value.AnyOf) > 0 {
			return i
		}
	}
	return -1
}

// stripExclusivity removes a branch's exclusivity not clause, if present,
// unwrapping a now-single-element allOf down to its remaining positive
// schema. Branches without the exclusivity clause are returned unchanged.
func stripExclusivity(branch *openapi3.SchemaRef) *openapi3.SchemaRef {
	i := exclusivityNotIndex(branch)
	if i < 0 {
		return branch
	}
	remaining := append(branch.Value.AllOf[:i:i], branch.Value.AllOf[i+1:]...)
	if len(remaining) == 1 {
		return remaining[0]
	}
	branch.Value.AllOf = remaining
	return branch
}

// CUE CONVERTER WORKAROUND: collapseHelperSchemasTyped removes the hidden
// helper schemas that schemaName must retain to prevent an encoder failure.
func collapseHelperSchemasTyped(doc *openapi3.T) error {
	base := map[string]string{}
	for name, ref := range doc.Components.Schemas {
		if !strings.HasPrefix(name, "_") {
			continue
		}
		if ref.Value == nil {
			return fmt.Errorf("helper schema %s has no value", name)
		}
		baseName, err := typedPassthroughBase(name, ref.Value, doc.Components.Schemas)
		if err != nil {
			return err
		}
		base[name] = baseName
	}
	if len(base) == 0 {
		return nil
	}
	for name := range base {
		delete(doc.Components.Schemas, name)
	}
	return walkOpenAPISchemas(doc, func(s *openapi3.Schema) error {
		for _, ref := range append(append(s.AllOf, s.AnyOf...), s.OneOf...) {
			rewriteHelperRef(ref, base)
		}
		for _, ref := range s.Properties {
			rewriteHelperRef(ref, base)
		}
		rewriteHelperRef(s.Items, base)
		deduplicateRefs(&s.AllOf)
		return nil
	})
}

func typedPassthroughBase(name string, helper *openapi3.Schema, schemas openapi3.Schemas) (string, error) {
	baseName := ""
	for _, ref := range helper.AllOf {
		if ref.Ref == "" {
			continue
		}
		if baseName != "" {
			return "", fmt.Errorf("helper schema %s extends more than one schema; cannot collapse it", name)
		}
		baseName = strings.TrimPrefix(ref.Ref, refPrefix)
	}
	if baseName == "" {
		return "", fmt.Errorf("helper schema %s extends no named schema; cannot collapse it", name)
	}
	base := schemas[baseName]
	if base == nil || base.Value == nil {
		return "", fmt.Errorf("helper schema %s extends unknown schema %s", name, baseName)
	}
	props, required, err := typedResolved(baseName, schemas, map[string]bool{})
	if err != nil {
		return "", err
	}
	helperProps, helperRequired := inlineConstraints(helper)
	for property, spec := range helperProps {
		if got := props[property]; got == nil || !sameValidationSchema(got, spec) {
			return "", fmt.Errorf("helper schema %s constrains %s.%s beyond %s; collapsing it would drop that constraint", name, name, property, baseName)
		}
	}
	for property := range helperRequired {
		if !required[property] {
			return "", fmt.Errorf("helper schema %s requires %q, which %s does not; collapsing it would drop that constraint", name, property, baseName)
		}
	}
	return baseName, nil
}

// sameValidationSchema compares only JSON Schema validation keywords. Helpers
// commonly re-declare a base field to host a CUE conditional and do not repeat
// documentation or examples; those annotations do not narrow the instance set.
func sameValidationSchema(a, b *openapi3.SchemaRef) bool {
	if a == nil || b == nil || a.Ref != b.Ref {
		return false
	}
	if a.Value == nil || b.Value == nil {
		return a.Value == b.Value
	}
	left, right := *a.Value, *b.Value
	left.Title, right.Title = "", ""
	left.Description, right.Description = "", ""
	left.Default, right.Default = nil, nil
	left.Example, right.Example = nil, nil
	left.Examples, right.Examples = nil, nil
	left.ExternalDocs, right.ExternalDocs = nil, nil
	left.Extensions, right.Extensions = nil, nil
	return reflect.DeepEqual(left, right)
}

func typedResolved(name string, schemas openapi3.Schemas, seen map[string]bool) (openapi3.Schemas, map[string]bool, error) {
	if seen[name] {
		return nil, nil, fmt.Errorf("cyclic allOf chain through %s", name)
	}
	seen[name] = true
	ref := schemas[name]
	if ref == nil || ref.Value == nil {
		return nil, nil, fmt.Errorf("unknown schema %s", name)
	}
	props, required := inlineConstraints(ref.Value)
	for _, item := range ref.Value.AllOf {
		if item.Ref == "" {
			continue
		}
		nestedProps, nestedRequired, err := typedResolved(strings.TrimPrefix(item.Ref, refPrefix), schemas, seen)
		if err != nil {
			return nil, nil, err
		}
		for n, p := range nestedProps {
			if props[n] == nil {
				props[n] = p
			}
		}
		for p := range nestedRequired {
			required[p] = true
		}
	}
	return props, required, nil
}

// inlineConstraints collects the properties and required names a schema
// declares itself, including those in inline allOf branches: the encoder
// renders `#Base & {...}` as allOf [$ref, {<extra constraints>}], so reading
// only the top level would miss them. Named $ref branches are left to the
// caller.
func inlineConstraints(s *openapi3.Schema) (openapi3.Schemas, map[string]bool) {
	props := openapi3.Schemas{}
	required := map[string]bool{}
	var collect func(*openapi3.Schema)
	collect = func(s *openapi3.Schema) {
		for n, p := range s.Properties {
			if props[n] == nil {
				props[n] = p
			}
		}
		for _, p := range s.Required {
			required[p] = true
		}
		for _, item := range s.AllOf {
			if item.Ref == "" && item.Value != nil {
				collect(item.Value)
			}
		}
	}
	collect(s)
	return props, required
}

func rewriteHelperRef(ref *openapi3.SchemaRef, base map[string]string) {
	if ref != nil {
		if b := base[strings.TrimPrefix(ref.Ref, refPrefix)]; b != "" {
			ref.Ref = refPrefix + b
		}
	}
}
func deduplicateRefs(refs *openapi3.SchemaRefs) {
	seen := map[string]bool{}
	out := (*refs)[:0:0]
	for _, ref := range *refs {
		if ref.Ref != "" && seen[ref.Ref] {
			continue
		}
		seen[ref.Ref] = true
		out = append(out, ref)
	}
	*refs = out
}

// CUE CONVERTER WORKAROUND: restoreRefDescriptionsTyped wraps references in
// allOf so field documentation and deprecation metadata survive the encoder.
func restoreRefDescriptionsTyped(doc *openapi3.T, docs, deprecations map[string]string) error {
	for name, ref := range doc.Components.Schemas {
		if ref.Value == nil {
			continue
		}
		for property, field := range ref.Value.Properties {
			if field == nil {
				continue
			}
			reason, deprecated := deprecations[name+"."+property]
			if field.Ref == "" {
				if field.Value != nil {
					if field.Value.Description == "" {
						field.Value.Description = docs[name+"."+property]
					}
					if deprecated {
						field.Value.Deprecated = true
						if reason != "" {
							field.Value.Description = appendDeprecationReason(field.Value.Description, reason)
						}
					}
				}
				continue
			}
			text := docs[name+"."+property]
			if deprecated && reason != "" {
				text = appendDeprecationReason(text, reason)
			}
			if text == "" && !deprecated {
				continue
			}
			ref.Value.Properties[property] = &openapi3.SchemaRef{Value: &openapi3.Schema{Description: text, Deprecated: deprecated, AllOf: openapi3.SchemaRefs{field}}}
		}
	}
	return nil
}

// restoreDefinitionDescriptionsTyped restores top-level definition docs the
// encoder omitted without replacing descriptions it already emitted.
func restoreDefinitionDescriptionsTyped(doc *openapi3.T, docs map[string]string) {
	for name, ref := range doc.Components.Schemas {
		if ref == nil || ref.Value == nil || ref.Value.Description != "" {
			continue
		}
		ref.Value.Description = docs[name]
	}
}

func applyFormatsTyped(doc *openapi3.T, formats map[string]string) error {
	for name, format := range formats {
		ref := doc.Components.Schemas[name]
		if ref == nil || ref.Value == nil {
			return fmt.Errorf("@gemara format %q was declared for definition #%s, but no schema %s exists in the generated document", format, name, name)
		}
		ref.Value.Format = format
	}
	return nil
}
func applyDefaultsTyped(doc *openapi3.T, defaults map[string]any) error {
	for key, value := range defaults {
		name, property, _ := strings.Cut(key, ".")
		if strings.HasPrefix(name, "_") {
			continue
		}
		holder := declaringSchema(doc.Components.Schemas, name, property)
		if holder == nil || holder.Properties[property].Value == nil {
			return fmt.Errorf("a default was recovered for %s, but that property is not in the generated schema", key)
		}
		holder.Properties[property].Value.Default = value
		required := holder.Required[:0:0]
		for _, item := range holder.Required {
			if item != property {
				required = append(required, item)
			}
		}
		holder.Required = required
	}
	return nil
}

// declaringSchema returns the schema that holds property on behalf of the
// named schema: the schema itself, one of its inline allOf branches, or a
// definition it embeds through an allOf $ref. CUE reports a field of an
// embedded definition on every definition that embeds it, but the encoder
// emits the field only on the embedded schema, so a directive keyed to the
// embedding definition must be applied there. The embedded definition's own
// key applies it too; both appliers are idempotent.
func declaringSchema(schemas openapi3.Schemas, name, property string) *openapi3.Schema {
	ref := schemas[name]
	if ref == nil || ref.Value == nil {
		return nil
	}
	seen := map[string]bool{name: true}
	var find func(*openapi3.Schema) *openapi3.Schema
	find = func(s *openapi3.Schema) *openapi3.Schema {
		if s.Properties[property] != nil {
			return s
		}
		for _, item := range s.AllOf {
			next := item.Value
			if item.Ref != "" {
				base := strings.TrimPrefix(item.Ref, refPrefix)
				if seen[base] || schemas[base] == nil {
					continue
				}
				seen[base] = true
				next = schemas[base].Value
			}
			if next == nil {
				continue
			}
			if found := find(next); found != nil {
				return found
			}
		}
		return nil
	}
	return find(ref.Value)
}

// CUE CONVERTER WORKAROUND: unenforcedCaveat is appended to a description that still asserts a
// cross-field/conditional CUE rule the generated OpenAPI schema does not
// (and, given the current upstream CUE OpenAPI encoder, cannot) enforce.
// Without it, the projection would keep making a promise its consumers
// cannot verify from the schema alone.
const unenforcedCaveat = "(Enforced by the CUE schema; not by this OpenAPI projection.)"

// appendUnenforcedCaveatsTyped appends unenforcedCaveat to every field marked
// @gemara(projectable=false) in the CUE source. An annotated field missing
// from the generated document is an error: continuing would hide a drift
// between the source declaration and its projection.
func appendUnenforcedCaveatsTyped(doc *openapi3.T, unprojectable map[string]bool) error {
	for key := range unprojectable {
		name, property, ok := strings.Cut(key, ".")
		if !ok {
			return fmt.Errorf("malformed unprojectable field key %q", key)
		}
		schema := doc.Components.Schemas[name]
		if schema == nil || schema.Value == nil {
			return fmt.Errorf("unprojectable field %s names schema %s, which is not in the generated document", key, name)
		}
		var field *openapi3.SchemaRef
		if holder := declaringSchema(doc.Components.Schemas, name, property); holder != nil {
			field = holder.Properties[property]
		}
		if field == nil || field.Value == nil {
			return fmt.Errorf("unprojectable field %s is not a property of %s in the generated document", key, name)
		}
		if strings.Contains(field.Value.Description, unenforcedCaveat) {
			continue
		}
		if field.Value.Description == "" {
			field.Value.Description = unenforcedCaveat
			continue
		}
		field.Value.Description += "\n\n" + unenforcedCaveat
	}
	return nil
}

func applyStatusTyped(doc *openapi3.T, absFiles, statuses map[string]string) {
	for name, ref := range doc.Components.Schemas {
		if ref != nil && ref.Value != nil {
			if status := statuses[absFiles[name]]; status != "" {
				if ref.Value.Extensions == nil {
					ref.Value.Extensions = map[string]any{}
				}
				ref.Value.Extensions["x-status"] = status
			}
		}
	}
}
