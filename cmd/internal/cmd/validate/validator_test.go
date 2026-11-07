// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot() string {
	return filepath.Join("..", "..", "..", "..")
}

func TestValidate_GoodBundle(t *testing.T) {
	root := repoRoot()
	files := []string{
		filepath.Join(root, "test", "test-data", "good-mapping-document.yaml"),
		filepath.Join(root, "test", "test-data", "good-osps.yml"),
	}

	docs, err := LoadPaths(files)
	if err != nil {
		t.Fatalf("LoadPaths failed: %v", err)
	}

	if len(docs) != 2 {
		t.Fatalf("expected 2 documents loaded, got %d", len(docs))
	}

	res := Validate(docs, Options{Strict: false})
	for _, diag := range res.Diagnostics {
		if diag.Severity == SeverityError {
			t.Errorf("unexpected validation error: %s (%s)", diag.Message, diag.Code)
		}
	}

	if !res.Passed {
		t.Errorf("expected bundle to pass validation, but failed")
	}
}

func TestValidate_MakeValidateBundle(t *testing.T) {
	root := repoRoot()
	files := []string{
		filepath.Join(root, "test", "test-data", "good-mapping-document.yaml"),
		filepath.Join(root, "test", "test-data", "good-osps.yml"),
		filepath.Join(root, "test", "test-data", "good-enforcement-log.yaml"),
	}

	docs, err := LoadPaths(files)
	if err != nil {
		t.Fatalf("LoadPaths failed: %v", err)
	}

	if len(docs) != 3 {
		t.Fatalf("expected 3 documents loaded, got %d", len(docs))
	}

	res := Validate(docs, Options{Strict: false})
	for _, diag := range res.Diagnostics {
		if diag.Severity == SeverityError {
			t.Errorf("unexpected validation error: %s (%s)", diag.Message, diag.Code)
		}
	}

	if !res.Passed {
		t.Errorf("expected make validate bundle to pass validation cleanly, but failed")
	}
}

