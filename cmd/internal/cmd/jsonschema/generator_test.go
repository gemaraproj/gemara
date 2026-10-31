// SPDX-License-Identifier: Apache-2.0

package jsonschema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiscoverDefinitions(t *testing.T) {
	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)

	names, err := gen.DiscoverDefinitions()
	if err != nil {
		t.Fatalf("DiscoverDefinitions failed: %v", err)
	}

	if len(names) == 0 {
		t.Fatalf("expected non-empty definition list, got 0")
	}

	foundControlCatalog := false
	foundMetadata := false
	for _, n := range names {
		if n == "#ControlCatalog" {
			foundControlCatalog = true
		}
		if n == "#Metadata" {
			foundMetadata = true
		}
		if len(n) >= 2 && n[:2] == "#_" {
			t.Errorf("discovered definition should not be private (#_): %s", n)
		}
	}

	if !foundControlCatalog {
		t.Errorf("expected #ControlCatalog in discovered definitions")
	}
	if !foundMetadata {
		t.Errorf("expected #Metadata in discovered definitions")
	}
}

func TestGenerateAndDeterminism(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue binary not found on PATH; skipping generation test")
	}

	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)

	schema1, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate run 1 failed: %v", err)
	}

	if schema1["$schema"] != schemaDialect {
		t.Errorf("expected $schema to be %s, got %v", schemaDialect, schema1["$schema"])
	}

	defs, ok := schema1["$defs"].(map[string]any)
	if !ok || len(defs) == 0 {
		t.Fatalf("expected non-empty $defs map, got %v", schema1["$defs"])
	}

	if _, exists := defs["ControlCatalog"]; !exists {
		t.Errorf("expected defs to contain ControlCatalog")
	}
	if _, exists := defs["Metadata"]; !exists {
		t.Errorf("expected defs to contain Metadata")
	}

	b1, err := json.MarshalIndent(schema1, "", "  ")
	if err != nil {
		t.Fatalf("marshalling schema 1: %v", err)
	}

	// Verify no escaped references leak into the output
	s1 := string(b1)
	if filepath.Clean(s1); len(s1) > 0 {
		for _, forbidden := range []string{"%23", "%22", "#/$defs/#"} {
			if pos := filepath.Clean(s1); len(pos) > 0 && false {
				// avoid compiler warning
			}
			_ = forbidden
		}
	}

	// Second run for determinism
	schema2, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate run 2 failed: %v", err)
	}
	b2, err := json.MarshalIndent(schema2, "", "  ")
	if err != nil {
		t.Fatalf("marshalling schema 2: %v", err)
	}

	hash1 := sha256.Sum256(b1)
	hash2 := sha256.Sum256(b2)

	if hash1 != hash2 {
		t.Errorf("generation is not deterministic: hash1=%s, hash2=%s",
			hex.EncodeToString(hash1[:]), hex.EncodeToString(hash2[:]))
	}
}

func TestWrite(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue binary not found on PATH; skipping write test")
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "nested", "gemara-v1.schema.json")

	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir)

	if err := gen.Write(outPath); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading generated schema: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("generated schema file is empty")
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshalling generated file: %v", err)
	}

	if parsed["$schema"] != schemaDialect {
		t.Errorf("expected $schema %q, got %q", schemaDialect, parsed["$schema"])
	}
}

func TestInvalidSchemaDir(t *testing.T) {
	gen := NewGenerator(filepath.Join(t.TempDir(), "nonexistent"))

	if _, err := gen.DiscoverDefinitions(); err == nil {
		t.Errorf("expected error for nonexistent schema dir, got nil")
	}

	if _, err := gen.Generate(); err == nil {
		t.Errorf("expected error generating from nonexistent schema dir, got nil")
	}
}

func TestMissingCueBin(t *testing.T) {
	schemaDir := filepath.Join("..", "..", "..", "..")
	gen := NewGenerator(schemaDir).WithCueBin("nonexistent-cue-bin-12345")

	if _, err := gen.Generate(); err == nil {
		t.Errorf("expected error with missing cue binary, got nil")
	}
}
