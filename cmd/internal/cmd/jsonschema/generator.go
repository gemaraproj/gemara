// SPDX-License-Identifier: Apache-2.0

package jsonschema

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

const schemaDialect = "https://json-schema.org/draft/2020-12/schema"

// Generator produces a canonical JSON Schema document from CUE definitions.
type Generator struct {
	schemaDir string
	cueBin    string
}

// NewGenerator creates a Generator targeting the specified CUE package directory.
func NewGenerator(schemaDir string) *Generator {
	return &Generator{
		schemaDir: schemaDir,
		cueBin:    "cue",
	}
}

// WithCueBin overrides the path to the cue binary.
func (g *Generator) WithCueBin(path string) *Generator {
	if path != "" {
		g.cueBin = path
	}
	return g
}

// DiscoverDefinitions returns all top-level exported definition names (#Name, excluding #_).
func (g *Generator) DiscoverDefinitions() ([]string, error) {
	absDir, err := filepath.Abs(g.schemaDir)
	if err != nil {
		return nil, fmt.Errorf("resolving schema directory: %w", err)
	}

	insts := load.Instances([]string{"."}, &load.Config{Dir: absDir})
	if len(insts) == 0 {
		return nil, fmt.Errorf("no CUE package found in %s", g.schemaDir)
	}
	inst := insts[0]
	if inst.Err != nil {
		return nil, fmt.Errorf("loading CUE package in %s: %w", g.schemaDir, inst.Err)
	}

	ctx := cuecontext.New()
	v := ctx.BuildInstance(inst)
	if err := v.Err(); err != nil {
		return nil, fmt.Errorf("building CUE instance in %s: %w", g.schemaDir, err)
	}

	it, err := v.Fields(cue.Definitions(true))
	if err != nil {
		return nil, fmt.Errorf("enumerating CUE definitions: %w", err)
	}

	namesMap := make(map[string]bool)
	for it.Next() {
		sel := it.Selector().String()
		if strings.HasPrefix(sel, "#") && !strings.HasPrefix(sel, "#_") {
			namesMap[sel] = true
		}
	}

	if len(namesMap) == 0 {
		return nil, fmt.Errorf("no definitions found in %s", g.schemaDir)
	}

	names := make([]string, 0, len(namesMap))
	for n := range namesMap {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}

// Generate builds the merged JSON Schema map containing all discovered definitions in $defs.
func (g *Generator) Generate() (map[string]any, error) {
	if _, err := exec.LookPath(g.cueBin); err != nil {
		return nil, fmt.Errorf("cue binary %q not found on PATH: %w", g.cueBin, err)
	}

	names, err := g.DiscoverDefinitions()
	if err != nil {
		return nil, err
	}

	absDir, err := filepath.Abs(g.schemaDir)
	if err != nil {
		return nil, fmt.Errorf("resolving schema directory: %w", err)
	}

	defs := make(map[string]any)

	for _, name := range names {
		cmd := exec.CommandContext(context.Background(), g.cueBin, "def", "-e", name, "--out", "jsonschema", ".")
		cmd.Dir = absDir

		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("cue def for %s failed in %s: %w (output: %s)", name, absDir, err, string(out))
		}

		clean := stripRefEncoding(string(out))
		var export map[string]any
		if err := json.Unmarshal([]byte(clean), &export); err != nil {
			return nil, fmt.Errorf("parsing JSON Schema exported for %s: %w", name, err)
		}

		if nested, ok := export["$defs"].(map[string]any); ok {
			for k, v := range nested {
				normK := normalizeDefName(k)
				defs[normK] = v
			}
		}

		body := make(map[string]any)
		for k, v := range export {
			if k != "$schema" && k != "$defs" {
				body[k] = v
			}
		}
		defs[normalizeDefName(name)] = body
	}

	// 1. Rewrite references to strict helpers to base definitions
	fixHelperRefs(defs)

	// 2. Drop hidden helper definitions (_*) from $defs
	for k := range defs {
		if strings.HasPrefix(k, "_") {
			delete(defs, k)
		}
	}

	// 3. Remove additionalProperties from base definitions that are extended
	removeBaseAdditionalProperties(defs)

	// 4. Rewrite additionalProperties: false to unevaluatedProperties: false on extending definitions
	fixUnevaluatedProperties(defs)

	// 5. Apply strict helper validation rules to base definitions
	if err := applyHelperRules(defs); err != nil {
		return nil, fmt.Errorf("applying helper rules: %w", err)
	}

	return map[string]any{
		"$schema": schemaDialect,
		"$defs":   defs,
	}, nil
}

// fixHelperRefs replaces any references to hidden strict helpers with their base definition.
func fixHelperRefs(v any) {
	switch node := v.(type) {
	case map[string]any:
		if ref, ok := node["$ref"].(string); ok {
			switch ref {
			case "#/$defs/_EvidenceStrict", "#/$defs/#_EvidenceStrict":
				node["$ref"] = "#/$defs/Evidence"
			case "#/$defs/_AssessmentLogStrict", "#/$defs/#_AssessmentLogStrict":
				node["$ref"] = "#/$defs/AssessmentLog"
			case "#/$defs/_MappingStrict", "#/$defs/#_MappingStrict":
				node["$ref"] = "#/$defs/Mapping"
			}
		}
		for _, child := range node {
			fixHelperRefs(child)
		}
	case []any:
		for _, item := range node {
			fixHelperRefs(item)
		}
	}
}

