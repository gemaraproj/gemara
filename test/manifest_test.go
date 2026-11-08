// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	cuejson "cuelang.org/go/encoding/json"
	cueyaml "cuelang.org/go/encoding/yaml"
)

// manifestPath is the fixture manifest, relative to this directory. Paths
// inside it are relative to the repository root.
const manifestPath = "test-data/fixtures.yaml"

// manifestSchema is the shape of fixtures.yaml. It is closed, so a misspelt
// key fails here instead of being ignored.
const manifestSchema = `
#Fixture: close({
	files: [string, ...string]
	definition: =~"^#"
	schema: close({valid: true}) | close({valid: false, error: string & !=""})
	reason: string & !=""
})
fixtures: [...#Fixture]
`

type fixture struct {
	Files      []string `json:"files"`
	Definition string   `json:"definition"`
	Schema     struct {
		Valid bool   `json:"valid"`
		Error string `json:"error"`
	} `json:"schema"`
	Reason string `json:"reason"`
}

func loadManifest(t *testing.T) []fixture {
	t.Helper()
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	file, err := cueyaml.Extract(manifestPath, data)
	if err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	shape := schemaCtx.CompileString(manifestSchema)
	v := shape.Unify(schemaCtx.BuildFile(file))
	if err := v.Validate(cue.Concrete(true)); err != nil {
		t.Fatalf("manifest does not match its schema: %v", cueerrors.Details(err, nil))
	}
	var m struct {
		Fixtures []fixture `json:"fixtures"`
	}
	if err := v.Decode(&m); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	return m.Fixtures
}

func TestFixtureManifest(t *testing.T) {
	for _, f := range loadManifest(t) {
		if len(f.Files) != 1 {
			t.Errorf("%v: schema outcomes are per file; list one file per row", f.Files)
			continue
		}
		file := f.Files[0]
		t.Run(file+"/"+f.Definition, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", file))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			def := schemaValue.LookupPath(cue.ParsePath(f.Definition))
			if def.Err() != nil {
				t.Fatalf("lookup %s: %v", f.Definition, def.Err())
			}
			var verr error
			switch filepath.Ext(file) {
			case ".json":
				verr = cuejson.Validate(data, def)
			case ".yaml", ".yml":
				verr = cueyaml.Validate(data, def)
			default:
				t.Fatalf("unsupported extension")
			}
			switch {
			case f.Schema.Valid && verr != nil:
				t.Errorf("listed as valid (%s), but: %v", f.Reason, verr)
			case !f.Schema.Valid && verr == nil:
				t.Errorf("listed as invalid (%s), but it validates", f.Reason)
			case !f.Schema.Valid:
				// Match any reported error, not just the first that
				// verr.Error() shows. e.Error() has no file positions, so
				// the match does not depend on where the repo is checked out.
				for _, e := range cueerrors.Errors(verr) {
					if strings.Contains(e.Error(), f.Schema.Error) {
						return
					}
				}
				t.Errorf("no error contains %q; got:\n%s", f.Schema.Error, cueerrors.Details(verr, nil))
			}
		})
	}
}

// TestFixtureManifestListsEveryFile fails when a file under test-data has no
// row, so a fixture cannot be added without saying what it is for.
func TestFixtureManifestListsEveryFile(t *testing.T) {
	fixtures := loadManifest(t)
	listed := map[string]bool{}
	var firsts []string
	for _, f := range fixtures {
		firsts = append(firsts, f.Files[0])
		for _, file := range f.Files {
			listed[file] = true
			if _, err := os.Stat(filepath.Join("..", file)); err != nil {
				t.Errorf("listed file %s: %v", file, err)
			}
		}
	}
	err := filepath.WalkDir("test-data", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == filepath.FromSlash(manifestPath) {
			return err
		}
		rel := "test/" + filepath.ToSlash(path)
		if !listed[rel] {
			t.Errorf("%s has no row in %s; add one, for example:\n%s", rel, manifestPath, rowTemplate(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Sorted rows put concurrent additions in different places in the file.
	if !slices.IsSorted(firsts) {
		t.Errorf("rows in %s are not sorted by their first file", manifestPath)
	}
}

// rowTemplate drafts a manifest row for an unlisted file, ready to paste and
// finish. The definition comes from the file's metadata.type and the outcome
// from its bad- prefix; both are guesses for the author to check.
func rowTemplate(rel string) string {
	def := "#?"
	if data, err := os.ReadFile(filepath.Join("..", rel)); err == nil {
		// YAML is a superset of JSON, so this reads both kinds of fixture.
		if file, err := cueyaml.Extract(rel, data); err == nil {
			typ := schemaCtx.BuildFile(file).LookupPath(cue.ParsePath("metadata.type"))
			if s, err := typ.String(); err == nil && s != "" {
				def = "#" + s
			}
		}
	}
	schema := "      valid: true\n"
	if strings.HasPrefix(filepath.Base(rel), "bad") {
		schema = "      valid: false\n      error: '<part of the CUE error message>'\n"
	}
	return "  - files: [" + rel + "]\n" +
		"    definition: '" + def + "'\n" +
		"    schema:\n" + schema +
		"    reason: '<what this fixture is for>'\n"
}
