// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/openapi"
)

// loadPrepared loads the CUE package at schemaDir and runs the pre-pass over
// every file. It returns the built value, what the pre-pass rewrote, and the
// instance (whose Files and positions the annotation pass needs).
func loadPrepared(schemaDir string) (cue.Value, *prepInfo, *build.Instance, error) {
	insts := load.Instances([]string{"."}, &load.Config{Dir: schemaDir})
	if len(insts) == 0 {
		return cue.Value{}, nil, nil, fmt.Errorf("no CUE package found in %s", schemaDir)
	}
	inst := insts[0]
	if inst.Err != nil {
		return cue.Value{}, nil, nil, fmt.Errorf("loading CUE package %s: %w", schemaDir, inst.Err)
	}

	info := &prepInfo{TimeFormats: map[string]string{}}
	for _, f := range inst.Files {
		fileInfo, err := prepare(f)
		if err != nil {
			return cue.Value{}, nil, nil, fmt.Errorf("preparing %s: %w", f.Filename, err)
		}
		for name, layout := range fileInfo.TimeFormats {
			info.TimeFormats[name] = layout
		}
		info.DroppedConditionals = append(info.DroppedConditionals, fileInfo.DroppedConditionals...)
	}

	v := cuecontext.New().BuildInstance(inst)
	if err := v.Err(); err != nil {
		return cue.Value{}, nil, nil, fmt.Errorf("building CUE package %s: %w", schemaDir, err)
	}
	return v, info, inst, nil
}

// loadRaw loads the CUE package WITHOUT the pre-pass. The coverage gate needs a
// ground truth the pre-pass cannot have altered; comparing prepared CUE against
// a document generated from prepared CUE would police the rewrite with itself.
func loadRaw(schemaDir string) (cue.Value, error) {
	insts := load.Instances([]string{"."}, &load.Config{Dir: schemaDir})
	if len(insts) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE package found in %s", schemaDir)
	}
	inst := insts[0]
	if inst.Err != nil {
		return cue.Value{}, fmt.Errorf("loading CUE package %s: %w", schemaDir, inst.Err)
	}
	v := cuecontext.New().BuildInstance(inst)
	if err := v.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("building CUE package %s: %w", schemaDir, err)
	}
	return v, nil
}

// generateRaw runs the upstream encoder and decodes its output into a plain
// map for the post-pass to rewrite.
func generateRaw(v cue.Value, title, version string) (map[string]any, error) {
	b, err := openapi.Gen(v, &openapi.Config{
		Info: map[string]string{"title": title, "version": version},
		// The encoder accepts only "3.0.0" and "3.1.0"; the document's own
		// declared version is restored by the post-pass.
		Version: "3.0.0",
		// Suppress two artefacts of CUE's open-list form ([#X, ...#X]): a
		// non-concrete `default` block that leaks raw CUE and @go()
		// attributes, and `additionalItems`, which OpenAPI 3.0 does not
		// define.
		FieldFilter: "Schema/(default|additionalItems)",
		NameFunc:    schemaName,
	})
	if err != nil {
		return nil, fmt.Errorf("generating OpenAPI from CUE: %w", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("decoding generated OpenAPI: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("generated OpenAPI decoded to nothing")
	}
	normalizeNumbers(doc)
	return doc, nil
}

// normalizeNumbers replaces every integral float64 in the decoded document with
// an int64.
//
// encoding/json decodes every JSON number into a float64, and goccy/go-yaml
// renders float64(1) as `1.0`. OpenAPI 3.0 requires minItems, maxItems,
// minLength, maxLength, minProperties and maxProperties to be non-negative
// integers, and `1.0` is not one — a valid-looking document that is off-spec,
// which is precisely the failure mode this migration exists to eliminate.
// TestIntegerKeywordsAreIntegers gates the result.
func normalizeNumbers(n any) {
	switch x := n.(type) {
	case map[string]any:
		for k, v := range x {
			if f, ok := v.(float64); ok {
				if i, ok := integralInt64(f); ok {
					x[k] = i
					continue
				}
			}
			normalizeNumbers(v)
		}
	case []any:
		for i, v := range x {
			if f, ok := v.(float64); ok {
				if n, ok := integralInt64(f); ok {
					x[i] = n
					continue
				}
			}
			normalizeNumbers(v)
		}
	}
}

// integralInt64 reports whether f holds an integral value representable as an
// int64, and that value if so. A non-integral or out-of-range number is left as
// a float64 rather than rounded or truncated.
func integralInt64(f float64) (int64, bool) {
	if f != math.Trunc(f) || math.IsInf(f, 0) || math.IsNaN(f) {
		return 0, false
	}
	if f < math.MinInt64 || f >= math.MaxInt64 {
		return 0, false
	}
	return int64(f), true
}

// schemaName names definitions after themselves without the leading '#', and
// returns "" for everything else so field paths are expanded in place rather
// than minted as schemas.
//
// Returning "" for the hidden #_… definitions is not an option: the encoder
// then fails with "unsupported op . for object type". They are named here and
// collapsed by collapseHelperSchemas.
func schemaName(_ cue.Value, path cue.Path) string {
	sels := path.Selectors()
	if len(sels) == 0 {
		return ""
	}
	last := sels[len(sels)-1]
	if !last.IsDefinition() {
		return ""
	}
	return strings.TrimPrefix(last.String(), "#")
}
