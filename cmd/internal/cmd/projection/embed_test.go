// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// convertSource runs the full pipeline over src as the only file of a
// throwaway CUE module, for behaviour that depends on how the encoder lays
// out definitions rather than on the repository's own schemas.
func convertSource(t *testing.T, src string) (*openapi3.T, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "cue.mod"), 0o755); err != nil {
		t.Fatal(err)
	}
	module := "module: \"example.com/p@v0\"\nlanguage: version: \"v0.15.1\"\n"
	if err := os.WriteFile(filepath.Join(dir, "cue.mod", "module.cue"), []byte(module), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.cue"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	c := NewConverter("", "")
	if err := c.Load(dir); err != nil {
		t.Fatal(err)
	}
	if err := c.Prep(); err != nil {
		return nil, err
	}
	if err := c.Convert(dir, WithVersion("test")); err != nil {
		return nil, err
	}
	if err := c.Post(); err != nil {
		return nil, err
	}
	return c.doc, nil
}

// A directive on a field of an embedded definition is seen by CUE on every
// definition that embeds it, but the encoder emits the field only on the
// embedded schema. It must land there instead of failing generation.
func TestFieldDirectivesOnEmbeddedDefinition(t *testing.T) {
	doc, err := convertSource(t, `@gemara(status="stable")
package p

#Base: {
	enabled: *false | bool @gemara(default=false)
}

#Outer: {
	#Base
	name: string
}
`)
	if err != nil {
		t.Fatal(err)
	}
	base := doc.Components.Schemas["Base"].Value
	if got := base.Properties["enabled"].Value.Default; got != false {
		t.Errorf("Base.enabled default = %v, want false", got)
	}
	if slicesContainString(base.Required, "enabled") {
		t.Errorf("Base.enabled still required despite its default: %v", base.Required)
	}
}
