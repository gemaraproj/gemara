// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"path/filepath"
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
