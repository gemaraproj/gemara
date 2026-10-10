// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
		// negative fixture stands for one rule; TestDigestFormat covers the full
		// digest format matrix.
		{"evidence that is addressed, addressed without a digest, and carried", "./test-data/good-evaluation-log-evidence.yaml", "#EvaluationLog", false, ""},
		{"sha256 digest of the wrong length", "./test-data/bad-evaluation-log-digest-sha256-length.yaml", "#EvaluationLog", true, "source.digest"},
		{"download-url with no URI scheme", "./test-data/bad-evaluation-log-download-url-no-scheme.yaml", "#EvaluationLog", true, "\"download-url\""},
		{"media-type with no subtype separator", "./test-data/bad-evaluation-log-media-type-malformed.yaml", "#EvaluationLog", true, "\"media-type\""},
		{"inline payload with a source digest", "./test-data/bad-evaluation-log-evidence-payload-with-digest.yaml", "#EvaluationLog", true, "inline payload cannot also have a source digest"},
		{"one evidence id naming two different items", "./test-data/bad-evaluation-log-evidence-id-reused.yaml", "#EvaluationLog", true, "names two different evidence items"},
		{"evidence source that is not declared", "./test-data/bad-evaluation-log-evidence-source-undeclared.yaml", "#EvaluationLog", true, "not declared in metadata.mapping-references"},

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

func TestEvidenceValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "allows inline payload with source provenance",
			input: `id: dependency-graph-snapshot
type: api-response
collected-at: "2026-02-10T15:05:00Z"
payload:
  dependencies: []
source:
  reference-id: github-api
  coordinate: /repos/acme/widget/dependency-graph/sbom
`,
		},
		{
			name: "rejects inline payload with retrievable source",
			input: `id: dependency-graph-snapshot
type: api-response
collected-at: "2026-02-10T15:05:00Z"
payload:
  dependencies: []
source:
  reference-id: github-api
  download-url: https://api.github.com/repos/acme/widget/dependency-graph/sbom
`,
			wantErr: true,
		},
		{
			name: "rejects inline payload with source digest",
			input: `id: dependency-graph-snapshot
type: api-response
collected-at: "2026-02-10T15:05:00Z"
payload:
  dependencies: []
source:
  reference-id: github-api
  digest: sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
`,
			wantErr: true,
		},
		{
			name: "rejects inline payload with source size",
			input: `id: dependency-graph-snapshot
type: api-response
collected-at: "2026-02-10T15:05:00Z"
payload:
  dependencies: []
source:
  reference-id: github-api
  size: 20
`,
			wantErr: true,
		},
		{
			name: "allows inline payload with source media-type",
			input: `id: dependency-graph-snapshot
type: api-response
collected-at: "2026-02-10T15:05:00Z"
payload: eyJkZXBlbmRlbmNpZXMiOltdfQ==
source:
  reference-id: github-api
  media-type: application/vnd.cyclonedx+json
`,
		},
		{
			name: "allows cited evidence with a digest and no address",
			input: `id: branch-protection-result
type: EvaluationLog
collected-at: "2026-02-10T15:05:00Z"
source:
  reference-id: evaluation-log
  entry-id: OSPS-AC-03.01
  digest: sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
`,
		},
	}

	def := schemaValue.LookupPath(cue.ParsePath("#_EvidenceStrict"))
	if def.Err() != nil {
		t.Fatalf("lookup #_EvidenceStrict: %v", def.Err())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cueyaml.Validate([]byte(tt.input), def)
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestDigestFormat(t *testing.T) {
	const hex64 = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	tests := []struct {
		name    string
		digest  string
		wantErr bool
	}{
		{"sha256 with 64 lowercase hex", "sha256:" + hex64, false},
		{"sha512 with 128 lowercase hex", "sha512:" + hex64 + hex64, false},
		{"blake3 with 64 lowercase hex", "blake3:" + hex64, false},
		// An algorithm this schema does not know is checked for grammar only, so a
		// misspelt name passes as unknown.
		{"unknown algorithm", "sha3-512:" + hex64 + hex64, false},
		{"misspelt sha256 passes as unknown", "sha-256:" + hex64, false},
		{"no algorithm separator", "sha256", true},
		{"uppercase algorithm name", "SHA256:" + hex64, true},
		{"sha256 one character short", "sha256:" + hex64[1:], true},
		{"sha256 in uppercase hex", "sha256:" + strings.ToUpper(hex64), true},
		{"sha512 with a sha256 length", "sha512:" + hex64, true},
		{"blake3 one character short", "blake3:" + hex64[1:], true},
	}

	def := schemaValue.LookupPath(cue.ParsePath("#Digest"))
	if def.Err() != nil {
		t.Fatalf("lookup #Digest: %v", def.Err())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := def.Unify(schemaValue.Context().CompileString(strconv.Quote(tt.digest))).Validate()
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// reverseKeys encodes an object with its keys in reverse order, where
// json.Marshal on the map itself would sort them.
func reverseKeys(object map[string]any) json.RawMessage {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	fields := make([]string, 0, len(keys))
	for _, key := range keys {
		value, _ := json.Marshal(object[key])
		fields = append(fields, strconv.Quote(key)+":"+string(value))
	}
	return json.RawMessage("{" + strings.Join(fields, ",") + "}")
}

// TestEvidenceIDWithinLog covers the rule that, within one log, an evidence id
// names exactly one evidence item. Each case edits a valid fixture in memory.
func TestEvidenceIDWithinLog(t *testing.T) {
	evaluationEvidence := func(doc map[string]any, assessment int) []any {
		evaluation := doc["evaluations"].([]any)[0].(map[string]any)
		return evaluation["assessment-logs"].([]any)[assessment].(map[string]any)["evidence"].([]any)
	}
	// The first audit result in the fixture carries no evidence, so skip it.
	auditEvidence := func(doc map[string]any, result int) []any {
		return doc["results"].([]any)[result+1].(map[string]any)["evidence"].([]any)
	}

	tests := []struct {
		name       string
		file       string
		definition string
		evidence   func(map[string]any, int) []any
		edit       func(first, second []any)
		wantErr    bool
	}{
		{
			name:       "evaluation log rejects two different items sharing an id",
			file:       "./test-data/good-evaluation-log-evidence.yaml",
			definition: "#EvaluationLog",
			evidence:   evaluationEvidence,
			edit: func(first, second []any) {
				second[0].(map[string]any)["id"] = first[0].(map[string]any)["id"]
			},
			wantErr: true,
		},
		{
			name:       "evaluation log allows the identical item repeated under one id",
			file:       "./test-data/good-evaluation-log-evidence.yaml",
			definition: "#EvaluationLog",
			evidence:   evaluationEvidence,
			edit:       func(first, second []any) { second[0] = first[0] },
		},
		{
			name:       "evaluation log allows the identical item repeated with its keys in another order",
			file:       "./test-data/good-evaluation-log-evidence.yaml",
			definition: "#EvaluationLog",
			evidence:   evaluationEvidence,
			edit:       func(first, second []any) { second[0] = reverseKeys(first[0].(map[string]any)) },
		},
		{
			name:       "audit log rejects two different items sharing an id",
			file:       "./test-data/good-audit-log.yaml",
			definition: "#AuditLog",
			evidence:   auditEvidence,
			edit: func(first, second []any) {
				second[0].(map[string]any)["id"] = first[0].(map[string]any)["id"]
			},
			wantErr: true,
		},
		{
			name:       "audit log allows the identical item repeated under one id",
			file:       "./test-data/good-audit-log.yaml",
			definition: "#AuditLog",
			evidence:   auditEvidence,
			edit:       func(first, second []any) { second[0] = first[0] },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			file, err := cueyaml.Extract(tt.file, data)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			var doc map[string]any
			if err := schemaCtx.BuildFile(file).Decode(&doc); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			tt.edit(tt.evidence(doc, 0), tt.evidence(doc, 1))
			edited, err := json.Marshal(doc)
			if err != nil {
				t.Fatalf("encode edited fixture: %v", err)
			}

			def := schemaValue.LookupPath(cue.ParsePath(tt.definition))
			err = cuejson.Validate(edited, def)
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestEvidenceSourceReference covers the rule that an evidence source names a
// mapping reference declared in its log. Each case edits a valid fixture in memory.
func TestEvidenceSourceReference(t *testing.T) {
	firstEvaluationEvidence := func(doc map[string]any) map[string]any {
		evaluation := doc["evaluations"].([]any)[0].(map[string]any)
		assessment := evaluation["assessment-logs"].([]any)[0].(map[string]any)
		return assessment["evidence"].([]any)[0].(map[string]any)
	}
	// The first audit result in the fixture carries no evidence, so use the second.
	firstAuditEvidence := func(doc map[string]any) map[string]any {
		return doc["results"].([]any)[1].(map[string]any)["evidence"].([]any)[0].(map[string]any)
	}
	undeclare := func(item map[string]any) {
		item["source"].(map[string]any)["reference-id"] = "not-declared"
	}

	tests := []struct {
		name       string
		file       string
		definition string
		edit       func(doc map[string]any)
		wantErr    bool
	}{
		{
			name:       "evaluation log rejects an evidence source that is not declared",
			file:       "./test-data/pvtr-baseline-scan.yaml",
			definition: "#EvaluationLog",
			edit:       func(doc map[string]any) { undeclare(firstEvaluationEvidence(doc)) },
			wantErr:    true,
		},
		{
			name:       "evaluation log allows inline evidence that names no source",
			file:       "./test-data/pvtr-baseline-scan.yaml",
			definition: "#EvaluationLog",
			edit:       func(doc map[string]any) { delete(firstEvaluationEvidence(doc), "source") },
		},
		{
			name:       "audit log rejects an evidence source that is not declared",
			file:       "./test-data/good-audit-log.yaml",
			definition: "#AuditLog",
			edit:       func(doc map[string]any) { undeclare(firstAuditEvidence(doc)) },
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			file, err := cueyaml.Extract(tt.file, data)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			var doc map[string]any
			if err := schemaCtx.BuildFile(file).Decode(&doc); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			tt.edit(doc)
			edited, err := json.Marshal(doc)
			if err != nil {
				t.Fatalf("encode edited fixture: %v", err)
			}

			def := schemaValue.LookupPath(cue.ParsePath(tt.definition))
			err = cuejson.Validate(edited, def)
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}
