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

func TestAssessmentIdentity(t *testing.T) {
	data, err := os.ReadFile("./test-data/good-evaluation-log-execution.yaml")
	if err != nil {
		t.Fatal(err)
	}
	def := schemaValue.LookupPath(cue.ParsePath("#EvaluationLog"))
	tests := []struct {
		name         string
		replacements []string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "cited plan with empty method",
			replacements: []string{"method-id: EM-AC-01-probe", `method-id: ""`},
			wantErr:      true,
			errContains:  `"execution-context"."method-id"`,
		},
		{
			name:         "fallback with empty method",
			replacements: []string{"method-id: EM-BASELINE-SCAN", `method-id: ""`},
			wantErr:      true,
			errContains:  `"execution-context"."method-id"`,
		},
		{
			name: "distinct plans and methods containing slashes",
			replacements: []string{
				"entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-probe",
				"entry-id: AP/AC-01\n    execution-context:\n      method-id: X",
				"entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-review",
				"entry-id: AP\n    execution-context:\n      method-id: AC-01/X",
			},
		},
		{
			name: "duplicate identities containing slashes",
			replacements: []string{
				"entry-id: AP-AC-01", "entry-id: AP/AC-01",
				"method-id: EM-AC-01-probe", "method-id: AC-01/X",
				"method-id: EM-AC-01-review", "method-id: AC-01/X",
			},
			wantErr:     true,
			errContains: "_uniqueAssessments",
		},
		{
			name: "literal hyphen plan differs from omitted plan",
			replacements: []string{
				"entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-probe",
				"entry-id: '-'\n    execution-context:\n      method-id: X",
				"    plan:\n      reference-id: security-policy\n      entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-review",
				"    execution-context:\n      method-id: X",
			},
		},
		{
			name: "literal hyphen method differs from omitted method",
			replacements: []string{
				"    plan:\n      reference-id: security-policy\n      entry-id: AP-AC-01\n", "",
				"method-id: EM-AC-01-probe", "method-id: '-'",
				"    execution-context:\n      method-id: EM-AC-01-review\n", "",
			},
		},
		{
			name:         "cited plan with an empty execution context",
			replacements: []string{"execution-context:\n      method-id: EM-AC-01-probe", "execution-context: {}"},
			wantErr:      true,
			errContains:  `"execution-context"."method-id"`,
		},
		{
			name: "two plan-less records with no method",
			replacements: []string{
				"    plan:\n      reference-id: security-policy\n      entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-probe\n", "",
				"    plan:\n      reference-id: security-policy\n      entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-review\n", "",
			},
			wantErr:     true,
			errContains: "_uniqueAssessments",
		},
		{
			name: "same plan id and method under two policies",
			replacements: []string{
				"reference-id: security-policy\n      entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-review",
				"reference-id: other-policy\n      entry-id: AP-AC-01\n    execution-context:\n      method-id: EM-AC-01-probe",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReplacer(tt.replacements...).Replace(string(data))
			validationErr := cueyaml.Validate([]byte(input), def)
			if tt.wantErr && validationErr == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tt.wantErr && validationErr != nil {
				t.Fatalf("unexpected validation error: %v", validationErr)
			}
			if validationErr != nil && !strings.Contains(validationErr.Error(), tt.errContains) {
				t.Errorf("error %q does not contain %q", validationErr.Error(), tt.errContains)
			}
		})
	}
}

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

		// AI agent capability catalog and ATR mappings (authored by ATR, validated against Gemara)
		{"valid AI agent capability catalog", "../examples/ai-agent/ai-agent-capability-catalog.yaml", "#CapabilityCatalog", false, ""},
		{"valid ATR categories to capabilities mapping", "../examples/ai-agent/atr-categories-to-capabilities-mapping.yaml", "#MappingDocument", false, ""},

		// RiskCatalog — positive
		{"valid risk catalog", "./test-data/good-risk-catalog.yaml", "#RiskCatalog", false, ""},

		// RiskCatalog — negative
		{"risk catalog with duplicate rank", "./test-data/bad-risk-catalog-duplicate-rank.yaml", "#RiskCatalog", true, ""},

		// Policy — positive
		{"valid policy", "./test-data/good-policy.yaml", "#Policy", false, ""},
		{"valid security policy", "./test-data/good-security-policy.yml", "#Policy", false, ""},
		{"policy with retention floor, ceiling and declared anchor", "./test-data/good-policy-retention.yaml", "#Policy", false, ""},
		{"policy executor with an execution environment", "./test-data/good-policy-executor-environment.yaml", "#Policy", false, ""},

		// Policy — negative (execution environment)
		{"executor environment digest without an algorithm", "./test-data/bad-policy-executor-environment-digest.yaml", "#Policy", true, `digests.0: invalid value "9f86d081`},
		{"observation vantage on the executor instead of the evidence", "./test-data/bad-policy-executor-environment-vantage.yaml", "#Policy", true, `"observation-vantage": field not allowed`},

		// Policy — negative (retention)
		{"retention obligation measured from an undeclared anchor", "./test-data/bad-policy-retention-undeclared-anchor.yaml", "#Policy", true, `does not satisfy list.Contains("employee-separation")`},
		{"retention duration below day precision", "./test-data/bad-policy-retention-subday-duration.yaml", "#Policy", true, `invalid value "PT1H"`},
		{"retention duration with no components", "./test-data/bad-policy-retention-empty-duration.yaml", "#Policy", true, `invalid value "P"`},
		{"retention obligation measured from collected-at with no duration", "./test-data/bad-policy-retention-collected-at-without-duration.yaml", "#Policy", true, "retention obligation RET-PII-MAX is measured from collected-at and must state a duration"},
		{"duplicate retention obligation id", "./test-data/bad-policy-retention-duplicate-obligation-id.yaml", "#Policy", true, `_uniqueRetentionObligationIds."RET-AUDIT-7Y"`},
		{"duplicate retention anchor id", "./test-data/bad-policy-retention-duplicate-anchor-id.yaml", "#Policy", true, `_uniqueRetentionAnchorIds."matter-closed"`},
		{"retention anchor redeclaring collected-at", "./test-data/bad-policy-retention-redeclared-collected-at.yaml", "#Policy", true, `_uniqueRetentionAnchorIds."collected-at"`},

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
		{"assessments that never ran omit start", "./test-data/good-evaluation-log-unstarted.yaml", "#EvaluationLog", false, ""},
		{"log author recording its execution environment, parameter inputs and per-evidence vantage", "./test-data/good-evaluation-log-author-environment-accepted.yaml", "#EvaluationLog", false, ""},
		{"assessments naming their executors and execution context", "./test-data/good-evaluation-log-execution.yaml", "#EvaluationLog", false, ""},

		// EvaluationLog — negative
		{"executed assessment missing start", "./test-data/bad-evaluation-log-missing-start.yaml", "#EvaluationLog", true, ""},
		{"evidence naming an undefined observation vantage", "./test-data/bad-evaluation-log-evidence-vantage.yaml", "#EvaluationLog", true, `"observation-vantage"`},
		{"parameter input that is not a string", "./test-data/bad-evaluation-log-parameter-input-type.yaml", "#EvaluationLog", true, `"parameter-inputs"."token-scope"`},
		{"execution environment on the evaluated target", "./test-data/bad-evaluation-log-target-environment.yaml", "#EvaluationLog", true, `target."execution-environment": field not allowed`},
		{"two assessments nothing tells apart", "./test-data/bad-evaluation-log-duplicate-assessment.yaml", "#EvaluationLog", true, "_uniqueAssessments"},
		{"a cited plan that never says which of its methods ran", "./test-data/bad-evaluation-log-plan-without-method.yaml", "#EvaluationLog", true, "\"execution-context\".\"method-id\""},

		// EnforcementLog — positive
		{"valid enforcement log", "./test-data/good-enforcement-log.yaml", "#EnforcementLog", false, ""},

		// EnforcementLog — negative
		{"enforcement action with invalid disposition", "./test-data/bad-enforcement-log.yaml", "#EnforcementLog", true, ""},
		{"enforcement action missing log reference", "./test-data/bad-enforcement-missing-log.yaml", "#EnforcementLog", true, ""},
		{"clear disposition with failed assessment", "./test-data/bad-enforcement-clear-failed.yaml", "#EnforcementLog", true, ""},

		// AuditLog — positive
		{"valid audit log", "./test-data/good-audit-log.yaml", "#AuditLog", false, ""},
		{"audit log evidence mapping with both coordinate and entry-id", "./test-data/good-audit-log-coordinate-and-entry-id.yaml", "#AuditLog", false, ""},
		{"audit log evidence with resolved conflicting retention obligations", "./test-data/good-audit-log-retention.yaml", "#AuditLog", false, ""},

		// AuditLog — negative
		{"audit log missing summary criteria and results", "./test-data/bad-audit-log.yaml", "#AuditLog", true, ""},
		{"audit log evidence source with invalid digest format", "./test-data/bad-audit-log-invalid-digest.yaml", "#AuditLog", true, ""},
		{"audit result referencing undeclared criteria", "./test-data/bad-audit-log-undeclared-criteria.yaml", "#AuditLog", true, ""},
		{"audit log evidence with neither payload nor source", "./test-data/bad-audit-log-evidence-neither.yaml", "#AuditLog", true, ""},
		{"audit log mapping reference url with no scheme", "./test-data/bad-audit-log-url-no-scheme.yaml", "#AuditLog", true, ""},
		{"audit log mapping reference url with a non-alphabetic scheme", "./test-data/bad-audit-log-url-invalid-scheme.yaml", "#AuditLog", true, ""},
		{"audit log target uri with no scheme", "./test-data/bad-audit-log-uri-no-scheme.yaml", "#AuditLog", true, ""},
		{"retention obligation reference without an effective instant", "./test-data/bad-audit-log-retention-missing-effective.yaml", "#AuditLog", true, "obligations.1.effective"},
		{"mutable preservation holds excluded from retention snapshots", "./test-data/bad-audit-log-retention-holds.yaml", "#AuditLog", true, "holds: field not allowed"},
		{"retention block with no obligations", "./test-data/bad-audit-log-retention-empty.yaml", "#AuditLog", true, "obligations"},

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
