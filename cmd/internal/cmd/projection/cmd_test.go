// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCue2OpenAPICmdStructure(t *testing.T) {
	cmd := NewCue2OpenAPICmd()
	if cmd.Use != "cue2openapi" {
		t.Fatalf("expected Use to be 'cue2openapi', got %q", cmd.Use)
	}

	flags := []string{"schema", "output", "manifest", "root", "version", "title"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag --%s to be defined on cue2openapi command", f)
		}
	}
}

func TestCue2OpenAPICmdHelp(t *testing.T) {
	cmd := NewCue2OpenAPICmd()
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
		t.Fatalf("unexpected error running cue2openapi --help: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "cue2openapi") {
		t.Errorf("expected help output to mention 'cue2openapi', got:\n%s", out)
	}
	if !strings.Contains(out, "--schema") {
		t.Errorf("expected help output to document '--schema', got:\n%s", out)
	}
}

func TestCue2OpenAPICmdExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test invoking CUE generation in -short mode")
	}

	tmpDir := t.TempDir()
	outYaml := filepath.Join(tmpDir, "openapi.yaml")
	manifestJSON := filepath.Join(tmpDir, "manifest.json")

	cmd := NewCue2OpenAPICmd()
	cmd.SetArgs([]string{
		"--schema", "../../../..",
		"--output", outYaml,
		"--manifest", manifestJSON,
		"--title", "Gemara CLI Test Title",
		"--version", "v9.9.9",
		"--root", "#EvaluationLog",
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cue2openapi execution failed: %v", err)
	}

	yamlContent, err := os.ReadFile(outYaml)
	if err != nil {
		t.Fatalf("reading generated openapi.yaml: %v", err)
	}
	if !strings.Contains(string(yamlContent), "Gemara CLI Test Title") {
		t.Errorf("expected openapi.yaml to contain title 'Gemara CLI Test Title'")
	}
	if !strings.Contains(string(yamlContent), "v9.9.9") {
		t.Errorf("expected openapi.yaml to contain version 'v9.9.9'")
	}

	manifestContent, err := os.ReadFile(manifestJSON)
	if err != nil {
		t.Fatalf("reading generated manifest.json: %v", err)
	}
	if len(manifestContent) == 0 {
		t.Errorf("expected manifest.json to be non-empty")
	}
}

func TestCue2OpenAPICmdExecutionWithoutManifest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test invoking CUE generation in -short mode")
	}

	tmpDir := t.TempDir()
	outYaml := filepath.Join(tmpDir, "openapi.yaml")

	cmd := NewCue2OpenAPICmd()
	cmd.SetArgs([]string{
		"--schema", "../../../..",
		"--output", outYaml,
		"--manifest", "",
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cue2openapi execution without manifest failed: %v", err)
	}

	if _, err := os.Stat(outYaml); err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
}

func TestCue2OpenAPICmdInvalidSchema(t *testing.T) {
	cmd := NewCue2OpenAPICmd()
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	cmd.SetArgs([]string{
		"--schema", nonExistent,
		"--output", filepath.Join(tmpDir, "out.yaml"),
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when running cue2openapi on nonexistent schema dir, got nil")
	}
}
