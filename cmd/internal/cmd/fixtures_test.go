// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/goccy/go-yaml"
)

// repoFixtures mirrors the positive cases of test/schema_test.go, which
// validates these same files against the CUE. Every one of them is a document
// the schema is meant to accept, so the generated OpenAPI must accept it too:
// where the two disagree, the projection — not the fixture — is wrong.
//
// Keep this list in step with the `false` (wantErr) rows over there.
var repoFixtures = []struct {
	file   string // relative to the repository root
	schema string
}{
	{"test/test-data/good-ccc.yaml", "ControlCatalog"},
	{"test/test-data/good-ccc.json", "ControlCatalog"},
	{"test/test-data/good-osps.yml", "ControlCatalog"},
	{"test/test-data/good-lifecycle.yaml", "ControlCatalog"},
	{"test/test-data/nested-good-ccc.yaml", "ControlCatalog"},
	{"test/test-data/good-aigf.yaml", "GuidanceCatalog"},
	{"test/test-data/good-aigf-principles.yaml", "PrincipleCatalog"},
	{"test/test-data/good-aigf-vectors.yaml", "VectorCatalog"},
	{"test/test-data/good-threat-catalog.yaml", "ThreatCatalog"},
	{"test/test-data/good-capability-catalog.yaml", "CapabilityCatalog"},
	{"test/test-data/good-vector-owasp-mapping.yaml", "MappingDocument"},
	{"test/test-data/good-risk-catalog.yaml", "RiskCatalog"},
	{"test/test-data/good-policy.yaml", "Policy"},
	{"test/test-data/good-security-policy.yml", "Policy"},
	{"test/test-data/good-mapping-document.yaml", "MappingDocument"},
	{"test/test-data/good-aigf-nist-mapping.yaml", "MappingDocument"},
	{"test/test-data/good-lexicon.yaml", "Lexicon"},
	{"test/test-data/pvtr-baseline-scan.yaml", "EvaluationLog"},
	{"test/test-data/good-evaluation-log-unstarted.yaml", "EvaluationLog"},
	{"test/test-data/good-enforcement-log.yaml", "EnforcementLog"},
	{"test/test-data/good-audit-log.yaml", "AuditLog"},
	{"examples/ai-agent/ai-agent-capability-catalog.yaml", "CapabilityCatalog"},
	{"examples/ai-agent/atr-categories-to-capabilities-mapping.yaml", "MappingDocument"},
}

// jsonValue reads a YAML or JSON fixture and returns it with the types
// encoding/json would produce, which is what VisitJSON expects.
func jsonValue(t *testing.T, path string) any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read fixture: %v", err)
	}
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("cannot parse fixture: %v", err)
	}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("cannot re-encode fixture as JSON: %v", err)
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("cannot decode fixture JSON: %v", err)
	}
	return out
}

// knownFixtureGaps lists fixtures the generated document still rejects, with
// the reason. These are NOT bugs in the projection — they are places where CUE
// computes a value that OpenAPI has no way to express.
//
// evaluationlog.cue:27-29 supplies each assessment log's
// requirement.reference-id from the enclosing control:
//
//	"assessment-logs": [...{requirement: "reference-id": (control."reference-id")}]
//
// so a serialized log legitimately omits it. #EntryMapping requires
// reference-id, and relaxing that globally would weaken every other use of the
// definition (MappingDocument, EvidenceMapping) where it really is required.
// Expressing "required except when derived from an ancestor" needs a per-site
// override the projection does not have.
//
// Whether to model this differently in CUE is a schema decision, not a
// converter one. Until it is made, the gap is visible here rather than absent
// from the gate.
var knownFixtureGaps = map[string]string{
	"test/test-data/pvtr-baseline-scan.yaml":            "requirement.reference-id is derived from the control",
	"test/test-data/good-evaluation-log-unstarted.yaml": "requirement.reference-id is derived from the control",
}

// TestRepoFixturesValidateAgainstTheDocument is the gate the projection did not
// have: the coverage gates prove every CUE field reaches the document, but not
// that the result still accepts real data. A constraint can be present, well
// formed and wrong — `oneOf: [{enum: […]}, {}]` is exactly that, valid OpenAPI
// that rejects every value it names. Only running documents through the
// generated schema catches it.
func TestRepoFixturesValidateAgainstTheDocument(t *testing.T) {
	out := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := convertCUEToOpenAPI("../../..", out, ConvertOpts{}); err != nil {
		t.Fatal(err)
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(out)
	if err != nil {
		t.Fatalf("the generated document does not load as OpenAPI 3: %v", err)
	}
	if doc.Components == nil || len(doc.Components.Schemas) == 0 {
		t.Fatal("the generated document has no component schemas")
	}

	checked := 0
	for _, f := range repoFixtures {
		t.Run(filepath.Base(f.file)+"/"+f.schema, func(t *testing.T) {
			ref, ok := doc.Components.Schemas[f.schema]
			if !ok {
				t.Fatalf("schema %s is not in the generated document", f.schema)
			}
			value := jsonValue(t, filepath.Join("../../..", f.file))
			err := ref.Value.VisitJSON(value)
			if reason, known := knownFixtureGaps[f.file]; known {
				// A gap that has closed must not stay allowlisted, or the next
				// real regression here hides behind it.
				if err == nil {
					t.Errorf("this fixture now validates; remove it from knownFixtureGaps (%s)", reason)
				}
				return
			}
			if err != nil {
				t.Errorf("the CUE accepts this document but the generated schema rejects it:\n%v", err)
			}
		})
		checked++
	}
	// A typo in a path or a schema name would otherwise shrink this gate
	// silently, which is the failure mode the whole change exists to stop.
	if checked != len(repoFixtures) {
		t.Fatalf("checked %d fixtures, want %d", checked, len(repoFixtures))
	}
	if checked < 20 {
		t.Fatalf("only %d fixtures are gated; the list has been truncated", checked)
	}
}
