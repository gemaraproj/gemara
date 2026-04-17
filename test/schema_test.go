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
var schemaCtx *cue.Context

func TestMain(m *testing.M) {
	schemaCtx = cuecontext.New()
	ctx := schemaCtx

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
		{"valid nested control catalog", "./test-data/nested-good-ccc.yaml", "#ControlCatalog", false, ""},

		// GuidanceCatalog — positive
		{"valid AI governance framework", "./test-data/good-aigf.yaml", "#GuidanceCatalog", false, ""},
		// PrinciplesCatalog — positive
		{"valid AIGF principles catalog", "./test-data/good-aigf-principles.yaml", "#PrincipleCatalog", false, ""},

		// VectorCatalog — positive
		{"valid AIGF vector catalog", "./test-data/good-aigf-vectors.yaml", "#VectorCatalog", false, ""},
		{"threats with vectors", "./test-data/good-threat-catalog.yaml", "#ThreatCatalog", false, ""},
		{"valid capability catalog", "./test-data/good-capability-catalog.yaml", "#CapabilityCatalog", false, ""},
		{"vector mapping", "./test-data/good-vector-owasp-mapping.yaml", "#MappingDocument", false, ""},

		// RiskCatalog — positive
		{"valid risk catalog", "./test-data/good-risk-catalog.yaml", "#RiskCatalog", false, ""},

		// Policy — positive
		{"valid policy", "./test-data/good-policy.yaml", "#Policy", false, ""},
		{"valid security policy", "./test-data/good-security-policy.yml", "#Policy", false, ""},

		// ControlCatalog — negative
		{"invalid YAML", "./test-data/bad.yaml", "#ControlCatalog", true, ""},
		{"invalid JSON", "./test-data/bad.json", "#ControlCatalog", true, ""},
		{"controls without groups", "./test-data/bad-no-groups.yaml", "#ControlCatalog", true, ""},

		// MappingDocument — positive
		{"valid mapping document", "./test-data/good-mapping-document.yaml", "#MappingDocument", false, ""},
		{"valid AIGF NIST 800-53 mapping", "./test-data/good-aigf-nist-mapping.yaml", "#MappingDocument", false, ""},

		// MappingDocument — negative
		{"invalid mapping document without mapping-references", "./test-data/bad-mapping-document.yaml", "#MappingDocument", true, ""},
		{"mapping missing targets for non-no-match relationship", "./test-data/bad-mapping-no-target.yaml", "#MappingDocument", true, ""},

		// Lexicon — positive
		{"valid lexicon", "./test-data/good-lexicon.yaml", "#Lexicon", false, ""},

		// Lexicon — negative
		{"lexicon with duplicate term ids", "./test-data/bad-lexicon-duplicate-term-id.yaml", "#Lexicon", true, ""},

		// GuidanceCatalog — negative
		{"retired guideline with recommendations", "./test-data/bad-lifecycle.yaml", "#GuidanceCatalog", true, ""},

		// EvaluationLog — positive
		{"valid PVTR baseline scan", "./test-data/pvtr-baseline-scan.yaml", "#EvaluationLog", false, ""},

		// EnforcementLog — positive
		{"valid enforcement log", "./test-data/good-enforcement-log.yaml", "#EnforcementLog", false, ""},

		// EnforcementLog — negative
		{"enforcement action with invalid disposition", "./test-data/bad-enforcement-log.yaml", "#EnforcementLog", true, ""},
		{"enforcement action missing log reference", "./test-data/bad-enforcement-missing-log.yaml", "#EnforcementLog", true, ""},
		{"clear disposition with failed assessment", "./test-data/bad-enforcement-clear-failed.yaml", "#EnforcementLog", true, ""},

		// AuditLog — positive
		{"valid audit log", "./test-data/good-audit-log.yaml", "#AuditLog", false, ""},

		// AuditLog — negative
		{"audit log missing summary criteria and results", "./test-data/bad-audit-log.yaml", "#AuditLog", true, ""},

		// CapabilityCatalog — negative
		{"capability with invalid group", "./test-data/bad-capability-invalid-group.yaml", "#CapabilityCatalog", true, ""},

		// ThreatCatalog — negative
		{"threat with invalid group", "./test-data/bad-threat-invalid-group.yaml", "#ThreatCatalog", true, ""},

		// PrincipleCatalog — negative
		{"principle with invalid group", "./test-data/bad-principle-invalid-group.yaml", "#PrincipleCatalog", true, ""},

		// ControlCatalog — negative (group validation)
		{"control with invalid group", "./test-data/bad-control-invalid-group.yaml", "#ControlCatalog", true, ""},

		// ControlCatalog — edge cases
		{"empty nested catalog", "./test-data/nested-empty.yaml", "#ControlCatalog", false, ""},
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
