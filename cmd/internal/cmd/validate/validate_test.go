// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCleanArtifacts(t *testing.T) {
	tempDir := t.TempDir()

	catalogContent := `metadata:
  id: test-catalog
  type: ControlCatalog
controls:
  - id: CTL-001
    title: Access Control
`
	evalContent := `metadata:
  id: test-eval-log
  type: EvaluationLog
  mapping-references:
    - id: test-catalog
      url: https://example.com/catalog
evaluations:
  - name: Access Control Eval
    control:
      reference-id: test-catalog
      entry-id: CTL-001
    assessment-logs:
      - requirement:
          reference-id: test-catalog
          entry-id: CTL-001
`
	catFile := filepath.Join(tempDir, "catalog.yaml")
	evalFile := filepath.Join(tempDir, "eval.yaml")

	if err := os.WriteFile(catFile, []byte(catalogContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evalFile, []byte(evalContent), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := RunValidation(ValidateOpts{
		Files: []string{catFile, evalFile},
	})
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if report.Errors != 0 || report.Warnings != 0 {
		t.Fatalf("expected 0 errors and warnings, got %d errors, %d warnings: %+v",
			report.Errors, report.Warnings, report.Diagnostics)
	}
}

func TestValidateBrokenForeignKey(t *testing.T) {
	tempDir := t.TempDir()

	catalogContent := `metadata:
  id: test-catalog
  type: ControlCatalog
controls:
  - id: CTL-001
    title: Access Control
`
	evalContent := `metadata:
  id: test-eval-log
  type: EvaluationLog
  mapping-references:
    - id: test-catalog
      url: https://example.com/catalog
evaluations:
  - name: Access Control Eval
    control:
      reference-id: test-catalog
      entry-id: CTL-NONEXISTENT
`
	catFile := filepath.Join(tempDir, "catalog.yaml")
	evalFile := filepath.Join(tempDir, "eval.yaml")

	_ = os.WriteFile(catFile, []byte(catalogContent), 0644)
	_ = os.WriteFile(evalFile, []byte(evalContent), 0644)

	report, err := RunValidation(ValidateOpts{
		Files: []string{catFile, evalFile},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Errors != 1 {
		t.Fatalf("expected 1 error for broken foreign key, got %d", report.Errors)
	}

	if report.Diagnostics[0].Rule != "broken-foreign-key" {
		t.Fatalf("expected rule 'broken-foreign-key', got %q", report.Diagnostics[0].Rule)
	}
}

func TestValidateMismatchedAssessmentReference(t *testing.T) {
	tempDir := t.TempDir()

	evalContent := `metadata:
  id: test-eval-log
  type: EvaluationLog
  mapping-references:
    - id: cat-A
      url: https://example.com/a
    - id: cat-B
      url: https://example.com/b
evaluations:
  - name: Eval
    control:
      reference-id: cat-A
      entry-id: CTL-001
    assessment-logs:
      - requirement:
          reference-id: cat-B
          entry-id: CTL-001
`
	evalFile := filepath.Join(tempDir, "eval.yaml")
	_ = os.WriteFile(evalFile, []byte(evalContent), 0644)

	report, err := RunValidation(ValidateOpts{
		Files: []string{evalFile},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Errors != 1 {
		t.Fatalf("expected 1 error for mismatched assessment reference, got %d", report.Errors)
	}

	if report.Diagnostics[0].Rule != "mismatched-assessment-reference" {
		t.Fatalf("expected rule 'mismatched-assessment-reference', got %q", report.Diagnostics[0].Rule)
	}
}

func TestValidateMappingDocumentSemantics(t *testing.T) {
	tempDir := t.TempDir()

	mappingContent := `metadata:
  id: test-mapping
  type: MappingDocument
mappings:
  - id: MAP-001
    source: SEC-1
    relationship: no-match
    targets:
      - entry-id: TGT-1
  - id: MAP-002
    source: SEC-2
    relationship: implements
    targets: []
`
	mapFile := filepath.Join(tempDir, "mapping.yaml")
	_ = os.WriteFile(mapFile, []byte(mappingContent), 0644)

	report, err := RunValidation(ValidateOpts{
		Files: []string{mapFile},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Errors != 2 {
		t.Fatalf("expected 2 errors for mapping semantics, got %d", report.Errors)
	}
}

func TestValidateCmdExecution(t *testing.T) {
	cmd := NewValidateCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected help command error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "validate") {
		t.Errorf("expected help output to contain 'validate', got %q", out)
	}
}
