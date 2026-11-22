// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stabilityError runs the check over files parsed as they are written.
func stabilityError(t *testing.T, files map[string]string) error {
	t.Helper()
	inst := instanceOf(t, files)
	statuses, err := fileStatuses(inst)
	if err != nil {
		t.Fatal(err)
	}
	return checkStableReferences(definitionUses(inst.Files), statuses)
}

func TestCheckStableReferences(t *testing.T) {
	const stable, experimental, deprecated = "@gemara(status=\"stable\")\npackage p\n", "@gemara(status=\"experimental\")\npackage p\n", "@gemara(status=\"deprecated\")\npackage p\n"
	tests := []struct {
		name  string
		files map[string]string
		want  []string // substrings of the error; nil when the files pass
	}{
		{
			name: "stable references stable",
			files: map[string]string{
				"a.cue": stable + "#A: b: #B\n",
				"b.cue": stable + "#B: string\n",
			},
		},
		{
			name: "experimental references experimental and stable",
			files: map[string]string{
				"a.cue": stable + "#A: string\n",
				"b.cue": experimental + "#B: {a: #A, c: #C}\n#C: string\n",
			},
		},
		{
			name: "a name in a comment is not a reference",
			files: map[string]string{
				"a.cue": stable + "// #B is experimental, so #A does not use it.\n#A: string\n",
				"b.cue": experimental + "#B: string\n",
			},
		},
		{
			name: "stable references experimental",
			files: map[string]string{
				"a.cue": stable + "#A: b: #B\n",
				"b.cue": experimental + "#B: string\n",
			},
			want: []string{"a.cue:3:8: stable a.cue names #B, which experimental b.cue declares"},
		},
		{
			name: "stable references deprecated",
			files: map[string]string{
				"a.cue": stable + "#A: b: #B\n",
				"b.cue": deprecated + "#B: string\n",
			},
			want: []string{"stable a.cue names #B, which deprecated b.cue declares"},
		},
		{
			// CUE merges both declarations into one #A, so the experimental
			// field is part of the stable schema.
			name: "experimental file adds to a stable definition",
			files: map[string]string{
				"a.cue": stable + "#A: b: string\n",
				"b.cue": experimental + "#A: c?: string\n",
			},
			want: []string{"a.cue:3:1: stable a.cue names #A, which experimental b.cue declares"},
		},
		{
			name: "reference inside a cross-field rule",
			files: map[string]string{
				"a.cue": stable + "#A: {\n\tb?: string\n\tif b != _|_ {\n\t\tc: #B\n\t}\n}\n",
				"b.cue": experimental + "#B: string\n",
			},
			want: []string{"a.cue:6:6: stable a.cue names #B"},
		},
		{
			name: "hidden definitions count",
			files: map[string]string{
				"a.cue": stable + "#A: b: #_B & _#C\n",
				"b.cue": experimental + "#_B: string\n_#C: string\n",
			},
			want: []string{"names #_B, which experimental b.cue", "names _#C, which experimental b.cue"},
		},
		{
			name: "each use is reported once per declaring file",
			files: map[string]string{
				"a.cue": stable + "#A: {b: #B, c: #B}\n",
				"b.cue": experimental + "#B: string\n",
			},
			want: []string{"a.cue:3:9: stable a.cue names #B"},
		},
		{
			name: "each non-stable declaring file is reported",
			files: map[string]string{
				"a.cue": stable + "#A: b: #B\n",
				"b.cue": experimental + "#B: c?: string\n",
				"c.cue": deprecated + "#B: d?: string\n",
			},
			want: []string{"names #B, which experimental b.cue declares", "names #B, which deprecated c.cue declares"},
		},
		{
			// A top-level embedded struct adds its fields to the package, so
			// b.cue declares part of #A here too.
			name: "experimental file adds to a stable definition through a top-level embedding",
			files: map[string]string{
				"a.cue": stable + "#A: b: string\n",
				"b.cue": experimental + "{\n\t#A: c?: string\n}\n",
			},
			want: []string{"stable a.cue names #A, which experimental b.cue declares"},
		},
		{
			name: "a selector names a definition of another package",
			files: map[string]string{
				"a.cue": stable + "import \"example.com/q\"\n#A: b: q.#B\n",
				"b.cue": experimental + "#B: string\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stabilityError(t, tt.files)
			if tt.want == nil {
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("err = nil, want one containing %q", tt.want)
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("err = %v\nwant it to contain %q", err, w)
				}
			}
			if n := strings.Count(err.Error(), "\n  "); n != len(tt.want) {
				t.Errorf("err lists %d problem(s), want %d:\n%v", n, len(tt.want), err)
			}
		})
	}
}

// A stable schema that references an experimental type must stop generation
// of a real document, not only make the unit fail. Prep must also read names
// before prepare strips comprehensions, or a reference that only appears
// inside a cross-field rule would go unchecked.
func TestStableReferenceInsideRuleFailsGeneration(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"cue.mod/module.cue": "module: \"example.com/p@v0\"\nlanguage: version: \"v0.15.1\"\n",
		"a.cue":              "@gemara(status=\"stable\")\npackage p\n\n// A has a rule.\n#A: {\n\tb?: string\n\t// c is required when b is set.\n\tif b != _|_ {\n\t\tc: #B\n\t}\n}\n",
		"b.cue":              "@gemara(status=\"experimental\")\npackage p\n\n// B is experimental.\n#B: string\n",
	}
	for name, src := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c := NewConverter("", "")
	if err := c.Load(dir); err != nil {
		t.Fatal(err)
	}
	if err := c.Prep(); err != nil {
		t.Fatal(err)
	}
	if err := c.Convert(dir, WithVersion("test")); err != nil {
		t.Fatal(err)
	}
	err := c.Post()
	if err == nil || !strings.Contains(err.Error(), "a.cue:9:6: stable a.cue names #B, which experimental b.cue declares") {
		t.Fatalf("err = %v, want the reference inside the rule reported", err)
	}
}

func TestRepositorySchemasReferenceOnlyStableTypes(t *testing.T) {
	_, info, inst, err := loadPrepared("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	statuses, err := fileStatuses(inst)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkStableReferences(info.DefinitionUses, statuses); err != nil {
		t.Fatal(err)
	}
	var names int
	for _, u := range info.DefinitionUses {
		if u.decl {
			names++
		}
	}
	if names == 0 {
		t.Fatal("no top-level definitions were found; definitionUses is not reading the schemas")
	}
}
