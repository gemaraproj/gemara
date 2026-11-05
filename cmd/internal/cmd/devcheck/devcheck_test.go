// SPDX-License-Identifier: Apache-2.0

package devcheck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckCueModule(t *testing.T) {
	// From internal/cmd/devcheck, root is 4 levels up: ../../../..
	root := filepath.Join("..", "..", "..", "..")
	res := checkCueModule(root)
	if !res.Passed {
		t.Fatalf("expected checkCueModule to pass, got error: %s", res.Message)
	}
	if res.Name != "CUE Module Structure" {
		t.Errorf("unexpected check name: %s", res.Name)
	}
}

func TestCheckCueModuleNonExistent(t *testing.T) {
	tmp := t.TempDir()
	res := checkCueModule(tmp)
	if res.Passed {
		t.Fatalf("expected checkCueModule to fail for empty dir")
	}
}

func TestCheckCueModuleMissingRequiredSchema(t *testing.T) {
	tmp := t.TempDir()
	// Create cue.mod but omit required schemas
	if err := os.MkdirAll(filepath.Join(tmp, "cue.mod"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "cue.mod", "module.cue"), []byte("module: test"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := checkCueModule(tmp)
	if res.Passed {
		t.Fatalf("expected checkCueModule to fail when core schemas are missing")
	}
}

func TestNewDevCheckCmdStructure(t *testing.T) {
	cmd := NewDevCheckCmd()
	if cmd.Use != "dev-check" {
		t.Errorf("expected command Use 'dev-check', got %q", cmd.Use)
	}

	flags := []string{"schema", "golden", "update-golden", "skip-slow"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s to exist", flag)
		}
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	crlf := []byte("line1\r\nline2\r\nline3\r\n")
	want := []byte("line1\nline2\nline3\n")
	got := normalizeLineEndings(crlf)
	if string(got) != string(want) {
		t.Errorf("normalizeLineEndings() = %q, want %q", got, want)
	}
}

func TestCheckOpenAPIGoldenDriftAndAutoUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	goldenFile := filepath.Join(tmpDir, "openapi.golden.yaml")

	// Write intentional mismatched content
	if err := os.WriteFile(goldenFile, []byte("mismatched: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repoRoot := filepath.Join("..", "..", "..", "..")
	opts := devCheckOptions{
		schemaDir:    repoRoot,
		goldenPath:   goldenFile,
		updateGolden: false,
	}

	// 1. Without --update-golden: must fail and detect drift
	res := checkOpenAPIGolden(opts)
	if res.Passed {
		t.Fatalf("expected drift to be detected, but check passed")
	}

	// 2. With --update-golden: must auto-update and pass
	opts.updateGolden = true
	resUpdate := checkOpenAPIGolden(opts)
	if !resUpdate.Passed {
		t.Fatalf("expected auto-update to succeed, got: %s", resUpdate.Message)
	}

	// 3. Verify file now matches freshly generated content
	opts.updateGolden = false
	resVerify := checkOpenAPIGolden(opts)
	if !resVerify.Passed {
		t.Fatalf("expected golden file to be in sync after update, got: %s", resVerify.Message)
	}
}
