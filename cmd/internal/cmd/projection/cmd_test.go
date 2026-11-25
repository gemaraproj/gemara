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

func TestCue2OpenAPICmdCollidingOutputAndManifestPaths(t *testing.T) {
	tests := []struct {
		name     string
		output   func(dir string) string
		manifest func(dir string) string
	}{
		{
			name:     "identical paths",
			output:   func(dir string) string { return filepath.Join(dir, "out.json") },
			manifest: func(dir string) string { return filepath.Join(dir, "out.json") },
		},
		{
			name:     "equivalent relative and cleaned paths",
			output:   func(dir string) string { return filepath.Join(dir, "sub", "..", "out.json") },
			manifest: func(dir string) string { return filepath.Join(dir, "out.json") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cmd := NewCue2OpenAPICmd()
			cmd.SetArgs([]string{
				"--schema", "../../../..",
				"--output", tc.output(tmpDir),
				"--manifest", tc.manifest(tmpDir),
			})
			t.Cleanup(func() {
				cmd.SetArgs(nil)
			})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error when output and manifest paths collide, got nil")
			}
			if !strings.Contains(err.Error(), "must not refer to the same file") {
				t.Fatalf("expected colliding path error message, got: %v", err)
			}
		})
	}
}

func TestCue2OpenAPICmdPreservesOutputsOnManifestEncodingFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test invoking CUE generation in -short mode")
	}

	tmpDir := t.TempDir()
	schemaDir := filepath.Join(tmpDir, "schema")
	cueModDir := filepath.Join(schemaDir, "cue.mod")
	if err := os.MkdirAll(cueModDir, 0755); err != nil {
		t.Fatalf("failed to create cue module directory: %v", err)
	}

	moduleFile := filepath.Join(cueModDir, "module.cue")
	moduleContent := []byte("module: \"example.com/p@v0\"\nlanguage: version: \"v0.15.1\"\n")
	if err := os.WriteFile(moduleFile, moduleContent, 0644); err != nil {
		t.Fatalf("failed to write module.cue: %v", err)
	}

	// Schema package with a definition ending in _value and valid file status.
	// Upstream CUE OpenAPI skips definitions ending with _value as internal helper schemas,
	// so it is omitted from doc.Components.Schemas.
	// Consequently, Convert and EncodeOpenAPI succeed, but EncodeManifest fails
	// because definition #Foo_value is declared in CUE but missing from the generated document.
	schemaFile := filepath.Join(schemaDir, "schema.cue")
	schemaContent := []byte("@gemara(status=\"stable\")\npackage p\n\n#Foo_value: string\n")
	if err := os.WriteFile(schemaFile, schemaContent, 0644); err != nil {
		t.Fatalf("failed to write schema.cue: %v", err)
	}

	outYaml := filepath.Join(tmpDir, "openapi.yaml")
	manifestJSON := filepath.Join(tmpDir, "manifest.json")

	initialOpenAPI := []byte("initial openapi content\n")
	initialManifest := []byte("initial manifest content\n")

	if err := os.WriteFile(outYaml, initialOpenAPI, 0644); err != nil {
		t.Fatalf("failed to seed openapi output: %v", err)
	}
	if err := os.WriteFile(manifestJSON, initialManifest, 0644); err != nil {
		t.Fatalf("failed to seed manifest output: %v", err)
	}

	cmd := NewCue2OpenAPICmd()
	cmd.SetArgs([]string{
		"--schema", schemaDir,
		"--output", outYaml,
		"--manifest", manifestJSON,
	})
	t.Cleanup(func() {
		cmd.SetArgs(nil)
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected command to fail when manifest encoding fails, got nil")
	}
	if !strings.Contains(err.Error(), "refusing to write a manifest that omits it") {
		t.Fatalf("expected manifest failure error, got: %v", err)
	}

	gotOpenAPI, readErr := os.ReadFile(outYaml)
	if readErr != nil {
		t.Fatalf("failed to read openapi output: %v", readErr)
	}
	if !bytes.Equal(gotOpenAPI, initialOpenAPI) {
		t.Errorf("expected openapi output to remain unchanged, got %q", string(gotOpenAPI))
	}

	gotManifest, readErr := os.ReadFile(manifestJSON)
	if readErr != nil {
		t.Fatalf("failed to read manifest output: %v", readErr)
	}
	if !bytes.Equal(gotManifest, initialManifest) {
		t.Errorf("expected manifest output to remain unchanged, got %q", string(gotManifest))
	}
}
