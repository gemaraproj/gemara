// SPDX-License-Identifier: Apache-2.0

package jsonschema

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCue2JSONSchemaCmdStructure(t *testing.T) {
	cmd := NewCue2JSONSchemaCmd()

	if cmd.Use != "cue2jsonschema" {
		t.Fatalf("expected Use to be 'cue2jsonschema', got %q", cmd.Use)
	}

	flags := []string{"schema", "output", "cue-bin"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("expected flag %q on cue2jsonschema command", f)
		}
	}
}

func TestCue2JSONSchemaCmdHelp(t *testing.T) {
	cmd := NewCue2JSONSchemaCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error running help: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "cue2jsonschema") {
		t.Errorf("expected help output to mention 'cue2jsonschema', got:\n%s", out)
	}
	if !strings.Contains(out, "--schema") {
		t.Errorf("expected help output to mention '--schema', got:\n%s", out)
	}
	if !strings.Contains(out, "--output") {
		t.Errorf("expected help output to mention '--output', got:\n%s", out)
	}
}

func TestCue2JSONSchemaCmdExecution(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue binary not found on PATH; skipping execution test")
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "schema.json")
	schemaDir := filepath.Join("..", "..", "..", "..")

	cmd := NewCue2JSONSchemaCmd()
	cmd.SetArgs([]string{
		"--schema", schemaDir,
		"--output", outPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing cue2jsonschema: %v", err)
	}
}

func TestCue2JSONSchemaCmdExecutionError(t *testing.T) {
	cmd := NewCue2JSONSchemaCmd()
	cmd.SetArgs([]string{
		"--schema", filepath.Join(t.TempDir(), "nonexistent"),
		"--output", filepath.Join(t.TempDir(), "out.json"),
	})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error executing with invalid schema dir, got nil")
	}
}
