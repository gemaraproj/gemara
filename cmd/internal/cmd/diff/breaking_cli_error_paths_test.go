// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBreakingCheckMissingCandidateOrBaseline(t *testing.T) {
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

	// 4. Missing allowlist file error propagation
	_, err = runBreakingCheck(breakingCheckOpts{
		basePath:      validYaml,
		candidatePath: validYaml,
		allowPath:     filepath.Join(tmpDir, "missing.allow"),
	})
	if err == nil {
		t.Fatalf("expected error loading non-existent allowlist, got nil")
	}

	// 5. Successful zero-breaking pass when candidate equals baseline with allowlist
	allowFile := filepath.Join(tmpDir, "test.allow")
	if err := os.WriteFile(allowFile, []byte("# comment\ncheck-id SomeSchema\n"), 0644); err != nil {
		t.Fatal(err)
	}
	code, err := runBreakingCheck(breakingCheckOpts{
		basePath:      validYaml,
		candidatePath: validYaml,
		allowPath:     allowFile,
	})
	if err != nil {
		t.Fatalf("unexpected error running breaking check on identical spec: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0 for identical spec, got %d", code)
	}
}

func TestBreakingCheckCmdExecutionWithAllowlist(t *testing.T) {
	tmpDir := t.TempDir()
	specYaml := filepath.Join(tmpDir, "spec.yaml")
	content := `
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
components:
  schemas:
    Item:
      type: object
      properties:
        name:
          type: string
`
	if err := os.WriteFile(specYaml, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	allowFile := filepath.Join(tmpDir, "allowlist.txt")
	if err := os.WriteFile(allowFile, []byte("# empty allowlist\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewBreakingCheckCmd()
	cmd.SetArgs([]string{
		"--base", specYaml,
		"--allow", allowFile,
		"--schema", tmpDir,
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	// When schema dir has no CUE files, CUE loading fails properly
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when schema dir has no CUE files, got nil")
	}
	if !strings.Contains(err.Error(), "generating candidate OpenAPI") {
		t.Fatalf("expected 'generating candidate OpenAPI' error, got: %v", err)
	}
}
