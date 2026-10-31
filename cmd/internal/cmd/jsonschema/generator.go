// SPDX-License-Identifier: Apache-2.0

package jsonschema

import (
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
		cmd := exec.Command(g.cueBin, "def", "-e", name, "--out", "jsonschema", ".")
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

	return map[string]any{
		"$schema": schemaDialect,
		"$defs":   defs,
	}, nil
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
	raw = strings.ReplaceAll(raw, "%23", "")
	raw = strings.ReplaceAll(raw, "%22", "")
	return raw
}

func normalizeDefName(raw string) string {
	name := strings.TrimPrefix(raw, "#")
	if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
		name = name[1 : len(name)-1]
	}
	return name
}
