// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

var defaultArtifacts = []string{
	"ControlCatalog", "GuidanceCatalog", "ThreatCatalog", "RiskCatalog",
	"CapabilityCatalog", "VectorCatalog", "PrincipleCatalog", "Policy",
	"MappingDocument", "Lexicon", "EvaluationLog", "EnforcementLog", "AuditLog",
}

// SchemaWalker performs schema-driven discovery of cross-document reference paths
// by walking the CUE AST for any struct containing a "reference-id" field.
type SchemaWalker struct {
	SchemaDir          string
	RefPathsByArtifact map[string][]string
}

// FindSchemaDir searches candidate locations for the Gemara CUE schema root.
func FindSchemaDir(startDir string) string {
	candidates := []string{}
	if startDir != "" {
		candidates = append(candidates, startDir, filepath.Join(startDir, ".."), filepath.Join(startDir, "../.."))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd, filepath.Join(wd, ".."), filepath.Join(wd, "../.."), filepath.Join(wd, "../../.."))
	}

	for _, cand := range candidates {
		abs, err := filepath.Abs(cand)
		if err != nil {
			continue
		}
		cueMod := filepath.Join(abs, "cue.mod")
		if fi, err := os.Stat(cueMod); err == nil && fi.IsDir() {
			return abs
		}
	}
	return ""
}

// NewSchemaWalker initializes a SchemaWalker by inspecting the CUE package at schemaDir.
func NewSchemaWalker(schemaDir string) (*SchemaWalker, error) {
	if schemaDir == "" {
		schemaDir = FindSchemaDir(".")
	}
	if schemaDir == "" {
		return nil, fmt.Errorf("no CUE schema directory found (cue.mod not found in search path)")
	}

	insts := load.Instances([]string{"."}, &load.Config{Dir: schemaDir})
	if len(insts) == 0 {
		return nil, fmt.Errorf("no CUE package found in %s", schemaDir)
	}
	inst := insts[0]
	if inst.Err != nil {
		return nil, fmt.Errorf("loading CUE package %s: %w", schemaDir, inst.Err)
	}

	v := cuecontext.New().BuildInstance(inst)
	if err := v.Err(); err != nil {
		return nil, fmt.Errorf("building CUE package %s: %w", schemaDir, err)
	}

	w := &SchemaWalker{
		SchemaDir:          schemaDir,
		RefPathsByArtifact: make(map[string][]string),
	}

	for _, art := range defaultArtifacts {
		found := make(map[string]bool)
		walkCUEValue(v.LookupPath(cue.ParsePath("#"+art)), art, 0, found)

		var paths []string
		prefix := art + "."
		for p := range found {
			relPath := strings.TrimPrefix(p, prefix)
			if relPath != p && relPath != "" {
				paths = append(paths, relPath)
			}
		}
		sort.Strings(paths)
		w.RefPathsByArtifact[art] = paths
	}

	return w, nil
}

func walkCUEValue(x cue.Value, path string, depth int, found map[string]bool) {
	if depth > 14 {
		return
	}

	// Lists: walk the element type.
	if elem := x.LookupPath(cue.MakePath(cue.AnyIndex)); elem.Exists() {
		if elem.IncompleteKind() == cue.BottomKind {
			// The element doesn't evaluate on its own (conditional rules on
			// the list), so walk each conjunct of the list instead.
			if op, args := x.Expr(); op == cue.AndOp {
				for _, a := range args {
					walkCUEValue(a, path, depth+1, found)
				}
			}
			return
		}
		walkCUEValue(elem, path+"[]", depth+1, found)
		return
	}

	if x.IncompleteKind()&cue.StructKind == 0 {
		return
	}

	if x.LookupPath(cue.MakePath(cue.Str("reference-id"))).Exists() {
		found[path] = true
	}

	it, _ := x.Fields(cue.Optional(true))
	for it.Next() {
		sel := it.Selector()
		if sel.IsDefinition() || strings.HasPrefix(sel.String(), "_") {
			continue
		}
		name := strings.Trim(sel.String(), `"?!`)
		if name == "reference-id" || name == "entries" {
			continue
		}
		walkCUEValue(it.Value(), path+"."+name, depth+1, found)
	}
}

// ExtractReferences dynamically walks root map along the paths discovered for doc.Type
// and extracts all reference-id and entry-id citations into doc.ReferenceUsages and doc.EntityUsages.
func (w *SchemaWalker) ExtractReferences(doc *Document, root map[string]any) {
	paths, ok := w.RefPathsByArtifact[doc.Type]
	if !ok {
		return
	}

	for _, p := range paths {
		tokens := strings.Split(p, ".")
		w.evaluatePath(doc, root, tokens, "")
	}
}

func (w *SchemaWalker) evaluatePath(doc *Document, current any, tokens []string, prefix string) {
	if len(tokens) == 0 {
		w.extractNodeReferences(doc, current, prefix)
		return
	}

	tok := tokens[0]
	isList := strings.HasSuffix(tok, "[]")
	fieldName := strings.TrimSuffix(tok, "[]")

	currMap, ok := current.(map[string]any)
	if !ok {
		return
	}

	val, exists := currMap[fieldName]
	if !exists || val == nil {
		return
	}

	nextPrefix := fieldName
	if prefix != "" {
		nextPrefix = prefix + "." + fieldName
	}

	if isList {
		switch list := val.(type) {
		case []any:
			for i, item := range list {
				itemPrefix := fmt.Sprintf("%s[%d]", nextPrefix, i)
				w.evaluatePath(doc, item, tokens[1:], itemPrefix)
			}
		case map[string]any:
			w.evaluatePath(doc, list, tokens[1:], nextPrefix)
		}
	} else {
		w.evaluatePath(doc, val, tokens[1:], nextPrefix)
	}
}

func (w *SchemaWalker) extractNodeReferences(doc *Document, node any, context string) {
	if node == nil {
		return
	}

	switch v := node.(type) {
	case []any:
		for i, item := range v {
			w.extractNodeReferences(doc, item, fmt.Sprintf("%s[%d]", context, i))
		}
		return
	case map[string]any:
		rID, _ := v["reference-id"].(string)
		if rID != "" {
			doc.AddReferenceUsage(rID, context)

			// Direct entry-id or risk-id (#EntryMapping or similar)
			if entryID, ok := v["entry-id"].(string); ok && entryID != "" {
				doc.AddEntityUsage(rID, entryID, "", context)
			} else if riskID, ok := v["risk-id"].(string); ok && riskID != "" {
				doc.AddEntityUsage(rID, riskID, "", context)
			}

			// #MultiEntryMapping entries
			if entries, ok := v["entries"].([]any); ok {
				for _, e := range entries {
					if eMap, ok := e.(map[string]any); ok {
						entryID, _ := eMap["reference-id"].(string)
						if entryID == "" {
							entryID, _ = eMap["entry-id"].(string)
						}
						if entryID != "" {
							doc.AddEntityUsage(rID, entryID, "", context)
						}
					} else if eStr, ok := e.(string); ok && eStr != "" {
						doc.AddEntityUsage(rID, eStr, "", context)
					}
				}
			}
		}
	}
}
