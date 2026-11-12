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

func TestCue2OpenAPICmdRejectsOverwritingInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test invoking CUE generation in -short mode")
	}

	tests := []struct {
		name       string
		setArgs    func(schemaDir, outDefault, manifestDefault string) []string
		targetFile func(schemaDir string) string
	}{
		{
			name: "output targets CUE source file",
			setArgs: func(schemaDir, outDefault, manifestDefault string) []string {
				return []string{
					"--schema", schemaDir,
					"--output", filepath.Join(schemaDir, "policy.cue"),
					"--manifest", manifestDefault,
				}
			},
			targetFile: func(schemaDir string) string {
				return filepath.Join(schemaDir, "policy.cue")
			},
		},
		{
			name: "manifest targets VERSION file",
			setArgs: func(schemaDir, outDefault, manifestDefault string) []string {
				return []string{
					"--schema", schemaDir,
					"--output", outDefault,
					"--manifest", filepath.Join(schemaDir, "VERSION"),
				}
			},
			targetFile: func(schemaDir string) string {
				return filepath.Join(schemaDir, "VERSION")
			},
		},
		{
			name: "output targets VERSION file via relative path traversal",
			setArgs: func(schemaDir, outDefault, manifestDefault string) []string {
				return []string{
					"--schema", schemaDir,
					"--output", filepath.Join(schemaDir, "sub", "..", "VERSION"),
					"--manifest", manifestDefault,
				}
			},
			targetFile: func(schemaDir string) string {
				return filepath.Join(schemaDir, "VERSION")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create a minimal CUE package in tmpDir with a VERSION file
			cueModDir := filepath.Join(tmpDir, "cue.mod")
			if err := os.MkdirAll(cueModDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(cueModDir, "module.cue"), []byte("module: \"test.example\"\nlanguage: version: \"v0.12.0\"\n"), 0644); err != nil {
				t.Fatal(err)
			}

			cueContent := []byte("@gemara(status=\"stable\")\npackage gemara\n\n#TestDef: {\n\tfield: string\n}\n")
			cuePath := filepath.Join(tmpDir, "policy.cue")
			if err := os.WriteFile(cuePath, cueContent, 0644); err != nil {
				t.Fatal(err)
			}

			versionContent := []byte("v1.2.3\n")
			versionPath := filepath.Join(tmpDir, "VERSION")
			if err := os.WriteFile(versionPath, versionContent, 0644); err != nil {
				t.Fatal(err)
			}

			outDefault := filepath.Join(tmpDir, "out.yaml")
			manifestDefault := filepath.Join(tmpDir, "manifest.json")

			target := tc.targetFile(tmpDir)
			originalContent, err := os.ReadFile(target)
			if err != nil {
				t.Fatalf("reading original target file: %v", err)
			}

			cmd := NewCue2OpenAPICmd()
			cmd.SetArgs(tc.setArgs(tmpDir, outDefault, manifestDefault))
			t.Cleanup(func() {
				cmd.SetArgs(nil)
			})

			err = cmd.Execute()
			if err == nil {
				t.Fatalf("expected error when output or manifest overwrites input, got nil")
			}
			if !strings.Contains(err.Error(), "would overwrite input file") {
				t.Fatalf("expected 'would overwrite input file' in error, got: %v", err)
			}

			// Verify input file was not altered or truncated
			afterContent, err := os.ReadFile(target)
			if err != nil {
				t.Fatalf("reading target file after execution: %v", err)
			}
			if string(afterContent) != string(originalContent) {
				t.Errorf("target file content was modified!\nGot: %q\nWant: %q", string(afterContent), string(originalContent))
			}
		})
	}
}
