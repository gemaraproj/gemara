// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
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

		// Evidence rules ride on #Evidence and #EvidenceMapping, which both logs use.
		// Their fixtures are evaluation logs so that reshaping the audit — which is
		// happening under #496 — cannot quietly stop them testing evidence. Each
		// negative fixture stands for one rule.
		{"evidence that is addressed, carried, and reused by two assessments", "./test-data/good-evaluation-log-evidence.yaml", "#EvaluationLog", false, ""},
		{"sha256 digest of the wrong length", "./test-data/bad-evaluation-log-digest-sha256-length.yaml", "#EvaluationLog", true, "source.digest"},
		{"download-url with no URI scheme", "./test-data/bad-evaluation-log-download-url-no-scheme.yaml", "#EvaluationLog", true, "\"download-url\""},
		{"media-type with no subtype separator", "./test-data/bad-evaluation-log-media-type-malformed.yaml", "#EvaluationLog", true, "\"media-type\""},
		{"inline payload with a source digest", "./test-data/bad-evaluation-log-evidence-payload-with-digest.yaml", "#EvaluationLog", true, "inline payload cannot also have a source digest"},
		{"one evidence id naming different items in two assessments", "./test-data/bad-evaluation-log-evidence-id-reused.yaml", "#EvaluationLog", true, "names two different evidence items"},
		{"one evidence item listed twice in one evidence list", "./test-data/bad-evaluation-log-evidence-id-repeated-in-list.yaml", "#EvaluationLog", true, "appears more than once in one evidence list"},
		{"evidence source that is not declared", "./test-data/bad-evaluation-log-evidence-source-undeclared.yaml", "#EvaluationLog", true, "not declared in metadata.mapping-references"},
		{"evidence with no id reported where the id is missing", "./test-data/bad-evaluation-log-evidence-missing-id.yaml", "#EvaluationLog", true, "evidence.0.id: incomplete value string"},
		{"entry-id on evidence that is not a Gemara artifact", "./test-data/bad-evaluation-log-evidence-entry-id-not-artifact.yaml", "#EvaluationLog", true, "entry-id is only for evidence that is a Gemara artifact"},

		// Fragments are validated against one definition, so a case that needs only
		// a digest or one evidence item does not carry a whole log. They live in
		// their own folder because they are not Gemara documents.
		{"sha512 digest", "./test-data/fragments/good-digest-sha512.yaml", "#Digest", false, ""},
		{"blake3 digest", "./test-data/fragments/good-digest-blake3.yaml", "#Digest", false, ""},
		{"digest with an unknown algorithm", "./test-data/fragments/good-digest-unknown-algorithm.yaml", "#Digest", false, ""},
		{"digest with a misspelt algorithm passes as unknown", "./test-data/fragments/good-digest-misspelt-algorithm.yaml", "#Digest", false, ""},
		{"digest with no algorithm separator", "./test-data/fragments/bad-digest-no-separator.yaml", "#Digest", true, ""},
		{"sha256 digest in uppercase hex", "./test-data/fragments/bad-digest-sha256-uppercase-hex.yaml", "#Digest", true, ""},
		{"sha512 digest with a sha256 length", "./test-data/fragments/bad-digest-sha512-length.yaml", "#Digest", true, ""},
		{"blake3 digest one character short", "./test-data/fragments/bad-digest-blake3-length.yaml", "#Digest", true, ""},
		{"inline payload with a source size", "./test-data/fragments/bad-evidence-payload-with-size.yaml", "#_EvidenceStrict", true, "inline payload cannot also have a source size"},
		{"inline payload with a source download-url", "./test-data/fragments/bad-evidence-payload-with-download-url.yaml", "#_EvidenceStrict", true, "inline payload cannot also have a source download-url"},

		// EvaluationLog — negative
		{"executed assessment missing start", "./test-data/bad-evaluation-log-missing-start.yaml", "#EvaluationLog", true, ""},

		// EnforcementLog — positive
		{"valid enforcement log", "./test-data/good-enforcement-log.yaml", "#EnforcementLog", false, ""},

		// EnforcementLog — negative
		{"enforcement action with invalid disposition", "./test-data/bad-enforcement-log.yaml", "#EnforcementLog", true, ""},
		{"enforcement action missing log reference", "./test-data/bad-enforcement-missing-log.yaml", "#EnforcementLog", true, ""},
		{"clear disposition with failed assessment", "./test-data/bad-enforcement-clear-failed.yaml", "#EnforcementLog", true, ""},

		// AuditLog — positive
		{"valid audit log", "./test-data/good-audit-log.yaml", "#AuditLog", false, ""},
		{"audit log evidence mapping with both coordinate and entry-id", "./test-data/good-audit-log-coordinate-and-entry-id.yaml", "#AuditLog", false, ""},

		// AuditLog — negative
		{"audit log missing summary criteria and results", "./test-data/bad-audit-log.yaml", "#AuditLog", true, ""},
		{"audit log evidence source with invalid digest format", "./test-data/bad-audit-log-invalid-digest.yaml", "#AuditLog", true, ""},
		{"audit result referencing undeclared criteria", "./test-data/bad-audit-log-undeclared-criteria.yaml", "#AuditLog", true, ""},
		{"audit log evidence with neither payload nor source", "./test-data/bad-audit-log-evidence-neither.yaml", "#AuditLog", true, ""},
		{"audit log mapping reference url with no scheme", "./test-data/bad-audit-log-url-no-scheme.yaml", "#AuditLog", true, ""},
		{"audit log mapping reference url with a non-alphabetic scheme", "./test-data/bad-audit-log-url-invalid-scheme.yaml", "#AuditLog", true, ""},
		{"audit log target uri with no scheme", "./test-data/bad-audit-log-uri-no-scheme.yaml", "#AuditLog", true, ""},

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

// TestSchemaValidationErrorsExclude covers what the table above cannot: an
// invalid document must fail for its own reason, and not also report errors
// against parts of it that are valid. Each row names a message that must not
// appear anywhere in the document's errors.
func TestSchemaValidationErrorsExclude(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		definition string
		excludes   string
	}{
		{"one bad mapping reference does not make audit evidence undeclared", "./test-data/bad-audit-log-url-no-scheme.yaml", "#AuditLog", "not declared in metadata.mapping-references"},
		{"one bad mapping reference does not make evaluation evidence undeclared", "./test-data/bad-evaluation-log-mapping-reference-url-no-scheme.yaml", "#EvaluationLog", "not declared in metadata.mapping-references"},
		{"evidence with neither payload nor source is reported as missing its source", "./test-data/bad-audit-log-evidence-neither.yaml", "#AuditLog", "invalid left-hand value"},
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

			// Validate stops at the first error. Unify and validate the whole
			// document instead, so every error it has is reported.
			file, err := cueyaml.Extract(tt.file, data)
			if err != nil {
				t.Fatalf("parse %s: %v", tt.file, err)
			}
			validationErr := def.Unify(schemaCtx.BuildFile(file)).Validate(cue.All(), cue.Concrete(true))
			if validationErr == nil {
				t.Fatal("expected validation error, got nil")
			}
			if details := cueerrors.Details(validationErr, nil); strings.Contains(details, tt.excludes) {
				t.Errorf("errors contain %q:\n%s", tt.excludes, details)
			}
		})
	}
}
