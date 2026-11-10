// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

func hasID(changes []Change, id string) bool {
	for _, c := range changes {
		if c.ID == id {
			return true
		}
	}
	return false
}

func TestBreakingChanges(t *testing.T) {
	base, err := loadWrapped("testdata/base.yaml")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	cases := []struct {
		name   string
		rev    string
		wantID string // "" means expect zero ERR changes
	}{
		{"identical", "testdata/base.yaml", ""},
		{"removed_required", "testdata/rev_remove.yaml", "response-property-became-optional"},
		{"deleted_required_property", "testdata/rev_delete_required_property.yaml", "response-required-property-removed"},
		{"added_required", "testdata/rev_add.yaml", "new-required-request-property"},
		{"enum_narrowed", "testdata/rev_enum_narrow.yaml", "request-property-enum-value-removed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rev, err := loadWrapped(tc.rev)
			if err != nil {
				t.Fatalf("load rev: %v", err)
			}
			got, err := breakingChanges(base, rev)
			if err != nil {
				t.Fatalf("breakingChanges: %v", err)
			}
			if tc.wantID == "" {
				if len(got) != 0 {
					t.Fatalf("expected no breaking changes, got %d: %+v", len(got), got)
				}
				return
			}
			if !hasID(got, tc.wantID) {
				t.Fatalf("expected change ID %q, got %+v", tc.wantID, got)
			}
		})
	}
}

func TestBreakingChangesReportsRemovedRequestProperty(t *testing.T) {
	base, err := loadWrapped("testdata/base.yaml")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	rev, err := loadWrapped("testdata/rev_delete_required_property.yaml")
	if err != nil {
		t.Fatalf("load rev: %v", err)
	}
	got, err := breakingChanges(base, rev)
	if err != nil {
		t.Fatalf("breakingChanges: %v", err)
	}
	if !hasID(got, checker.RequestPropertyRemovedId) {
		t.Fatalf("expected removed request property, got %+v", got)
	}
}

// TestBreakingChangesExemptsExperimental confirms that a schema carrying
// x-status "experimental" is exempted from the gate: loadWrapped maps it to
// oasdiff's x-stability-level "alpha", which the default (beta) threshold
// filters out before the breaking-change checks run. A break that would fail
// for a stable schema must produce zero ERR changes here.
func TestBreakingChangesExemptsExperimental(t *testing.T) {
	base, err := loadWrapped("testdata/base_experimental.yaml")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	rev, err := loadWrapped("testdata/rev_experimental_break.yaml")
	if err != nil {
		t.Fatalf("load rev: %v", err)
	}
	got, err := breakingChanges(base, rev)
	if err != nil {
		t.Fatalf("breakingChanges: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected experimental schema to be exempt, got %d changes: %+v", len(got), got)
	}
}

// TestBreakingChangesCarriesSchema confirms each reported change carries the
// schema name (the internal "/_schema/" wrapper stripped) so multi-schema
// output is attributable and the allowlist can be keyed per schema.
func TestBreakingChangesCarriesSchema(t *testing.T) {
	base, err := loadWrapped("testdata/base.yaml")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	rev, err := loadWrapped("testdata/rev_remove.yaml")
	if err != nil {
		t.Fatalf("load rev: %v", err)
	}
	got, err := breakingChanges(base, rev)
	if err != nil {
		t.Fatalf("breakingChanges: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected at least one breaking change")
	}
	for _, c := range got {
		if c.Schema != "ControlEvaluation" {
			t.Fatalf("expected schema %q, got %+v", "ControlEvaluation", c)
		}
	}
}

func TestLoadAllowlistOpenErrorPropagates(t *testing.T) {
	// A directory opens but cannot be read as a file; the scan/read (or on some
	// platforms the open) must surface a non-nil error rather than being silenced.
	dir := t.TempDir()
	if _, err := loadAllowlist(dir); err == nil {
		t.Fatalf("expected error for directory path, got nil")
	}

	// A non-existent path must also error rather than returning a nil map silently.
	missing := filepath.Join(dir, "does-not-exist.txt")
	got, err := loadAllowlist(missing)
	if err == nil {
		t.Fatalf("expected error for missing path, got nil")
	}
	if got != nil {
		t.Fatalf("expected nil map on error, got %+v", got)
	}
}

func TestLoadAllowlistEmptyPath(t *testing.T) {
	got, err := loadAllowlist("")
	if err != nil {
		t.Fatalf("loadAllowlist(\"\"): %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil empty map for empty path")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %+v", got)
	}
}

func TestLoadAllowlistParsesEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allow.txt")
	content := "# a comment\n\nfirst-id\n  second-id  \n# another\nthird-id\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing allowlist: %v", err)
	}
	got, err := loadAllowlist(path)
	if err != nil {
		t.Fatalf("loadAllowlist: %v", err)
	}
	for _, id := range []string{"first-id", "second-id", "third-id"} {
		if !got[id] {
			t.Fatalf("expected %q in allowlist, got %+v", id, got)
		}
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %+v", got)
	}
}

func TestFilterAllowed(t *testing.T) {
	changes := []Change{
		{ID: "a", Schema: "Foo", Text: "x"},
		{ID: "b", Schema: "Foo", Text: "y"},
	}
	got, unused := filterAllowed(changes, map[string]bool{"a Foo": true})
	if len(got) != 1 || got[0].ID != "b" {
		t.Fatalf("expected only [b], got %+v", got)
	}
	if len(unused) != 0 {
		t.Fatalf("expected 0 unused entries, got %+v", unused)
	}
}

func TestFilterAllowedSchemaScoped(t *testing.T) {
	changes := []Change{
		{ID: "a", Schema: "Foo", Text: "x"},
		{ID: "a", Schema: "Bar", Text: "y"},
	}
	// A schema-scoped entry drops only the matching change, not every change
	// sharing the check ID.
	got, unused := filterAllowed(changes, map[string]bool{"a Foo": true})
	if len(got) != 1 || got[0].Schema != "Bar" {
		t.Fatalf("expected only the Bar change to remain, got %+v", got)
	}
	if len(unused) != 0 {
		t.Fatalf("expected 0 unused entries, got %+v", unused)
	}
	// A bare check ID (no schema) matches nothing: entries must be schema-scoped.
	got, unused = filterAllowed(changes, map[string]bool{"a": true})
	if len(got) != 2 {
		t.Fatalf("expected a bare ID to drop nothing, got %+v", got)
	}
	if len(unused) != 1 || unused[0] != "a" {
		t.Fatalf("expected [a] unused, got %+v", unused)
	}
}

func TestFilterAllowedUnusedEntries(t *testing.T) {
	changes := []Change{
		{ID: "req-prop-removed", Schema: "User", Text: "property removed"},
	}
	// Matching exception, stale exception, and exception with same check ID on different schema
	allowlist := map[string]bool{
		"req-prop-removed User":   true, // matched
		"req-prop-removed Order":  true, // same check ID, different schema (stale/unused)
		"some-other-check Action": true, // completely unrelated (stale/unused)
	}
	remaining, unused := filterAllowed(changes, allowlist)
	if len(remaining) != 0 {
		t.Fatalf("expected remaining to be empty, got %+v", remaining)
	}
	if len(unused) != 2 {
		t.Fatalf("expected 2 unused entries, got %d: %+v", len(unused), unused)
	}
	unusedMap := make(map[string]bool)
	for _, u := range unused {
		unusedMap[u] = true
	}
	if !unusedMap["req-prop-removed Order"] || !unusedMap["some-other-check Action"] {
		t.Fatalf("expected specific unused entries, got %+v", unused)
	}
}