func TestValidate_MappingDocumentNoDuplicateUndeclaredError(t *testing.T) {
	yamlData := `metadata:
  id: MAP-TEST
  type: MappingDocument
  version: 1.0.0
  mapping-references: []
source-reference:
  reference-id: UNDECLARED-SRC
target-reference:
  reference-id: UNDECLARED-TGT
mappings: []
`
	tmpDir := t.TempDir()
	mapFile := filepath.Join(tmpDir, "map.yaml")
	if err := os.WriteFile(mapFile, []byte(yamlData), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	doc, err := LoadDocument(mapFile)
	if err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	res := Validate([]*Document{doc}, Options{Strict: false})
	srcCount := 0
	tgtCount := 0
	for _, d := range res.Diagnostics {
		if d.Code == "undeclared-mapping-reference" {
			if strings.Contains(d.Message, "UNDECLARED-SRC") {
				srcCount++
			}
			if strings.Contains(d.Message, "UNDECLARED-TGT") {
				tgtCount++
			}
		}
	}
	if srcCount != 1 {
		t.Errorf("expected UNDECLARED-SRC to be reported once, got %d", srcCount)
	}
	if tgtCount != 1 {
		t.Errorf("expected UNDECLARED-TGT to be reported once, got %d", tgtCount)
	}
}

func TestValidate_UndeclaredMappingReference(t *testing.T) {
	doc := &Document{
		FilePath:          "test-mapping.yaml",
		ID:                "TEST-MAP-001",
		Type:              "MappingDocument",
		MappingReferences: make(map[string]MappingReference),
		ReferenceUsages: []ReferenceUsage{
			{
				ReferenceID: "UNDECLARED-REF",
				Context:     "source-reference",
			},
		},
	}

	res := Validate([]*Document{doc}, Options{})
	if res.Passed {
		t.Fatalf("expected validation failure due to undeclared mapping reference")
	}

	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "undeclared-mapping-reference" && d.Severity == SeverityError {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected diagnostic with code 'undeclared-mapping-reference', got: %+v", res.Diagnostics)
	}
}

func TestValidate_DanglingEntryReference(t *testing.T) {
	targetCatalog := &Document{
		FilePath: "controls.yaml",
		ID:       "OSPS",
		Type:     "ControlCatalog",
		DeclaredEntities: map[string]bool{
			"OSPS-QA-01": true,
			"OSPS-QA-02": true,
		},
	}

	mappingDoc := &Document{
		FilePath: "mapping.yaml",
		ID:       "MAP-01",
		Type:     "MappingDocument",
		MappingReferences: map[string]MappingReference{
			"OSPS": {ID: "OSPS", Title: "OSPS Baseline"},
		},
		EntityUsages: []EntityUsage{
			{
				ReferenceID: "OSPS",
				EntryID:     "OSPS-DOES-NOT-EXIST",
				Context:     "mappings.source",
			},
		},
	}

	res := Validate([]*Document{mappingDoc, targetCatalog}, Options{})
	if res.Passed {
		t.Fatalf("expected validation failure due to dangling entry reference")
	}

	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "dangling-entry-reference" && d.Severity == SeverityError {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected diagnostic with code 'dangling-entry-reference', got: %+v", res.Diagnostics)
	}
}

func TestValidate_DanglingSubEntryReference(t *testing.T) {
	targetCatalog := &Document{
		FilePath: "controls.yaml",
		ID:       "OSPS",
		Type:     "ControlCatalog",
		DeclaredEntities: map[string]bool{
			"OSPS-QA-02": true,
		},
		SubEntities: map[string]map[string]bool{
			"OSPS-QA-02": {
				"OSPS-QA-02.01": true,
			},
		},
	}

	evalLog := &Document{
		FilePath: "eval.yaml",
		ID:       "EVAL-01",
		Type:     "EvaluationLog",
		MappingReferences: map[string]MappingReference{
			"OSPS": {ID: "OSPS", Title: "OSPS Baseline"},
		},
		EntityUsages: []EntityUsage{
			{
				ReferenceID: "OSPS",
				EntryID:     "OSPS-QA-02",
				SubEntryID:  "OSPS-QA-02.99-NONEXISTENT",
				Context:     "evaluations.assessment-logs.requirement",
			},
		},
	}

	res := Validate([]*Document{evalLog, targetCatalog}, Options{})
	if res.Passed {
		t.Fatalf("expected validation failure due to dangling subentry reference")
	}

	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "dangling-subentry-reference" && d.Severity == SeverityError {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected diagnostic with code 'dangling-subentry-reference', got: %+v", res.Diagnostics)
	}
}

func TestValidate_StrictUnresolvedWarning(t *testing.T) {
	doc := &Document{
		FilePath: "eval.yaml",
		ID:       "EVAL-01",
		Type:     "EvaluationLog",
		MappingReferences: map[string]MappingReference{
			"EXTERNAL-CAT": {ID: "EXTERNAL-CAT", Title: "External"},
		},
		EntityUsages: []EntityUsage{
			{
				ReferenceID: "EXTERNAL-CAT",
				EntryID:     "SOME-ID",
				Context:     "evaluations.control",
			},
		},
	}

	// Non-strict: unresolvable external catalog is tolerated
	resNonStrict := Validate([]*Document{doc}, Options{Strict: false})
	if !resNonStrict.Passed {
		t.Errorf("expected non-strict mode to pass on external unresolvable catalog, got: %+v", resNonStrict.Diagnostics)
	}

	// Strict: unresolvable external catalog fails validation
	resStrict := Validate([]*Document{doc}, Options{Strict: true})
	if resStrict.Passed {
		t.Fatalf("expected strict mode to fail on unresolvable catalog reference")
	}
}

func TestValidate_DeterministicCatalogPathResolution(t *testing.T) {
	// Nested doc with "osps" in its folder path, but a different ID and base filename
	nestedUnrelatedDoc := &Document{
		FilePath: "some/nested/osps-directory/unrelated.yaml",
		ID:       "UNRELATED",
		Type:     "ControlCatalog",
		DeclaredEntities: map[string]bool{
			"UNRELATED-01": true,
		},
	}

	// Exact target catalog matching "osps"
	targetCatalog := &Document{
		FilePath: "catalogs/good-osps.yml",
		ID:       "OSPS",
		Type:     "ControlCatalog",
		DeclaredEntities: map[string]bool{
			"OSPS-QA-01": true,
		},
	}

	consumerDoc := &Document{
		FilePath: "mapping.yaml",
		ID:       "MAP-01",
		Type:     "MappingDocument",
		MappingReferences: map[string]MappingReference{
			"osps": {ID: "osps", Title: "OSPS Baseline"},
		},
		EntityUsages: []EntityUsage{
			{
				ReferenceID: "osps",
				EntryID:     "OSPS-QA-01",
				Context:     "mappings.source",
			},
		},
	}

	res := Validate([]*Document{consumerDoc, nestedUnrelatedDoc, targetCatalog}, Options{})
	if !res.Passed {
		t.Fatalf("expected deterministic resolution to targetCatalog, but got diagnostics: %+v", res.Diagnostics)
	}
}

func TestValidate_EvaluationLogEnforcementTolerance_Issue499(t *testing.T) {
	// EvaluationLog target (Issue #499: #ControlEvaluation lacks top-level id)
	evalLog := &Document{
		FilePath:         "eval-log.yaml",
		ID:               "EVAL-LOG-01",
		Type:             "EvaluationLog",
		DeclaredEntities: map[string]bool{},
	}

	// Enforcement log citing evaluation log
	enforcementLog := &Document{
		FilePath: "enforcement-log.yaml",
		ID:       "ENF-01",
		Type:     "EnforcementLog",
		MappingReferences: map[string]MappingReference{
			"EVAL-LOG-01": {ID: "EVAL-LOG-01", Title: "Evaluation Log"},
		},
		EntityUsages: []EntityUsage{
			{
				ReferenceID: "EVAL-LOG-01",
				EntryID:     "finding-123",
				Context:     "actions.justification.assessments.log",
			},
		},
	}

	res := Validate([]*Document{enforcementLog, evalLog}, Options{})
	if !res.Passed {
		t.Fatalf("expected EvaluationLog reference to be tolerated per #499, but got diagnostics: %+v", res.Diagnostics)
	}
}

func TestValidate_MultiEntryMappingWithReferenceID(t *testing.T) {
	root := repoRoot()
	files := []string{
		filepath.Join(root, "test", "test-data", "good-threat-catalog.yaml"),
		filepath.Join(root, "test", "test-data", "good-capability-catalog.yaml"),
	}

	docs, err := LoadPaths(files)
	if err != nil {
		t.Fatalf("LoadPaths failed: %v", err)
	}

	// Verify that capabilities entries with reference-id are extracted into EntityUsages
	var threatDoc *Document
	for _, d := range docs {
		if d.Type == "ThreatCatalog" {
			threatDoc = d
			break
		}
	}
	if threatDoc == nil {
		t.Fatalf("threat catalog document not found")
	}

	foundCap002 := false
	for _, u := range threatDoc.EntityUsages {
		if u.ReferenceID == "EXAMPLE-CAPABILITY-CATALOG" && u.EntryID == "CAP-002" {
			foundCap002 = true
			break
		}
	}
	if !foundCap002 {
		t.Errorf("expected CAP-002 from MultiEntryMapping entries to be extracted into EntityUsages")
	}

	res := Validate(docs, Options{Strict: false})
	for _, d := range res.Diagnostics {
		if d.Severity == SeverityError {
			t.Errorf("unexpected error validating threat and capability catalogs: %s (%s)", d.Message, d.Code)
		}
	}
}

func TestValidate_PolicyNestedRiskMapping(t *testing.T) {
	root := repoRoot()
	riskCatalogPath := filepath.Join(root, "test", "test-data", "good-risk-catalog.yaml")

	docs, err := LoadPaths([]string{riskCatalogPath})
	if err != nil {
		t.Fatalf("LoadPaths failed: %v", err)
	}

	policyDoc := &Document{
		FilePath: "policy.yaml",
		ID:       "TEST-POLICY",
		Type:     "Policy",
		MappingReferences: map[string]MappingReference{
			"EXAMPLE-RISK-CATALOG": {
				ID:    "EXAMPLE-RISK-CATALOG",
				Title: "Example Risk Catalog",
			},
		},
		ReferenceUsages: []ReferenceUsage{
			{ReferenceID: "EXAMPLE-RISK-CATALOG", Context: "risks.mitigated"},
		},
		EntityUsages: []EntityUsage{
			{ReferenceID: "EXAMPLE-RISK-CATALOG", EntryID: "NON-EXISTENT-RISK", Context: "risks.mitigated"},
		},
	}

	allDocs := append(docs, policyDoc)
	res := Validate(allDocs, Options{Strict: true})
	if res.Passed {
		t.Fatalf("expected validation failure for non-existent risk in policy")
	}

	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "dangling-entry-reference" && d.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected dangling-entry-reference diagnostic for policy risk, got: %+v", res.Diagnostics)
	}
}

