// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

// Change is an ERR-level backward-incompatible change reported by oasdiff.
// Schema is the Gemara schema (component) the change was found on.
type Change struct {
	ID     string
	Schema string
	Text   string
}

// strictSchemaSeverityLevels promotes schema-level warnings that can change a
// stable artifact contract. Path, parameter, header, and webhook warnings are
// not relevant to the synthetic component-schema operations.
func strictSchemaSeverityLevels() map[string]checker.Level {
	return map[string]checker.Level{
		checker.RequestBodyAllOfRemovedId:         checker.ERR,
		checker.RequestPropertyAllOfRemovedId:     checker.ERR,
		checker.RequestPropertyPatternChangedId:   checker.ERR,
		checker.RequestPropertyRemovedId:          checker.ERR,
		checker.ResponseOptionalPropertyRemovedId: checker.ERR,
		checker.ResponseRequiredPropertyRemovedId: checker.ERR,
	}
}

// filterAllowed drops changes matched by the allowlist. An entry keyed
// "<id> <schema>" drops only that change on that schema. It returns the
// remaining unallowed changes along with the set of allowlist entries that
// did not match any breaking change in changes.
func filterAllowed(changes []Change, allowed map[string]bool) ([]Change, []string) {
	var out []Change
	matched := make(map[string]bool)
	for _, c := range changes {
		key := c.ID + " " + c.Schema
		if allowed[key] {
			matched[key] = true
			continue
		}
		out = append(out, c)
	}
	var unused []string
	for k := range allowed {
		if !matched[k] {
			unused = append(unused, k)
		}
	}
	return out, unused
}

// loadAllowlist reads an allowlist file of oasdiff exceptions, one per line.
// Each entry is a schema-scoped "<id> <schema>" pair (see filterAllowed). Blank
// lines and lines beginning with '#' are ignored. An empty path yields an empty
// (non-nil) map and a nil error.
func loadAllowlist(path string) (allowed map[string]bool, err error) {
	allowed = make(map[string]bool)
	if path == "" {
		return allowed, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening allowlist %q: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing allowlist %q: %w", path, cerr)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		allowed[line] = true
	}
	if serr := scanner.Err(); serr != nil {
		return nil, fmt.Errorf("reading allowlist %q: %w", path, serr)
	}
	return allowed, err
}

// loadWrapped loads an OpenAPI document and synthesizes bidirectional paths for
// every component schema so oasdiff (which diffs the API surface, not bare
// component schemas) can classify schema-level changes. Each schema is exposed
// as a GET response body (consumer direction: removed required property =
// breaking) and a POST request body (producer direction: new required property =
// breaking).
func loadWrapped(path string) (*openapi3.T, error) {
	doc, err := openapi3.NewLoader().LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	if doc.Paths == nil {
		doc.Paths = openapi3.NewPaths()
	}
	if doc.Components == nil {
		doc.Components = &openapi3.Components{}
	}
	for name := range doc.Components.Schemas {
		ref := doc.Components.Schemas[name] // already-resolved SchemaRef
		mt := openapi3.NewContentWithJSONSchemaRef(ref)
		emptyOK := openapi3.NewResponses(openapi3.WithStatus(200,
			&openapi3.ResponseRef{Value: openapi3.NewResponse().WithDescription("ok")}))
		get := &openapi3.Operation{
			OperationID: "get_" + name,
			Responses: openapi3.NewResponses(openapi3.WithStatus(200,
				&openapi3.ResponseRef{Value: openapi3.NewResponse().WithDescription("ok").WithContent(mt)})),
		}
		post := &openapi3.Operation{
			OperationID: "post_" + name,
			RequestBody: &openapi3.RequestBodyRef{
				Value: openapi3.NewRequestBody().WithRequired(true).WithContent(mt),
			},
			Responses: emptyOK,
		}
		// A schema the converter marked experimental (x-status) is exempted from
		// the gate the way the old compat test exempted non-stable schemas: map
		// it to oasdiff's x-stability-level "alpha", which the default (beta)
		// threshold filters out before the breaking-change checks run.
		if schemaStatus(ref) == "experimental" {
			get.Extensions = map[string]any{diff.XStabilityLevelExtension: "alpha"}
			post.Extensions = map[string]any{diff.XStabilityLevelExtension: "alpha"}
		}
		doc.Paths.Set("/_schema/"+name, &openapi3.PathItem{Get: get, Post: post})
	}
	return doc, nil
}

// schemaStatus returns the value of the converter-emitted x-status extension on
// a schema, or "" if absent. Extension values arrive either as a decoded string
// or as raw JSON depending on the loader path, so both are handled.
func schemaStatus(ref *openapi3.SchemaRef) string {
	if ref == nil || ref.Value == nil || ref.Value.Extensions == nil {
		return ""
	}
	switch v := ref.Value.Extensions["x-status"].(type) {
	case string:
		return v
	case json.RawMessage:
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			return s
		}
	}
	return ""
}

// breakingChanges returns the ERR-level backward-incompatible changes from base
// to rev. Both documents must already be wrapped via loadWrapped.
func breakingChanges(base, rev *openapi3.T) ([]Change, error) {
	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(),
		&load.SpecInfo{Url: "base", Spec: base},
		&load.SpecInfo{Url: "rev", Spec: rev})
	if err != nil {
		return nil, err
	}
	loc := checker.NewDefaultLocalizer()
	var out []Change
	config := checker.NewConfig(checker.GetAllChecks(), checker.WithSeverityLevels(strictSchemaSeverityLevels()))
	for _, c := range checker.CheckBackwardCompatibility(config, d, sources) {
		if c.GetLevel() == checker.ERR {
			// Paths are the synthesized "/_schema/<name>" wrappers; expose just
			// the schema name so output and allowlist entries stay user-facing.
			schema := strings.TrimPrefix(c.GetPath(), "/_schema/")
			out = append(out, Change{ID: c.GetId(), Schema: schema, Text: c.GetUncolorizedText(loc)})
		}
	}
	return out, nil
}
