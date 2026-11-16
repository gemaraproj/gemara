// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestSchemaStatusEdgeCases verifies schemaStatus handling across string, json.RawMessage,
// and invalid/missing extension types.
func TestSchemaStatusEdgeCases(t *testing.T) {
	// Nil ref or nil Value
	if status := schemaStatus(nil); status != "" {
		t.Fatalf("expected empty status for nil ref, got %q", status)
	}
	if status := schemaStatus(&openapi3.SchemaRef{}); status != "" {
		t.Fatalf("expected empty status for nil Value, got %q", status)
	}

	// Schema without extensions
	refNoExt := &openapi3.SchemaRef{
		Value: openapi3.NewSchema(),
	}
	if status := schemaStatus(refNoExt); status != "" {
		t.Fatalf("expected empty status when Extensions is nil, got %q", status)
	}

	// Schema with string x-status
	refStr := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Extensions: map[string]any{
				"x-status": "experimental",
			},
		},
	}
	if status := schemaStatus(refStr); status != "experimental" {
		t.Fatalf("expected 'experimental', got %q", status)
	}

	// Schema with json.RawMessage x-status (valid quoted JSON string)
	refRaw := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Extensions: map[string]any{
				"x-status": json.RawMessage(`"experimental"`),
			},
		},
	}
	if status := schemaStatus(refRaw); status != "experimental" {
		t.Fatalf("expected 'experimental' from json.RawMessage, got %q", status)
	}

	// Schema with json.RawMessage x-status (invalid / non-string JSON)
	refRawInvalid := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Extensions: map[string]any{
				"x-status": json.RawMessage(`{invalid json string}`),
			},
		},
	}
	if status := schemaStatus(refRawInvalid); status != "" {
		t.Fatalf("expected empty string on unmarshal error, got %q", status)
	}

	// Schema with unexpected type for x-status (e.g. integer or boolean)
	refOtherType := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Extensions: map[string]any{
				"x-status": 12345,
			},
		},
	}
	if status := schemaStatus(refOtherType); status != "" {
		t.Fatalf("expected empty string for non-string/raw type, got %q", status)
	}
}

// TestLoadWrappedPathsAndComponents confirms loadWrapped synthesizes Paths and Components
// when given a document that has nil Paths or Components.
func TestLoadWrappedPathsAndComponents(t *testing.T) {
	doc := openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Minimal Doc",
			Version: "1.0.0",
		},
	}
	data, err := doc.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal doc: %v", err)
	}
	tmp := filepath.Join(t.TempDir(), "minimal.json")
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		t.Fatalf("write minimal doc: %v", err)
	}

	wrapped, err := loadWrapped(tmp)
	if err != nil {
		t.Fatalf("loadWrapped on minimal doc failed: %v", err)
	}
	if wrapped.Paths == nil {
		t.Fatalf("expected non-nil Paths after loadWrapped")
	}
	if wrapped.Components == nil {
		t.Fatalf("expected non-nil Components after loadWrapped")
	}
}

// TestLoadWrappedInvalidFile confirms loadWrapped returns an error for invalid paths or corrupt data.
func TestLoadWrappedInvalidFile(t *testing.T) {
	// Missing file
	if _, err := loadWrapped(filepath.Join(t.TempDir(), "nonexistent.yaml")); err == nil {
		t.Fatalf("expected error loading nonexistent file, got nil")
	}

	// Corrupt content
	corruptFile := filepath.Join(t.TempDir(), "corrupt.yaml")
	if err := os.WriteFile(corruptFile, []byte(":\n  invalid: [yaml"), 0644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	if _, err := loadWrapped(corruptFile); err == nil {
		t.Fatalf("expected error loading corrupt file, got nil")
	}
}

// TestFilterAllowedEdgeCases validates nil changes, nil and empty allowlists,
// and case-sensitive matching.
func TestFilterAllowedEdgeCases(t *testing.T) {
	// Empty changes slice
	if got := filterAllowed(nil, map[string]bool{"id schema": true}); len(got) != 0 {
		t.Fatalf("expected empty slice for nil changes, got %+v", got)
	}

	// Empty allowlist map
	changes := []Change{
		{ID: "err-1", Schema: "Control", Text: "text 1"},
		{ID: "err-2", Schema: "Policy", Text: "text 2"},
	}
	if got := filterAllowed(changes, nil); len(got) != 2 {
		t.Fatalf("expected all changes preserved with nil allowlist, got %+v", got)
	}
	if got := filterAllowed(changes, map[string]bool{}); len(got) != 2 {
		t.Fatalf("expected all changes preserved with empty allowlist, got %+v", got)
	}

	// Positive control: exact matching entry successfully filters the target change
	if got := filterAllowed(changes, map[string]bool{"err-1 Control": true}); len(got) != 1 || got[0].ID != "err-2" {
		t.Fatalf("expected exact entry to filter err-1 and leave err-2, got %+v", got)
	}

	// Case sensitivity & boundary check
	// "err-1 Control" should NOT match "ERR-1 Control" or "err-1 control"
	strictAllow := map[string]bool{
		"ERR-1 Control": true,
		"err-1 control": true,
	}
	if got := filterAllowed(changes, strictAllow); len(got) != 2 {
		t.Fatalf("expected exact case-sensitive matching, but changes were filtered: %+v", got)
	}
}
