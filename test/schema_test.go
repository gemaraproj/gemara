// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	cuejson "cuelang.org/go/encoding/json"
	cueyaml "cuelang.org/go/encoding/yaml"
)

var schemaValue cue.Value

func TestMain(m *testing.M) {
	ctx := cuecontext.New()

	schemaDir, err := filepath.Abs("..")
	if err != nil {
		panic("failed to resolve schema directory: " + err.Error())
	}

	cfg := &load.Config{
		Dir: schemaDir,
	}
	instances := load.Instances([]string{"."}, cfg)
	if len(instances) != 1 {
		panic("expected exactly one CUE instance")
	}

	schemaValue = ctx.BuildInstance(instances[0])
	if schemaValue.Err() != nil {
		panic("failed to build CUE schema: " + schemaValue.Err().Error())
	}

	os.Exit(m.Run())
}

func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		definition  string
		wantErr     bool
		errContains string
	}{
		// ControlCatalog — positive
		{"valid control catalog YAML", "./test-data/good-ccc.yaml", "#ControlCatalog", false, ""},
		{"valid control catalog JSON", "./test-data/good-ccc.json", "#ControlCatalog", false, ""},
		{"valid OSPS baseline", "./test-data/good-osps.yml", "#ControlCatalog", false, ""},
		{"valid lifecycle catalog", "./test-data/good-lifecycle.yaml", "#ControlCatalog", false, ""},

		// GuidanceCatalog — positive
		{"valid AI governance framework", "./test-data/good-aigf.yaml", "#GuidanceCatalog", false, ""},

		// Policy — positive
		{"valid policy", "./test-data/good-policy.yaml", "#Policy", false, ""},
		{"valid security policy", "./test-data/good-security-policy.yml", "#Policy", false, ""},

		// ControlCatalog — negative
		{"invalid YAML against ControlCatalog", "./test-data/bad.yaml", "#ControlCatalog", true, ""},
		{"invalid JSON against ControlCatalog", "./test-data/bad.json", "#ControlCatalog", true, ""},

		// GuidanceCatalog — negative
		{"retired guideline with recommendations", "./test-data/bad-lifecycle.yaml", "#GuidanceCatalog", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("read %s: %v", tt.file, err)
			}

			def := schemaValue.LookupPath(cue.ParsePath(tt.definition))
			if def.Err() != nil {
				t.Fatalf("lookup %s: %v", tt.definition, def.Err())
			}

			var validationErr error
			switch {
			case strings.HasSuffix(tt.file, ".json"):
				validationErr = cuejson.Validate(data, def)
			case strings.HasSuffix(tt.file, ".yaml"), strings.HasSuffix(tt.file, ".yml"):
				validationErr = cueyaml.Validate(data, def)
			default:
				t.Fatalf("unsupported file extension: %s", tt.file)
			}

			if tt.wantErr && validationErr == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && validationErr != nil {
				t.Errorf("unexpected validation error: %v", validationErr)
			}
			if tt.errContains != "" && validationErr != nil {
				if !strings.Contains(validationErr.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", validationErr.Error(), tt.errContains)
				}
			}
		})
	}
}
