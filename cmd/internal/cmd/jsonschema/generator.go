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
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/literal"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/token"
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

	// 3. Narrow unevaluatedProperties on extended base definitions and schemas
	narrowExtendedProperties(defs)

	// 4. Relax AssessmentLog.requirement to only require entry-id, as CUE derives reference-id
	// from the enclosing control (evaluationlog.cue:27-29).
	relaxAssessmentLogRequirement(defs)

	// 5. Apply strict helper validation rules to base definitions
	if err := g.applyHelperRules(defs, absDir); err != nil {
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

// convertAdditionalToUnevaluated replaces additionalProperties: false with unevaluatedProperties: false.
func convertAdditionalToUnevaluated(schema map[string]any) {
	if ap, ok := schema["additionalProperties"].(bool); ok && !ap {
		delete(schema, "additionalProperties")
		schema["unevaluatedProperties"] = false
	}
}

// narrowExtendedProperties finds base definitions that are extended by other schemas
// (via $ref sitting next to sibling properties or inside an allOf) and converts their
// additionalProperties: false to unevaluatedProperties: false, allowing extending schemas
// to compose additional properties while rejecting undeclared nested properties.
func narrowExtendedProperties(defs map[string]any) {
	extendedBases := make(map[string]bool)

	var findExtensions func(v any)
	findExtensions = func(v any) {
		switch node := v.(type) {
		case map[string]any:
			// Check if $ref sits next to sibling properties
			if ref, ok := node["$ref"].(string); ok {
				if props, ok := node["properties"].(map[string]any); ok && len(props) > 0 {
					base := strings.TrimPrefix(ref, "#/$defs/")
					extendedBases[base] = true
					convertAdditionalToUnevaluated(node)
					for _, p := range props {
						if pMap, ok := p.(map[string]any); ok {
							delete(pMap, "additionalProperties")
							delete(pMap, "unevaluatedProperties")
						}
					}
				}
			}
			// Check if $ref sits inside an allOf
			if allOf, ok := node["allOf"].([]any); ok {
				for _, item := range allOf {
					if itemMap, ok := item.(map[string]any); ok {
						if ref, ok := itemMap["$ref"].(string); ok {
							base := strings.TrimPrefix(ref, "#/$defs/")
							extendedBases[base] = true
						}
						convertAdditionalToUnevaluated(itemMap)
					}
				}
				convertAdditionalToUnevaluated(node)
			}
			for _, child := range node {
				findExtensions(child)
			}
		case []any:
			for _, item := range node {
				findExtensions(item)
			}
		}
	}

	for _, def := range defs {
		findExtensions(def)
	}

	for base := range extendedBases {
		if baseDef, ok := defs[base].(map[string]any); ok {
			delete(baseDef, "additionalProperties")
			delete(baseDef, "unevaluatedProperties")
		}
	}
}

// relaxAssessmentLogRequirement allows AssessmentLog.requirement to omit reference-id,
// since CUE derives requirement."reference-id" from the parent control (evaluationlog.cue:27-29).
// This allows serialized logs that omit the redundant reference-id to validate against JSON Schema.
func relaxAssessmentLogRequirement(defs map[string]any) {
	if al, ok := defs["AssessmentLog"].(map[string]any); ok {
		if props, ok := al["properties"].(map[string]any); ok {
			props["requirement"] = map[string]any{
				"type":        "object",
				"description": "Requirement should map to the assessment requirement for this assessment.",
				"additionalProperties": false,
				"properties": map[string]any{
					"entry-id": map[string]any{
						"type":        "string",
						"description": "entry-id is the identifier being mapped to in the referenced artifact",
					},
					"reference-id": map[string]any{
						"type":        "string",
						"description": "reference-id is the id for a MappingReference entry in the artifact's metadata",
					},
					"remarks": map[string]any{
						"type":        "string",
						"description": "remarks is prose describing the mapping relationship",
					},
				},
				"required": []any{"entry-id"},
			}
		}
	}

	// Also relax the cross-field requirement on ControlEvaluation.assessment-logs
	if ce, ok := defs["ControlEvaluation"].(map[string]any); ok {
		if props, ok := ce["properties"].(map[string]any); ok {
			if alProp, ok := props["assessment-logs"].(map[string]any); ok {
				if allOf, ok := alProp["allOf"].([]any); ok {
					var filtered []any
					for _, item := range allOf {
						if itemMap, ok := item.(map[string]any); ok {
							if desc, ok := itemMap["description"].(string); ok && strings.Contains(desc, "control reference and the assessments' references match") {
								continue
							}
						}
						filtered = append(filtered, item)
					}
					alProp["allOf"] = filtered
				}
			}
		}
	}
}

func excludedValues(e ast.Expr) (string, []string, bool) {
	b, ok := e.(*ast.BinaryExpr)
	if !ok {
		return "", nil, false
	}
	if b.Op == token.LAND {
		left, lv, lok := excludedValues(b.X)
		right, rv, rok := excludedValues(b.Y)
		if !lok || !rok || left != right {
			return "", nil, false
		}
		return left, append(lv, rv...), true
	}
	if b.Op != token.NEQ {
		return "", nil, false
	}
	id, isIdent := b.X.(*ast.Ident)
	lit, isLit := b.Y.(*ast.BasicLit)
	if !isIdent || !isLit || lit.Kind != token.STRING {
		return "", nil, false
	}
	value, err := literal.Unquote(lit.Value)
	if err != nil {
		return "", nil, false
	}
	return id.Name, []string{value}, true
}

// extractHelperRules parses CUE source files in schemaDir to dynamically extract
// the condition values from hidden helper definitions (#_AssessmentLogStrict, #_MappingStrict).
func (g *Generator) extractHelperRules(absDir string) (unexecuted []string, noMatch []string) {
	insts := load.Instances([]string{"."}, &load.Config{Dir: absDir})
	if len(insts) == 0 || insts[0].Err != nil {
		return []string{"Not Run", "Unknown", "Not Applicable"}, []string{"no-match"}
	}

	for _, f := range insts[0].Files {
		for _, decl := range f.Decls {
			field, ok := decl.(*ast.Field)
			if !ok {
				continue
			}
			labelName, _, err := ast.LabelName(field.Label)
			if err != nil {
				continue
			}

			var walkStruct func(ast.Node)
			walkStruct = func(n ast.Node) {
				switch x := n.(type) {
				case *ast.BinaryExpr:
					if x.Op == token.AND {
						walkStruct(x.X)
						walkStruct(x.Y)
					}
				case *ast.StructLit:
					for _, elt := range x.Elts {
						if comp, ok := elt.(*ast.Comprehension); ok {
							for _, clause := range comp.Clauses {
								if ifClause, ok := clause.(*ast.IfClause); ok {
									if subject, vals, ok := excludedValues(ifClause.Condition); ok {
										if labelName == "#_AssessmentLogStrict" && subject == "result" {
											unexecuted = vals
										} else if labelName == "#_MappingStrict" && subject == "relationship" {
											noMatch = vals
										}
									}
								}
							}
						}
					}
				}
			}
			walkStruct(field.Value)
		}
	}

	if len(unexecuted) == 0 {
		unexecuted = []string{"Not Run", "Unknown", "Not Applicable"}
	}
	if len(noMatch) == 0 {
		noMatch = []string{"no-match"}
	}
	return unexecuted, noMatch
}

// applyHelperRules layers the strict validation rules from hidden helper definitions
// (#_AssessmentLogStrict, #_EvidenceStrict, #_MappingStrict) onto their respective
// base definitions (AssessmentLog, Evidence, Mapping). This ensures JSON Schema validation
// stays aligned with CUE and OpenAPI (ref: #537).
func (g *Generator) applyHelperRules(defs map[string]any, absDir string) error {
	appendAllOf := func(target map[string]any, rule map[string]any) {
		existing, ok := target["allOf"].([]any)
		if !ok {
			target["allOf"] = []any{rule}
			return
		}
		target["allOf"] = append(existing, rule)
	}

	unexecutedResults, noMatchRelationship := g.extractHelperRules(absDir)
	unexecutedEnum := make([]any, len(unexecutedResults))
	for i, v := range unexecutedResults {
		unexecutedEnum[i] = v
	}
	noMatchEnum := make([]any, len(noMatchRelationship))
	for i, v := range noMatchRelationship {
		noMatchEnum[i] = v
	}

	// 1. #_AssessmentLogStrict: start required unless unexecuted
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
						"enum": unexecutedEnum,
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
						"enum": noMatchEnum,
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
