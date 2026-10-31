// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"encoding/json"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/openapi"
	"github.com/getkin/kin-openapi/openapi3"
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

	info := &prepInfo{HelperRules: map[string][]helperRule{}}
	for _, f := range inst.Files {
		fileInfo, err := prepare(f)
		if err != nil {
			return cue.Value{}, nil, nil, fmt.Errorf("preparing %s: %w", f.Filename, err)
		}
		info.DroppedConditionals = append(info.DroppedConditionals, fileInfo.DroppedConditionals...)
		for helper, rules := range fileInfo.HelperRules {
			info.HelperRules[helper] = append(info.HelperRules[helper], rules...)
		}
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

// generateRaw runs the upstream encoder and decodes its output into a typed
// OpenAPI document. CUE CONVERTER WORKAROUND: The tuple normalization precedes
// decoding because the CUE encoder emits an OpenAPI-incompatible items array
// for open lists.
func generateRaw(v cue.Value, title, version string) (*openapi3.T, error) {
	f, err := openapi.Generate(v, &openapi.Config{
		Info: map[string]string{"title": title, "version": version},
		// The encoder is an OpenAPI 3.0 generator: "3.1.0" is accepted but
		// changes only how exclusiveMinimum/exclusiveMaximum are written, and
		// the rest of its output (e.g. `nullable`) stays 3.0. Ask for 3.0 so
		// the document is what it declares; Convert sets the patch version.
		Version: "3.0.0",
		// CUE CONVERTER WORKAROUND: Suppress two artefacts of CUE's open-list form ([#X, ...#X]): a
		// non-concrete `default` block that leaks raw CUE and @go()
		// attributes, and `additionalItems`, which OpenAPI 3.0 does not
		// define.
		FieldFilter: "Schema/(default|additionalItems)",
		NameFunc:    schemaName,
	})
	if err != nil {
		return nil, fmt.Errorf("generating OpenAPI from CUE: %w", err)
	}
	generated := v.Context().BuildFile(f)
	if err := generated.Err(); err != nil {
		return nil, fmt.Errorf("building generated OpenAPI: %w", err)
	}
	b, err := json.Marshal(generated)
	if err != nil {
		return nil, fmt.Errorf("encoding generated OpenAPI: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("decoding generated OpenAPI: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("generated OpenAPI decoded to nothing")
	}
	if err := unwrapTupleItems(raw); err != nil {
		return nil, err
	}
	b, err = json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("encoding normalized OpenAPI: %w", err)
	}
	var doc openapi3.T
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("decoding typed OpenAPI: %w", err)
	}
	return &doc, nil
}

// schemaName names definitions after themselves without the leading '#', and
// returns "" for everything else so field paths are expanded in place rather
// than minted as schemas.
//
// CUE CONVERTER WORKAROUND: Returning "" for the hidden #_… definitions is not an option: the encoder
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
