// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBreakingCheckRequiresBase(t *testing.T) {
	// No basePath supplied: the check must error rather than silently pass.
	if _, err := runBreakingCheck(breakingCheckOpts{candidatePath: "testdata/base.yaml"}); err == nil {
		t.Fatalf("expected error when --base is missing, got nil")
	}
}

// TestRunBreakingCheckAllowSuppresses exercises the full --allow path: a real
// detected breaking change is suppressed when its "<id> <schema>" entry is in
// the allowlist, and the check then passes.
func TestRunBreakingCheckAllowSuppresses(t *testing.T) {
	// Without an allowlist, dropping a required property is breaking (exit 1).
	code, err := runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/rev_remove.yaml",
	})
	if err != nil {
		t.Fatalf("runBreakingCheck: %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1 without allowlist, got %d", code)
	}

	// Allowlisting that exact "<id> <schema>" pair suppresses it (exit 0).
	allow := filepath.Join(t.TempDir(), ".oasdiff-allow")
	content := "# result became optional is an accepted, reviewed change\nresponse-property-became-optional ControlEvaluation\n"
	if err := os.WriteFile(allow, []byte(content), 0644); err != nil {
		t.Fatalf("writing allowlist: %v", err)
	}
	code, err = runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/rev_remove.yaml",
		allowPath:     allow,
	})
	if err != nil {
		t.Fatalf("runBreakingCheck (allowed): %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0 with matching allowlist entry, got %d", code)
	}
}

func TestRunBreakingCheckLocalBaseline(t *testing.T) {
	// base vs a mutated schema, both provided as local OpenAPI files, exercises
	// the diff+exit-code path without touching the network or CUE generation.
	code, err := runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/rev_remove.yaml",
	})
	if err != nil {
		t.Fatalf("runBreakingCheck: %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1 for a breaking change, got %d", code)
	}
	code, err = runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/base.yaml",
	})
	if err != nil {
		t.Fatalf("runBreakingCheck: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0 for no change, got %d", code)
	}
}

func TestNewBreakingCheckCmdStructure(t *testing.T) {
	cmd := NewBreakingCheckCmd()
	if cmd.Use != "breaking-check" {
		t.Fatalf("expected command Use to be 'breaking-check', got %q", cmd.Use)
	}

	flags := []string{"schema", "base", "allow"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined on breaking-check command", f)
		}
	}
}

func TestBreakingCheckCmdHelp(t *testing.T) {
	cmd := NewBreakingCheckCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	t.Cleanup(func() {
		cmd.SetOut(nil)
		cmd.SetErr(nil)
		cmd.SetArgs(nil)
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error running breaking-check --help: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "breaking-check") {
		t.Errorf("expected help output to mention 'breaking-check', got:\n%s", out)
	}
	if !strings.Contains(out, "--base") {
		t.Errorf("expected help output to document '--base', got:\n%s", out)
	}
}

func TestBreakingCheckCmdMissingBase(t *testing.T) {
	cmd := NewBreakingCheckCmd()
	cmd.SetArgs([]string{})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when --base is omitted, got nil")
	}
	if !strings.Contains(err.Error(), "--base is required") {
		t.Fatalf("expected error message to mention '--base is required', got: %v", err)
	}
}

func TestBreakingCheckCmdExecutionIdentical(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test invoking CUE generation in -short mode")
	}

	tmpDir := t.TempDir()
	baseYaml := filepath.Join(tmpDir, "base.yaml")
	if err := convertCUEToOpenAPI("../../../..", baseYaml); err != nil {
		t.Fatalf("convertCUEToOpenAPI failed: %v", err)
	}

	cmd := NewBreakingCheckCmd()
	cmd.SetArgs([]string{
		"--base", baseYaml,
		"--schema", "../../../..",
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected breaking-check to pass on identical base and schema, got: %v", err)
	}
}

func TestBreakingCheckCmdBreakingDetected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test invoking CUE generation in -short mode")
	}

	cmd := NewBreakingCheckCmd()
	cmd.SetArgs([]string{
		"--base", "testdata/base.yaml",
		"--schema", "../../../..",
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected breaking changes to be detected, got nil")
	}
	if !strings.Contains(err.Error(), "breaking changes detected") {
		t.Fatalf("expected error 'breaking changes detected', got: %v", err)
	}
}

func TestBreakingCheckCmdInvalidSchema(t *testing.T) {
	cmd := NewBreakingCheckCmd()
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	cmd.SetArgs([]string{
		"--base", "testdata/base.yaml",
		"--schema", nonExistent,
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when running breaking-check on nonexistent schema dir, got nil")
	}
}