func TestValidate_GuidanceCatalogExtendsReplacedByAndExemptions(t *testing.T) {
	rootMap := map[string]any{
		"metadata": map[string]any{
			"id":             "GUIDANCE-01",
			"type":           "GuidanceCatalog",
			"gemara-version": "1.1.0",
			"version":        "1.0.0",
			"description":    "Guidance catalog test",
			"mapping-references": []any{
				map[string]any{"id": "BASE-GUIDE", "title": "Base Guidance"},
				map[string]any{"id": "ALT-GUIDE", "title": "Alt Guidance"},
			},
		},
		"title": "Guidance Catalog",
		"guidelines": []any{
			map[string]any{
				"id":    "G-01",
				"title": "First Guideline",
				"extends": map[string]any{
					"reference-id": "BASE-GUIDE",
					"entry-id":     "BG-01",
				},
				"replaced-by": map[string]any{
					"reference-id": "BASE-GUIDE",
					"entry-id":     "BG-02",
				},
			},
		},
		"exemptions": []any{
			map[string]any{
				"id":     "EX-01",
				"reason": "Temporary exemption",
				"redirect": map[string]any{
					"reference-id": "ALT-GUIDE",
					"entries": []any{
						map[string]any{"reference-id": "ALT-01"},
					},
				},
			},
		},
	}

	doc := &Document{
		FilePath: "guidance.yaml",
		ID:       "GUIDANCE-01",
		Type:     "GuidanceCatalog",
		MappingReferences: map[string]MappingReference{
			"BASE-GUIDE": {ID: "BASE-GUIDE", Title: "Base Guidance"},
			"ALT-GUIDE":  {ID: "ALT-GUIDE", Title: "Alt Guidance"},
		},
		DeclaredEntities: make(map[string]bool),
		SubEntities:      make(map[string]map[string]bool),
	}
	extractDocumentContents(doc, rootMap)

	// Check that extends and replaced-by were registered
	hasExtends := false
	hasReplacedBy := false
	hasRedirect := false
	for _, u := range doc.EntityUsages {
		if u.ReferenceID == "BASE-GUIDE" && u.EntryID == "BG-01" {
			hasExtends = true
		}
		if u.ReferenceID == "BASE-GUIDE" && u.EntryID == "BG-02" {
			hasReplacedBy = true
		}
		if u.ReferenceID == "ALT-GUIDE" && u.EntryID == "ALT-01" {
			hasRedirect = true
		}
	}

	if !hasExtends {
		t.Errorf("expected guideline extends entity usage to be extracted")
	}
	if !hasReplacedBy {
		t.Errorf("expected guideline replaced-by entity usage to be extracted")
	}
	if !hasRedirect {
		t.Errorf("expected exemption redirect entity usage to be extracted")
	}
}
