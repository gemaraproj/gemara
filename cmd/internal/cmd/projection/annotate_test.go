// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"
)

// mustValue parses and builds src into a cue.Value for tests that need one
// without loading the whole repo schema.
func mustValue(t *testing.T, src string) cue.Value {
	t.Helper()
	f, err := parser.ParseFile("test.cue", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	v := cuecontext.New().BuildFile(f)
	if err := v.Err(); err != nil {
		t.Fatalf("build: %v", err)
	}
	return v
}

func TestFileStatusesAndSchemaFiles(t *testing.T) {
	v, _, inst, err := loadPrepared("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	statuses, err := fileStatuses(inst)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) == 0 {
		t.Fatal("no @gemara(status=...) attributes found")
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

	files, err := schemaFiles(v)
	if err != nil {
		t.Fatal(err)
	}
	if got := files["Datetime"]; got != "metadata.cue" {
		t.Errorf("Datetime declared in %q, want metadata.cue", got)
	}
	if got := files["AuditLog"]; got != "auditlog.cue" {
		t.Errorf("AuditLog declared in %q, want auditlog.cue", got)
	}
}

// instanceOf parses each source as a file of one build instance, for tests of
// file-level attributes that do not need the whole repo schema.
func instanceOf(t *testing.T, files map[string]string) *build.Instance {
	t.Helper()
	inst := &build.Instance{}
	for name, src := range files {
		f, err := parser.ParseFile(name, src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		inst.Files = append(inst.Files, f)
	}
	return inst
}

// A file still carrying the pre-@gemara @status attribute must not be
// skipped: its definitions would silently lose x-status and fall under
// breaking-check as if they were stable.
func TestFileStatusesRejectsLegacyStatusAttribute(t *testing.T) {
	inst := instanceOf(t, map[string]string{
		"legacy.cue": "@status(\"experimental\")\npackage p\n",
	})
	_, err := fileStatuses(inst)
	if err == nil || !strings.Contains(err.Error(), "legacy.cue") || !strings.Contains(err.Error(), "@status") {
		t.Fatalf("err = %v, want an error naming legacy.cue and @status", err)
	}
}

func TestFileStatusesRejectsFileWithoutStatus(t *testing.T) {
	inst := instanceOf(t, map[string]string{
		"a.cue": "@gemara(status=\"stable\")\npackage p\n",
		"b.cue": "package p\n",
	})
	_, err := fileStatuses(inst)
	if err == nil || !strings.Contains(err.Error(), "b.cue") {
		t.Fatalf("err = %v, want an error naming b.cue", err)
	}
}

func TestFileStatusesReadsGemaraStatus(t *testing.T) {
	inst := instanceOf(t, map[string]string{
		"a.cue": "@gemara(status=\"deprecated\")\npackage p\n",
	})
	got, err := fileStatuses(inst)
	if err != nil {
		t.Fatal(err)
	}
	if got["a.cue"] != "deprecated" {
		t.Errorf("a.cue status = %q, want deprecated", got["a.cue"])
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
	v, _, inst, err := loadPrepared("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	statuses, err := fileStatuses(inst)
	if err != nil {
		t.Fatal(err)
	}
	absFiles, err := schemaAbsFiles(v)
	if err != nil {
		t.Fatal(err)
	}

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
	// Every definition in a file that declares @gemara(status=...) must be stamped.
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

// A bare @deprecated() marks a field deprecated with no reason; a value maps
// the field to its reason text.
func TestFieldDeprecations(t *testing.T) {
	v := mustValue(t, `package p
#Widget: {
	current: string
	legacy?: string @deprecated()
	oldName?: string @deprecated("use current instead")
}
`)
	got, err := fieldDeprecations(v)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["Widget.current"]; ok {
		t.Error("current is not deprecated, but was recorded")
	}
	reason, ok := got["Widget.legacy"]
	if !ok {
		t.Fatal("legacy: expected a deprecation entry")
	}
	if reason != "" {
		t.Errorf("legacy reason = %q, want empty (bare @deprecated())", reason)
	}
	reason, ok = got["Widget.oldName"]
	if !ok {
		t.Fatal("oldName: expected a deprecation entry")
	}
	if reason != "use current instead" {
		t.Errorf("oldName reason = %q, want %q", reason, "use current instead")
	}
}

// Explicit @gemara(default=...) directives carry the defaults the encoder
// cannot emit because FieldFilter strips `default` wholesale.
func TestFieldDefaultsFromCUE(t *testing.T) {
	v, _, _, err := loadPrepared("../../../..")
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
		"Recommendation.required": false,
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

func TestDefinitionFormatsFromGemaraAttributes(t *testing.T) {
	v := mustValue(t, `package p
#Datetime: string @gemara(format="date-time")
#Plain: string
`)
	formats, err := definitionFormats(v)
	if err != nil {
		t.Fatal(err)
	}
	if got := formats["Datetime"]; got != "date-time" {
		t.Errorf("Datetime format = %q, want date-time", got)
	}
	if _, ok := formats["Plain"]; ok {
		t.Error("Plain has no @gemara(format=...), but a format was recorded")
	}
}

func TestUnprojectableFieldsFromGemaraAttributes(t *testing.T) {
	v := mustValue(t, `package p
#Widget: {
	complete: string @gemara(projectable=true)
	partial: string @gemara(projectable=false)
}
`)
	got, err := unprojectableFields(v)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["Widget.complete"]; ok {
		t.Error("Widget.complete is projectable, but was recorded as unprojectable")
	}
	if _, ok := got["Widget.partial"]; !ok {
		t.Error("Widget.partial has projectable=false, but was not recorded")
	}
}

func TestUnprojectableFieldsRejectInvalidProjectableValue(t *testing.T) {
	v := mustValue(t, `package p
#Widget: { partial: string @gemara(projectable="sometimes") }
`)
	if _, err := unprojectableFields(v); err == nil {
		t.Error("expected projectable=sometimes to be rejected")
	}
}
