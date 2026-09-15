// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/literal"
)

// fieldDocs collects the doc comment of every field of every definition,
// keyed "Schema.property". The encoder cannot attach these to a $ref, so
// restoreRefDescriptions puts them back in a form OpenAPI 3.0 permits.
// A definition whose fields cannot be enumerated is an error, not a skip: it
// contributes nothing to the map, so every description it should have carried
// is dropped from the document AND from the gate that checks descriptions
// survived — the failure hiding itself, which is what this migration exists to
// stop.
func fieldDocs(v cue.Value) (map[string]string, error) {
	out := map[string]string{}
	defs, err := v.Fields(cue.Definitions(true))
	if err != nil {
		return nil, fmt.Errorf("enumerating definitions for doc comments: %w", err)
	}
	var failed []string
	for defs.Next() {
		sel := defs.Selector()
		if !sel.IsDefinition() {
			continue
		}
		schema := strings.TrimPrefix(sel.String(), "#")
		// Only a struct has fields to document. Fields() on an enum or scalar
		// alias (#Severity, #Datetime, #Email …) errors with "cannot use value
		// … as struct"; that is not a failure to enumerate, it is a definition
		// with nothing to enumerate, so it is filtered by kind rather than by
		// swallowing the error below.
		if defs.Value().IncompleteKind() != cue.StructKind {
			continue
		}
		fields, err := defs.Value().Fields(cue.Optional(true))
		if err != nil {
			failed = append(failed, fmt.Sprintf("#%s: %v", schema, err))
			continue
		}
		for fields.Next() {
			if text := docText(fields.Value()); text != "" {
				out[schema+"."+fields.Selector().Unquoted()] = text
			}
		}
	}
	if len(failed) > 0 {
		sort.Strings(failed)
		return nil, fmt.Errorf("cannot enumerate the fields of %d definition(s), so their doc "+
			"comments would be silently dropped: %s", len(failed), strings.Join(failed, "; "))
	}
	return out, nil
}

// docText joins a value's doc comments exactly as the upstream encoder does in
// encoding/openapi's getDoc (build.go:222-236 in v0.15.4): raw comment text
// joined with a blank line, trimmed once at the end. Matching it matters
// because a restored description sits in the same document as encoder-emitted
// ones, and the two must not render differently.
func docText(v cue.Value) string {
	var parts []string
	for _, g := range v.Doc() {
		parts = append(parts, g.Text())
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

// fileStatuses reads the file-level @status("…") attribute of each CUE file,
// keyed by absolute filename. This replaces reading the first ten lines of
// each file as text.
func fileStatuses(inst *build.Instance) (map[string]string, error) {
	out := map[string]string{}
	for _, f := range inst.Files {
		for _, d := range f.Decls {
			a, ok := d.(*ast.Attribute)
			if !ok {
				continue
			}
			key, body := a.Split()
			if key != "status" {
				continue
			}
			s, err := literal.Unquote(body)
			if err != nil {
				return nil, fmt.Errorf("%s: malformed @status attribute %s: %w", f.Filename, body, err)
			}
			out[f.Filename] = s
			break
		}
	}
	return out, nil
}

// schemaFiles maps each definition's schema name to the base name of the file
// it was declared in.
func schemaFiles(v cue.Value) map[string]string {
	out := map[string]string{}
	it, _ := v.Fields(cue.Definitions(true))
	for it.Next() {
		sel := it.Selector()
		if !sel.IsDefinition() {
			continue
		}
		if fn := it.Value().Pos().Filename(); fn != "" {
			out[strings.TrimPrefix(sel.String(), "#")] = filepath.Base(fn)
		}
	}
	return out
}

// schemaAbsFiles is schemaFiles keyed to the absolute path, for joining
// against fileStatuses.
func schemaAbsFiles(v cue.Value) map[string]string {
	out := map[string]string{}
	it, _ := v.Fields(cue.Definitions(true))
	for it.Next() {
		sel := it.Selector()
		if !sel.IsDefinition() {
			continue
		}
		if fn := it.Value().Pos().Filename(); fn != "" {
			out[strings.TrimPrefix(sel.String(), "#")] = fn
		}
	}
	return out
}

// applyStatus stamps x-status on each schema from its declaring file's
// @status attribute.
func applyStatus(doc map[string]any, absFiles, statuses map[string]string) {
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	for name, raw := range schemas {
		s, _ := raw.(map[string]any)
		if s == nil {
			continue
		}
		if status := statuses[absFiles[name]]; status != "" {
			s["x-status"] = status
		}
	}
}

// buildManifest inverts schemaFiles into the filename → schema names map the
// website consumes, with names sorted for a stable diff.
func buildManifest(files map[string]string) map[string][]string {
	out := map[string][]string{}
	for name, file := range files {
		out[file] = append(out[file], name)
	}
	for _, names := range out {
		sort.Strings(names)
	}
	return out
}

// fieldDefaults recovers the default of every field that declares one, keyed
// "Schema.property".
//
// generateRaw strips `default` wholesale via FieldFilter, because the open-list
// form [#X, ...#X] produces a non-concrete one that makes openapi.Gen fail
// outright ("cannot convert incomplete value"). That filter is not selective,
// so it also removes the three real defaults the schema declares
// (#Lifecycle's "Active", and `required: *false | bool` in two places). They
// are read back from the CUE here and restored by applyDefaults.
func fieldDefaults(v cue.Value) (map[string]any, error) {
	out := map[string]any{}
	defs, err := v.Fields(cue.Definitions(true))
	if err != nil {
		return nil, fmt.Errorf("enumerating definitions for defaults: %w", err)
	}
	for defs.Next() {
		sel := defs.Selector()
		if !sel.IsDefinition() || defs.Value().IncompleteKind() != cue.StructKind {
			continue
		}
		schema := strings.TrimPrefix(sel.String(), "#")
		fields, err := defs.Value().Fields(cue.Optional(true))
		if err != nil {
			return nil, fmt.Errorf("enumerating the fields of #%s for defaults: %w", schema, err)
		}
		for fields.Next() {
			d, ok := fields.Value().Default()
			if !ok {
				continue
			}
			val, ok, err := scalarDefault(d)
			if err != nil {
				return nil, fmt.Errorf("#%s.%s: %w", schema, fields.Selector().Unquoted(), err)
			}
			if !ok {
				continue
			}
			out[schema+"."+fields.Selector().Unquoted()] = val
		}
	}
	return out, nil
}

// scalarDefault converts a CUE default to the Go value OpenAPI should carry,
// reporting false for anything that is not a scalar.
//
// A struct or list default is CUE's own bookkeeping — the open-list artefact
// FieldFilter exists to suppress — not a value a producer could omit and have
// filled in, so it is deliberately not projected.
func scalarDefault(d cue.Value) (any, bool, error) {
	switch d.Kind() {
	case cue.BoolKind:
		b, err := d.Bool()
		return b, err == nil, err
	case cue.StringKind:
		s, err := d.String()
		return s, err == nil, err
	case cue.IntKind:
		i, err := d.Int64()
		return i, err == nil, err
	case cue.FloatKind, cue.NumberKind:
		f, err := d.Float64()
		return f, err == nil, err
	}
	return nil, false, nil
}
