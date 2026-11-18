// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewValidateCmdStructure(t *testing.T) {
	cmd := NewValidateCmd()
	if cmd.Use != "validate [paths...]" {
		t.Errorf("expected Use to be 'validate [paths...]', got %q", cmd.Use)
	}

	if cmd.Flags().Lookup("strict") == nil {
		t.Errorf("expected 'strict' flag on validate command")
	}
	if cmd.Flags().Lookup("format") == nil {
		t.Errorf("expected 'format' flag on validate command")
	}
	if cmd.Flags().Lookup("quiet") == nil {
		t.Errorf("expected 'quiet' flag on validate command")
	}
}

func TestRunValidate_JSONOutput(t *testing.T) {
	root := repoRoot()
	files := []string{
		filepath.Join(root, "test", "test-data", "good-mapping-document.yaml"),
		filepath.Join(root, "test", "test-data", "good-osps.yml"),
	}

	var stdout, stderr bytes.Buffer
	err := RunValidate(&stdout, &stderr, files, Options{Format: "json"})
	if err != nil {
		t.Fatalf("unexpected error running validate in JSON mode: %v", err)
	}

	var res ValidationResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v\nOutput: %s", err, stdout.String())
	}

	if res.TotalFiles != 2 {
		t.Errorf("expected totalFiles 2, got %d", res.TotalFiles)
	}
	if !res.Passed {
		t.Errorf("expected validation to pass, but result.Passed was false")
	}
}

func TestRunValidate_TextOutputSuccess(t *testing.T) {
	root := repoRoot()
	files := []string{
		filepath.Join(root, "test", "test-data", "good-mapping-document.yaml"),
		filepath.Join(root, "test", "test-data", "good-osps.yml"),
	}

	var stdout, stderr bytes.Buffer
	err := RunValidate(&stdout, &stderr, files, Options{Format: "text"})
	if err != nil {
		t.Fatalf("unexpected error running validate: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "SUCCESS:") {
		t.Errorf("expected success message in stdout, got:\n%s", out)
	}
}

func TestRunValidate_FailureReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad-mapping.yaml")

	content := `title: Broken Mapping
metadata:
  id: MAP-BAD-01
  type: MappingDocument
source-reference:
  reference-id: UNDECLARED
  entry-type: Control
target-reference:
  reference-id: UNDECLARED
  entry-type: Guideline
mappings:
  - id: M-01
    source: NO-SUCH-ID
    relationship: relates-to
    targets:
      - entry-id: T-01
`
	if err := os.WriteFile(badFile, []byte(content), 0644); err != nil {
		t.Fatalf("write bad fixture: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := RunValidate(&stdout, &stderr, []string{badFile}, Options{Format: "text"})
	if err == nil {
		t.Fatalf("expected error running validate on broken fixture, got nil")
	}

	errOut := stderr.String()
	if !strings.Contains(errOut, "undeclared-mapping-reference") {
		t.Errorf("expected stderr to report undeclared-mapping-reference, got:\n%s", errOut)
	}
}

func TestRunValidate_EmptyPaths(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := RunValidate(&stdout, &stderr, []string{}, Options{})
	if err == nil {
		t.Fatalf("expected error with empty paths, got nil")
	}
}

func TestRunValidate_JSONOutputFailure(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad-mapping.yaml")

	content := `title: Broken Mapping
metadata:
  id: MAP-BAD-01
  type: MappingDocument
source-reference:
  reference-id: UNDECLARED
  entry-type: Control
target-reference:
  reference-id: UNDECLARED
  entry-type: Guideline
mappings:
  - id: M-01
    source: NO-SUCH-ID
    relationship: relates-to
    targets:
      - entry-id: T-01
`
	if err := os.WriteFile(badFile, []byte(content), 0644); err != nil {
		t.Fatalf("write bad fixture: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := RunValidate(&stdout, &stderr, []string{badFile}, Options{Format: "json"})
	if err == nil {
		t.Fatalf("expected error running validate on broken fixture in JSON mode")
	}

	// In JSON mode, stderr must not receive diagnostic error streams (stdout has the JSON)
	if stderr.Len() != 0 {
		t.Errorf("expected clean stderr in JSON mode, got: %q", stderr.String())
	}

	var res ValidationResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("stdout should be valid JSON even on validation failure: %v\nOutput: %s", err, stdout.String())
	}

	if res.Passed {
		t.Errorf("expected res.Passed to be false")
	}
	if len(res.Diagnostics) == 0 {
		t.Errorf("expected diagnostics in JSON output")
	}
}
