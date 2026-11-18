// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"os"
)

func TestNewSchemaWalker(t *testing.T) {
	schemaDir := FindSchemaDir("../..")
	if schemaDir == "" {
		t.Fatalf("could not find schema directory")
	}

	walker, err := NewSchemaWalker(schemaDir)
	if err != nil {
		t.Fatalf("NewSchemaWalker failed: %v", err)
	}

	totalPaths := 0
	for _, art := range defaultArtifacts {
		paths := walker.RefPathsByArtifact[art]
		t.Logf("Artifact #%s has %d reference path(s): %v", art, len(paths), paths)
		totalPaths += len(paths)
	}

	t.Logf("Total reference paths across all %d artifacts: %d", len(defaultArtifacts), totalPaths)
	if totalPaths != 62 {
		t.Errorf("expected 62 reference paths discovered by walker, got %d", totalPaths)
	}
}

func TestSchemaWalker_ExtractReferences(t *testing.T) {
	schemaDir := FindSchemaDir("../..")
	if schemaDir == "" {
		t.Fatalf("could not find schema directory")
	}

	walker, err := NewSchemaWalker(schemaDir)
	if err != nil {
		t.Fatalf("NewSchemaWalker failed: %v", err)
	}

	// Test 1: ControlCatalog
	ctrlPath := filepath.Join(schemaDir, "test/test-data/good-ccc.yaml")
	data, err := os.ReadFile(ctrlPath)
	if err != nil {
		t.Fatalf("read file %s: %v", ctrlPath, err)
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatalf("unmarshal yaml: %v", err)
	}

	doc := &Document{
		FilePath:          ctrlPath,
		Type:              "ControlCatalog",
		MappingReferences: make(map[string]MappingReference),
		DeclaredEntities:  make(map[string]bool),
		SubEntities:       make(map[string]map[string]bool),
	}

	walker.ExtractReferences(doc, root)

	if len(doc.ReferenceUsages) == 0 {
		t.Errorf("expected ReferenceUsages in good-control-catalog, got 0")
	}
	if len(doc.EntityUsages) == 0 {
		t.Errorf("expected EntityUsages in good-control-catalog, got 0")
	}

	t.Logf("Discovered %d ReferenceUsages and %d EntityUsages in good-control-catalog.yaml",
		len(doc.ReferenceUsages), len(doc.EntityUsages))

	// Verify specific references
	foundNIST := false
	for _, ref := range doc.ReferenceUsages {
		if ref.ReferenceID == "NIST-800-53" {
			foundNIST = true
			break
		}
	}
	if !foundNIST {
		t.Errorf("expected reference-id 'NIST-800-53' to be extracted")
	}

	// Test 2: Policy with Retention (Issue #513)
	policyPath := filepath.Join(schemaDir, "test/test-data/good-policy-retention.yaml")
	pData, err := os.ReadFile(policyPath)
	if err == nil {
		var pRoot map[string]any
		if err := yaml.Unmarshal(pData, &pRoot); err == nil {
			pDoc := &Document{
				FilePath:          policyPath,
				Type:              "Policy",
				MappingReferences: make(map[string]MappingReference),
				DeclaredEntities:  make(map[string]bool),
				SubEntities:       make(map[string]map[string]bool),
			}
			walker.ExtractReferences(pDoc, pRoot)

			t.Logf("Discovered %d ReferenceUsages and %d EntityUsages in good-policy-retention.yaml",
				len(pDoc.ReferenceUsages), len(pDoc.EntityUsages))
			foundRetention := false
			for _, ref := range pDoc.ReferenceUsages {
				if ref.ReferenceID == "gdpr" {
					foundRetention = true
					break
				}
			}
			if !foundRetention {
				t.Errorf("expected retention reference 'gdpr' to be dynamically discovered via AST walk")
			}
		}
	}
}
