// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"path/filepath"
	"testing"
)

func TestFileStatusesAndSchemaFiles(t *testing.T) {
	v, _, inst, err := loadPrepared("../../..")
	if err != nil {
		t.Fatal(err)
	}
	statuses, err := fileStatuses(inst)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) == 0 {
		t.Fatal("no @status attributes found")
	}
	var metadata string
	for abs, status := range statuses {
		if filepath.Base(abs) == "metadata.cue" {
			metadata = status
		}
	}
	if metadata != "stable" {
		t.Errorf("metadata.cue status = %q, want stable", metadata)
	}

	files := schemaFiles(v)
	if got := files["Datetime"]; got != "metadata.cue" {
		t.Errorf("Datetime declared in %q, want metadata.cue", got)
	}
	if got := files["AuditLog"]; got != "auditlog.cue" {
		t.Errorf("AuditLog declared in %q, want auditlog.cue", got)
	}
}

func TestBuildManifest(t *testing.T) {
	got := buildManifest(map[string]string{
		"AuditLog": "auditlog.cue",
		"Evidence": "auditlog.cue",
		"Datetime": "metadata.cue",
	})
	if want := []string{"AuditLog", "Evidence"}; len(got["auditlog.cue"]) != 2 ||
		got["auditlog.cue"][0] != want[0] || got["auditlog.cue"][1] != want[1] {
		t.Errorf("auditlog.cue = %v, want %v", got["auditlog.cue"], want)
	}
}

// The join between fileStatuses (keyed by inst.Files[].Filename) and
// schemaAbsFiles (keyed by Value.Pos().Filename()) is what makes x-status work
// at all. If those two path strings ever disagree, every x-status silently
// vanishes — so assert against real repository data rather than trusting it.
func TestApplyStatusStampsRealSchemas(t *testing.T) {
	v, _, inst, err := loadPrepared("../../..")
	if err != nil {
		t.Fatal(err)
	}
	statuses, err := fileStatuses(inst)
	if err != nil {
		t.Fatal(err)
	}
	absFiles := schemaAbsFiles(v)

	schemas := map[string]any{}
	for name := range absFiles {
		schemas[name] = map[string]any{"type": "object"}
	}
	doc := map[string]any{"components": map[string]any{"schemas": schemas}}
	applyStatus(doc, absFiles, statuses)

	stamped := 0
	for name, raw := range schemas {
		s := raw.(map[string]any)
		got, ok := s["x-status"].(string)
		if !ok {
			continue
		}
		switch got {
		case "stable", "experimental", "deprecated":
		default:
			t.Errorf("%s has unexpected x-status %q", name, got)
		}
		stamped++
	}
	if stamped == 0 {
		t.Fatal("no schema was stamped with x-status; the fileStatuses/schemaAbsFiles path join is broken")
	}
	// Every definition in a file that declares @status must be stamped.
	for name, abs := range absFiles {
		if statuses[abs] == "" {
			continue
		}
		if _, ok := schemas[name].(map[string]any)["x-status"]; !ok {
			t.Errorf("%s is declared in %s which has a status, but was not stamped", name, abs)
		}
	}
	t.Logf("stamped %d schemas from %d status-carrying files", stamped, len(statuses))
}

// The three defaults the schema actually declares must be recovered from CUE,
// because the encoder cannot emit them: FieldFilter strips `default` wholesale
// to suppress the non-concrete one the open-list form produces.
func TestFieldDefaultsFromCUE(t *testing.T) {
	v, _, _, err := loadPrepared("../../..")
	if err != nil {
		t.Fatal(err)
	}
	defaults, err := fieldDefaults(v)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{
		"Control.state":           "Active",
		"AcceptedMethod.required": false,
	} {
		got, ok := defaults[key]
		if !ok {
			t.Errorf("%s has no recovered default", key)
			continue
		}
		if got != want {
			t.Errorf("%s default = %#v, want %#v", key, got, want)
		}
	}
}