// removeBaseAdditionalProperties finds all definitions that are referenced as base schemas
// by other definitions (via $ref) and removes additionalProperties/unevaluatedProperties
// so that extending schemas can compose additional properties without being rejected.
func removeBaseAdditionalProperties(defs map[string]any) {
	bases := make(map[string]bool)
	var findBases func(v any)
	findBases = func(v any) {
		switch node := v.(type) {
		case map[string]any:
			if ref, ok := node["$ref"].(string); ok {
				base := strings.TrimPrefix(ref, "#/$defs/")
				bases[base] = true
			}
			for _, child := range node {
				findBases(child)
			}
		case []any:
			for _, item := range node {
				findBases(item)
			}
		}
	}
	for _, def := range defs {
		findBases(def)
	}

	for base := range bases {
		if baseDef, ok := defs[base].(map[string]any); ok {
			delete(baseDef, "additionalProperties")
			delete(baseDef, "unevaluatedProperties")
		}
	}
}

// fixUnevaluatedProperties rewrites additionalProperties: false to unevaluatedProperties: false
// where $ref is present (directly or in child elements), enabling Draft 2020-12 validators
// to recognize properties evaluated by referenced schemas while rejecting undeclared properties.
func fixUnevaluatedProperties(defs map[string]any) {
	for _, def := range defs {
		if defMap, ok := def.(map[string]any); ok {
			if _, ok := defMap["$ref"].(string); ok {
				if ap, ok := defMap["additionalProperties"].(bool); ok && !ap {
					delete(defMap, "additionalProperties")
					defMap["unevaluatedProperties"] = false
				}
				if props, ok := defMap["properties"].(map[string]any); ok {
					for _, p := range props {
						if pMap, ok := p.(map[string]any); ok {
							delete(pMap, "additionalProperties")
							delete(pMap, "unevaluatedProperties")
						}
					}
				}
			}
		}
	}

	var fixNested func(v any)
	fixNested = func(v any) {
		switch node := v.(type) {
		case map[string]any:
			if _, hasRef := node["$ref"].(string); hasRef {
				if ap, ok := node["additionalProperties"].(bool); ok && !ap {
					delete(node, "additionalProperties")
					node["unevaluatedProperties"] = false
				}
			}
			for _, child := range node {
				fixNested(child)
			}
		case []any:
			for _, item := range node {
				fixNested(item)
			}
		}
	}
	for _, def := range defs {
		fixNested(def)
	}
}

// applyHelperRules layers the strict validation rules from hidden helper definitions
// (#_AssessmentLogStrict, #_EvidenceStrict, #_MappingStrict) onto their respective
// base definitions (AssessmentLog, Evidence, Mapping). This ensures JSON Schema validation
// stays aligned with CUE and OpenAPI (ref: #537).
func applyHelperRules(defs map[string]any) error {
	appendAllOf := func(target map[string]any, rule map[string]any) {
		existing, ok := target["allOf"].([]any)
		if !ok {
			target["allOf"] = []any{rule}
			return
		}
		target["allOf"] = append(existing, rule)
	}

	// 1. #_AssessmentLogStrict: start required unless unexecuted
	// If result != "Not Run" && result != "Unknown" && result != "Not Applicable" => start is required.
	al, ok := defs["AssessmentLog"].(map[string]any)
	if !ok {
		return fmt.Errorf("required definition AssessmentLog missing from defs")
	}
	appendAllOf(al, map[string]any{
		"anyOf": []any{
			map[string]any{
				"properties": map[string]any{
					"result": map[string]any{
						"type": "string",
						"enum": []any{"Not Run", "Unknown", "Not Applicable"},
					},
				},
			},
			map[string]any{
				"required": []any{"start"},
			},
		},
	})

	// 2. #_EvidenceStrict: at least one of payload or source MUST be present
	ev, ok := defs["Evidence"].(map[string]any)
	if !ok {
		return fmt.Errorf("required definition Evidence missing from defs")
	}
	appendAllOf(ev, map[string]any{
		"anyOf": []any{
			map[string]any{
				"required": []any{"payload"},
			},
			map[string]any{
				"required": []any{"source"},
			},
		},
	})

	// 3. #_MappingStrict: targets required when relationship != "no-match"
	m, ok := defs["Mapping"].(map[string]any)
	if !ok {
		return fmt.Errorf("required definition Mapping missing from defs")
	}
	appendAllOf(m, map[string]any{
		"anyOf": []any{
			map[string]any{
				"properties": map[string]any{
					"relationship": map[string]any{
						"type": "string",
						"enum": []any{"no-match"},
					},
				},
			},
			map[string]any{
				"required": []any{"targets"},
			},
		},
	})

	return nil
}

// Write writes the canonical JSON Schema to outputPath.
func (g *Generator) Write(outputPath string) error {
	schema, err := g.Generate()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON Schema: %w", err)
	}
	b = append(b, '\n')

	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return os.WriteFile(outputPath, b, 0o644)
}

func stripRefEncoding(raw string) string {
	raw = strings.ReplaceAll(raw, "#/$defs/#", "#/$defs/")
	raw = strings.ReplaceAll(raw, "%23", "")
	raw = strings.ReplaceAll(raw, "%22", "")
	raw = strings.ReplaceAll(raw, "#/$defs/_EvidenceStrict", "#/$defs/Evidence")
	raw = strings.ReplaceAll(raw, "#/$defs/_AssessmentLogStrict", "#/$defs/AssessmentLog")
	raw = strings.ReplaceAll(raw, "#/$defs/_MappingStrict", "#/$defs/Mapping")
	return raw
}

func normalizeDefName(raw string) string {
	name := strings.TrimPrefix(raw, "#")
	if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
		name = name[1 : len(name)-1]
	}
	return name
}
