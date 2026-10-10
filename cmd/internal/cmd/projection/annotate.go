// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
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

// definitionDocs collects definition doc comments keyed by schema name. The
// encoder can omit these top-level descriptions, so the post-pass restores
// them for schemas it emits.
func definitionDocs(v cue.Value) (map[string]string, error) {
	out := map[string]string{}
	defs, err := v.Fields(cue.Definitions(true))
	if err != nil {
		return nil, fmt.Errorf("enumerating definitions for doc comments: %w", err)
	}
	for defs.Next() {
		sel := defs.Selector()
		if !sel.IsDefinition() {
			continue
		}
		if text := docText(defs.Value()); text != "" {
			out[strings.TrimPrefix(sel.String(), "#")] = text
		}
	}
	return out, nil
}

// fieldDeprecations collects the @deprecated attribute of every field of
// every definition, keyed "Schema.property". Presence of a key means the
// field is deprecated; its value is the attribute's optional reason argument
// (empty for a bare @deprecated()). The upstream encoder does not read this
// attribute (it only recognizes @protobuf(...,deprecated), which this schema
// has no reason to use), so restoreRefDescriptions applies it in the
// post-pass, the same way it restores descriptions the encoder drops.
func fieldDeprecations(v cue.Value) (map[string]string, error) {
	out := map[string]string{}
	defs, err := v.Fields(cue.Definitions(true))
	if err != nil {
		return nil, fmt.Errorf("enumerating definitions for deprecations: %w", err)
	}
	var failed []string
	for defs.Next() {
		sel := defs.Selector()
		if !sel.IsDefinition() {
			continue
		}
		schema := strings.TrimPrefix(sel.String(), "#")
		if defs.Value().IncompleteKind() != cue.StructKind {
			continue
		}
		fields, err := defs.Value().Fields(cue.Optional(true))
		if err != nil {
			failed = append(failed, fmt.Sprintf("#%s: %v", schema, err))
			continue
		}
		for fields.Next() {
			attr := fields.Value().Attribute("deprecated")
			if attr.Err() != nil {
				continue
			}
			reason := ""
			if attr.NumArgs() > 0 {
				reason, err = attr.String(0)
				if err != nil {
					return nil, fmt.Errorf("#%s.%s: malformed @deprecated attribute: %w",
						schema, fields.Selector().Unquoted(), err)
				}
			}
			out[schema+"."+fields.Selector().Unquoted()] = reason
		}
	}
	if len(failed) > 0 {
		sort.Strings(failed)
		return nil, fmt.Errorf("cannot enumerate the fields of %d definition(s), so their "+
			"deprecations would be silently dropped: %s", len(failed), strings.Join(failed, "; "))
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

// fileStatuses reads the file-level @gemara(status="…") attribute of each CUE file,
// keyed by absolute filename. This replaces reading the first ten lines of
// each file as text.
//
// Every file must declare a status. A file without one would have its
// definitions emitted without x-status, which breaking-check reads as stable,
// so a missing or legacy @status attribute is an error rather than a skip.
func fileStatuses(inst *build.Instance) (map[string]string, error) {
	out := map[string]string{}
	for _, f := range inst.Files {
		for _, d := range f.Decls {
			a, ok := d.(*ast.Attribute)
			if !ok {
				continue
			}
			key, _ := a.Split()
			if key == "status" {
				return nil, fmt.Errorf("%s: legacy @status attribute; use @gemara(status=\"…\")", f.Filename)
			}
			if key != "gemara" {
				continue
			}
			directive, err := parseFileGemaraDirective(a)
			if err != nil {
				return nil, fmt.Errorf("%s: malformed @gemara attribute: %w", f.Filename, err)
			}
			s := directive.status
			switch s {
			case "experimental", "stable", "deprecated":
			default:
				return nil, fmt.Errorf("%s: @gemara status must be experimental, stable, or deprecated, got %q", f.Filename, s)
			}
			out[f.Filename] = s
			break
		}
		if out[f.Filename] == "" {
			return nil, fmt.Errorf("%s: no file-level @gemara(status=\"…\") attribute", f.Filename)
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
// @gemara(status="...") attribute.
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

// fieldDefaults reads explicit @gemara(default=...) directives, keyed
// "Schema.property". The CUE default remains the source of validation behavior;
// this directive declares the corresponding wire-level OpenAPI default.
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
			attr := fields.Value().Attribute("gemara")
			if attr.Err() != nil {
				continue
			}
			directive, err := parseGemaraDirective(attr)
			if err != nil {
				return nil, fmt.Errorf("#%s.%s: malformed @gemara attribute: %w", schema, fields.Selector().Unquoted(), err)
			}
			if directive.defaultValue == nil {
				continue
			}
			semantic, hasDefault := fields.Value().Default()
			if !hasDefault {
				return nil, fmt.Errorf("#%s.%s: @gemara(default=...) requires a matching CUE default", schema, fields.Selector().Unquoted())
			}
			expected, scalar, err := scalarDefault(semantic)
			if err != nil || !scalar || fmt.Sprint(expected) != *directive.defaultValue {
				return nil, fmt.Errorf("#%s.%s: @gemara(default=%s) does not match its CUE default", schema, fields.Selector().Unquoted(), *directive.defaultValue)
			}
			out[schema+"."+fields.Selector().Unquoted()] = expected
		}
	}
	return out, nil
}

// unprojectableFields returns fields explicitly marked @gemara(projectable=false).
// They carry CUE constraints the OpenAPI projection cannot faithfully represent.
func unprojectableFields(v cue.Value) (map[string]bool, error) {
	out := map[string]bool{}
	defs, err := v.Fields(cue.Definitions(true))
	if err != nil {
		return nil, fmt.Errorf("enumerating definitions for projectability: %w", err)
	}
	for defs.Next() {
		sel := defs.Selector()
		if !sel.IsDefinition() || defs.Value().IncompleteKind() != cue.StructKind {
			continue
		}
		schema := strings.TrimPrefix(sel.String(), "#")
		fields, err := defs.Value().Fields(cue.Optional(true))
		if err != nil {
			return nil, fmt.Errorf("enumerating the fields of #%s for projectability: %w", schema, err)
		}
		for fields.Next() {
			attr := fields.Value().Attribute("gemara")
			if attr.Err() != nil {
				continue
			}
			directive, err := parseGemaraDirective(attr)
			if err != nil {
				return nil, fmt.Errorf("#%s.%s: malformed @gemara attribute: %w", schema, fields.Selector().Unquoted(), err)
			}
			if directive.projectable != nil && !*directive.projectable {
				out[schema+"."+fields.Selector().Unquoted()] = true
			}
		}
	}
	return out, nil
}

func definitionFormats(v cue.Value) (map[string]string, error) {
	out := map[string]string{}
	defs, err := v.Fields(cue.Definitions(true))
	if err != nil {
		return nil, fmt.Errorf("enumerating definitions for formats: %w", err)
	}
	for defs.Next() {
		if !defs.Selector().IsDefinition() {
			continue
		}
		attr := defs.Value().Attribute("gemara")
		if attr.Err() != nil {
			continue
		}
		directive, err := parseGemaraDirective(attr)
		if err != nil {
			return nil, fmt.Errorf("#%s: malformed @gemara attribute: %w", defs.Selector(), err)
		}
		if directive.format == nil {
			continue
		}
		out[strings.TrimPrefix(defs.Selector().String(), "#")] = *directive.format
	}
	return out, nil
}

type gemaraDirective struct {
	status       string
	format       *string
	defaultValue *string
	projectable  *bool
}

// parseGemaraDirective validates the closed projection-directive contract using
// CUE's parsed attribute arguments, not source-text splitting.
func parseGemaraDirective(attr cue.Attribute) (gemaraDirective, error) {
	if err := attr.Err(); err != nil {
		return gemaraDirective{}, err
	}
	var directive gemaraDirective
	for i := range attr.NumArgs() {
		key, value := attr.Arg(i)
		if value == "" {
			return gemaraDirective{}, fmt.Errorf("%q requires a value", key)
		}
		switch key {
		case "status":
			if directive.status != "" {
				return gemaraDirective{}, fmt.Errorf("duplicate \"status\" argument")
			}
			directive.status = value
		case "format":
			if directive.format != nil {
				return gemaraDirective{}, fmt.Errorf("duplicate \"format\" argument")
			}
			directive.format = &value
		case "default":
			if directive.defaultValue != nil {
				return gemaraDirective{}, fmt.Errorf("duplicate \"default\" argument")
			}
			directive.defaultValue = &value
		case "projectable":
			if directive.projectable != nil {
				return gemaraDirective{}, fmt.Errorf("duplicate \"projectable\" argument")
			}
			switch value {
			case "true":
				projectable := true
				directive.projectable = &projectable
			case "false":
				projectable := false
				directive.projectable = &projectable
			default:
				return gemaraDirective{}, fmt.Errorf("\"projectable\" must be true or false, got %q", value)
			}
		default:
			return gemaraDirective{}, fmt.Errorf("unknown argument %q", key)
		}
	}
	return directive, nil
}

// parseFileGemaraDirective adapts a file attribute into a CUE value so file
// directives receive the same balanced-token parser as value directives.
func parseFileGemaraDirective(attr *ast.Attribute) (gemaraDirective, error) {
	f := &ast.File{Decls: []ast.Decl{&ast.Field{
		Label: ast.NewIdent("directive"),
		Value: ast.NewIdent("_"),
		Attrs: []*ast.Attribute{attr},
	}}}
	v := cuecontext.New().BuildFile(f)
	if err := v.Err(); err != nil {
		return gemaraDirective{}, err
	}
	return parseGemaraDirective(v.LookupPath(cue.MakePath(cue.Str("directive"))).Attribute("gemara"))
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
