// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunBreakingCheckCandidateAndAllowlistErrorPaths exercises CLI error propagation
// when candidate or baseline files are missing, candidate YAML is malformed,
// or a referenced allowlist cannot be loaded.
func TestRunBreakingCheckCandidateAndAllowlistErrorPaths(t *testing.T) {
	tmpDir := t.TempDir()
	validYaml := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validYaml, []byte("openapi: 3.0.0\ninfo:\n  title: Test\n  version: 1.0.0\npaths: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	nonExistent := filepath.Join(tmpDir, "nonexistent.yaml")

	// 1. Non-existent candidate path
	_, err := runBreakingCheck(breakingCheckOpts{
		basePath:      validYaml,
		candidatePath: nonExistent,
	})
	if err == nil || !strings.Contains(err.Error(), "loading candidate") {
		t.Fatalf("expected error loading missing candidate, got: %v", err)
	}

	// 2. Non-existent baseline path
	_, err = runBreakingCheck(breakingCheckOpts{
		basePath:      nonExistent,
		candidatePath: validYaml,
	})
	if err == nil || !strings.Contains(err.Error(), "loading baseline") {
		t.Fatalf("expected error loading missing baseline, got: %v", err)
	}

	// 3. Malformed candidate YAML
	badYaml := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(badYaml, []byte("openapi: [malformed"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = runBreakingCheck(breakingCheckOpts{
		basePath:      validYaml,
		candidatePath: badYaml,
	})
	if err == nil || !strings.Contains(err.Error(), "loading candidate") {
		t.Fatalf("expected error loading malformed candidate YAML, got: %v", err)
	}

	// 4. Missing allowlist file error propagation (verifies exact error wrapping from loadAllowlist)
	_, err = runBreakingCheck(breakingCheckOpts{
		basePath:      validYaml,
		candidatePath: validYaml,
		allowPath:     filepath.Join(tmpDir, "missing.allow"),
	})
	if err == nil || !strings.Contains(err.Error(), "opening allowlist") {
		t.Fatalf("expected error containing 'opening allowlist', got: %v", err)
	}
}

// TestBreakingCheckCmdMissingCUESchemaError exercises command execution error behavior
// when --schema points to a directory without valid CUE package files.
func TestBreakingCheckCmdMissingCUESchemaError(t *testing.T) {
	tmpDir := t.TempDir()
	specYaml := filepath.Join(tmpDir, "spec.yaml")
	if err := os.WriteFile(specYaml, []byte("openapi: 3.0.0\ninfo:\n  title: Test\n  version: 1.0.0\npaths: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewBreakingCheckCmd()
	cmd.SetArgs([]string{
		"--base", specYaml,
		"--schema", tmpDir,
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when schema dir has no CUE files, got nil")
	}
	if !strings.Contains(err.Error(), "generating candidate OpenAPI") {
		t.Fatalf("expected error containing 'generating candidate OpenAPI', got: %v", err)
	}
}
