// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func writeGenerated(t *testing.T, outputPath, manifestPath, root string, options ...ConvertOption) {
	t.Helper()
	c := NewConverter(manifestPath, root)
	if err := c.Load("../../../.."); err != nil {
		t.Fatal(err)
	}
	if err := c.Prep(); err != nil {
		t.Fatal(err)
	}
	if err := c.Convert("../../../..", options...); err != nil {
		t.Fatal(err)
	}
	if err := c.Post(); err != nil {
		t.Fatal(err)
	}
	if err := c.WriteOpenAPI(outputPath); err != nil {
		t.Fatal(err)
	}
	if manifestPath != "" {
		if err := c.WriteManifest(manifestPath); err != nil {
			t.Fatal(err)
		}
	}
}

// loadGenerated runs the converter over the repo's CUE package and returns the
// decoded document's schemas.
func loadGenerated(t *testing.T) (map[string]any, []byte) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "projection.yaml")
	writeGenerated(t, out, "", "")
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		OpenAPI    string `yaml:"openapi"`
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	if spec.OpenAPI != "3.0.3" {
		t.Errorf("projection = %q, want 3.0.3", spec.OpenAPI)
	}
	return spec.Components.Schemas, data
}

func enumOf(t *testing.T, schemas map[string]any, name string) []string {
	t.Helper()
	s, _ := schemas[name].(map[string]any)
	raw, _ := s["enum"].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, v.(string))
	}
	return out
}

// requireRef asserts a hidden helper reference was collapsed to its named base.
func requireRef(t *testing.T, items map[string]any, wantRef, label string) {
	t.Helper()
	if ref, _ := items["$ref"].(string); ref != wantRef {
		t.Errorf("%s items.$ref = %q, want %q", label, ref, wantRef)
	}
}

func TestConverter(t *testing.T) {
	schemas, data := loadGenerated(t)

	// Disjunctions of string literals emit enum values. Expected values come
	// from #MethodType (policy.cue) and #Lifecycle (collections.cue).
	if got := enumOf(t, schemas, "MethodType"); !slices.Equal(got, []string{"Behavioral", "Intent", "Remediation", "Gate"}) {
		t.Errorf("MethodType enum = %v", got)
	}
	if got := enumOf(t, schemas, "Lifecycle"); !slices.Equal(got, []string{"Active", "Draft", "Deprecated", "Retired"}) {
		t.Errorf("Lifecycle enum = %v", got)
	}

	// Hidden definitions and hidden fields don't leak into the output.
	for name, raw := range schemas {
		if strings.HasPrefix(name, "_") {
			t.Errorf("hidden definition %s leaked into schemas", name)
		}
		s, _ := raw.(map[string]any)
		for prop := range mapOf(s["properties"]) {
			if strings.HasPrefix(prop, "_") {
				t.Errorf("hidden field %s.%s leaked into properties", name, prop)
			}
		}
	}
	if strings.Contains(string(data), "schemas/_") {
		t.Errorf("ref to a hidden definition leaked into output")
	}

	// Refs to a hidden definition resolve to their visible base type
	// (#_MappingStrict -> Mapping, mappingdocument.cue), not merely to
	// something non-hidden. The helper's rule is projected onto the base;
	// helperrules_test.go covers that.
	mappings, ok := mapOf(mapOf(schemas["MappingDocument"])["properties"])["mappings"].(map[string]any)
	if !ok {
		t.Fatalf("MappingDocument.mappings missing: %v", schemas["MappingDocument"])
	}
	requireRef(t, mapOf(mappings["items"]), "#/components/schemas/Mapping", "MappingDocument.mappings")

	// The at-least-one Evidence rule is projected onto Evidence;
	// helperrules_test.go covers that. Here the hidden helper must not leak.
	evidence, ok := mapOf(mapOf(schemas["AssessmentLog"])["properties"])["evidence"].(map[string]any)
	if !ok {
		t.Fatalf("AssessmentLog.evidence missing: %v", schemas["AssessmentLog"])
	}
	if strings.Contains(fmt.Sprint(evidence), "schemas/_") {
		t.Errorf("AssessmentLog.evidence retains a reference to a hidden helper")
	}

	// x-status survives, carried from each file's @gemara(status=...) attribute.
	md, _ := schemas["Metadata"].(map[string]any)
	if md["x-status"] != "stable" {
		t.Errorf("Metadata x-status = %v, want stable", md["x-status"])
	}
	al, _ := schemas["AuditLog"].(map[string]any)
	if al["x-status"] != "experimental" {
		t.Errorf("AuditLog x-status = %v, want experimental", al["x-status"])
	}
}

// The document has always carried an info.description, and --root replaces it
// with the named definition's doc comment.
func TestConverterInfoDescription(t *testing.T) {
	read := func(t *testing.T, root string) map[string]any {
		t.Helper()
		out := filepath.Join(t.TempDir(), "projection.yaml")
		writeGenerated(t, out, "", root)
		data, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		var spec struct {
			Info map[string]any `yaml:"info"`
		}
		if err := yaml.Unmarshal(data, &spec); err != nil {
			t.Fatal(err)
		}
		return spec.Info
	}

	if got := read(t, "")["description"]; got != "Gemara schema definitions" {
		t.Errorf("default info.description = %v, want %q", got, "Gemara schema definitions")
	}

	got, _ := read(t, "#Metadata")["description"].(string)
	if got == "" || got == "Gemara schema definitions" {
		t.Errorf("--root did not override info.description, got %q", got)
	}
	if !strings.Contains(got, "Metadata") {
		t.Errorf("info.description from #Metadata = %q, want the definition's doc comment", got)
	}
}

func TestConverterOptions(t *testing.T) {
	out := filepath.Join(t.TempDir(), "projection.yaml")
	writeGenerated(t, out, "", "", WithTitle("Test projection"), WithVersion("test-version"))

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Info struct {
			Title   string `yaml:"title"`
			Version string `yaml:"version"`
		} `yaml:"info"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Info.Title != "Test projection" {
		t.Errorf("info.title = %q, want %q", spec.Info.Title, "Test projection")
	}
	if spec.Info.Version != "test-version" {
		t.Errorf("info.version = %q, want %q", spec.Info.Version, "test-version")
	}
}

// A --root that names no definition used to yield an error cue.Value whose
// Doc() is nil, leaving the default description in place: a typo produced a
// plausible, wrong document instead of a complaint.
func TestConverterRejectsUnknownRoot(t *testing.T) {
	c := NewConverter("", "#Metdata")
	if err := c.Load("../../../.."); err != nil {
		t.Fatal(err)
	}
	if err := c.Prep(); err != nil {
		t.Fatal(err)
	}
	err := c.Convert("../../../..")
	if err == nil {
		t.Fatal("expected an error for a --root naming no definition, got nil")
	}
	if !strings.Contains(err.Error(), "#Metdata") {
		t.Errorf("error %q does not name the offending root path", err)
	}
}

func TestConverterWritesManifest(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "schema-manifest.json")
	writeGenerated(t, filepath.Join(dir, "projection.yaml"), manifest, "")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string][]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got["metadata.cue"], "Datetime") {
		t.Errorf("manifest metadata.cue = %v, want it to contain Datetime", got["metadata.cue"])
	}
}
