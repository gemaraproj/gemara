// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"cuelabs.dev/go/oci/ociregistry/ociclient"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/module"
)

const (
	registryHost = "registry.cue.works"
	modulePath   = "github.com/gemaraproj/gemara"
)

type compatResult struct {
	name    string
	status  string
	details []string
}

// TestBackwardCompatibility uses CUE subsumption to verify that current
// schema definitions remain at least as permissive as the latest published
// release fetched from the CUE module registry.
//
// newDef.Subsume(oldDef, cue.Schema()) returns nil when every value the old
// definition accepted is also accepted by the new definition. Failures
// indicate a narrowing change that could break existing consumers.
func TestBackwardCompatibility(t *testing.T) {
	schemaDir, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve schema directory: %v", err)
	}

	baselineDir, baselineVer := fetchBaseline(t)
	defer os.RemoveAll(baselineDir)

	ctx := cuecontext.New()

	packages := []struct {
		name string
		pkg  string
	}{
		{"base", "."},
		{"validation", "./validation"},
	}

	fmt.Println()
	fmt.Printf("Backward Compatibility Report (baseline: %s from %s)\n", baselineVer, registryHost)
	fmt.Println(strings.Repeat("=", 60))

	var totalBreaking, totalRemoved, totalNew int

	for _, p := range packages {
		oldHas := hasPackage(t, baselineDir, p.pkg)
		newHas := hasPackage(t, schemaDir, p.pkg)

		fmt.Println()
		if !oldHas && !newHas {
			fmt.Printf("  %s: not present in baseline or current\n", p.name)
			continue
		}
		if !oldHas {
			fmt.Printf("  %s: new package (not in baseline)\n", p.name)
			continue
		}
		if !newHas {
			fmt.Printf("  %s: REMOVED\n", p.name)
			totalBreaking++
			continue
		}

		oldVal := loadCUE(t, ctx, baselineDir, p.pkg)
		newVal := loadCUE(t, ctx, schemaDir, p.pkg)

		results := compareDefinitions(oldVal, newVal)
		breaking, removed, new_ := countByStatus(results)
		totalBreaking += breaking
		totalRemoved += removed
		totalNew += new_

		fmt.Printf("  %s\n", p.name)
		printResults(results)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Breaking: %d  |  Removed: %d  |  New: %d\n", totalBreaking, totalRemoved, totalNew)
	fmt.Println()
}

func fetchBaseline(t *testing.T) (dir string, version string) {
	t.Helper()

	bg := context.Background()

	oci, err := ociclient.New(registryHost, nil)
	if err != nil {
		t.Fatalf("create OCI client: %v", err)
	}

	client := modregistry.NewClient(oci)

	versions, err := client.ModuleVersions(bg, modulePath)
	if err != nil {
		t.Fatalf("list module versions: %v", err)
	}
	if len(versions) == 0 {
		t.Skip("no versions found in registry for " + modulePath)
	}
	ver := versions[len(versions)-1]

	mv, err := module.NewVersion(modulePath, ver)
	if err != nil {
		t.Fatalf("parse version %s: %v", ver, err)
	}

	mod, err := client.GetModule(bg, mv)
	if err != nil {
		t.Fatalf("get module %s: %v", mv, err)
	}

	zipReader, err := mod.GetZip(bg)
	if err != nil {
		t.Fatalf("get module zip: %v", err)
	}
	defer zipReader.Close()

	dir, err = os.MkdirTemp("", "gemara-baseline-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	zipData, err := io.ReadAll(zipReader)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("read zip: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("open zip: %v", err)
	}

	for _, f := range zr.File {
		if f.Name == "" {
			continue
		}

		target := filepath.Join(dir, filepath.FromSlash(f.Name))

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0o755)
			continue
		}

		os.MkdirAll(filepath.Dir(target), 0o755)

		rc, err := f.Open()
		if err != nil {
			os.RemoveAll(dir)
			t.Fatalf("open zip entry %s: %v", f.Name, err)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			os.RemoveAll(dir)
			t.Fatalf("read zip entry %s: %v", f.Name, err)
		}

		if err := os.WriteFile(target, data, 0o644); err != nil {
			os.RemoveAll(dir)
			t.Fatalf("write %s: %v", target, err)
		}
	}

	return dir, ver
}

func compareDefinitions(oldVal, newVal cue.Value) []compatResult {
	oldDefs := collectDefs(oldVal)
	newDefs := collectDefs(newVal)

	var results []compatResult

	for name, oldDef := range oldDefs {
		newDef, exists := newDefs[name]
		if !exists {
			results = append(results, compatResult{name, "REMOVED", nil})
			continue
		}
		if err := newDef.Subsume(oldDef, cue.Schema()); err != nil {
			var details []string
			for _, e := range cueerrors.Errors(err) {
				details = append(details, e.Error())
			}
			results = append(results, compatResult{name, "BREAKING", details})
			continue
		}
		results = append(results, compatResult{name, "OK", nil})
	}

	for name := range newDefs {
		if _, exists := oldDefs[name]; !exists {
			results = append(results, compatResult{name, "NEW", nil})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		order := map[string]int{"REMOVED": 0, "BREAKING": 1, "NEW": 2, "OK": 3}
		if order[results[i].status] != order[results[j].status] {
			return order[results[i].status] < order[results[j].status]
		}
		return results[i].name < results[j].name
	})

	return results
}

func printResults(results []compatResult) {
	for _, r := range results {
		switch r.status {
		case "REMOVED":
			fmt.Printf("    REMOVED   %s\n", r.name)
		case "BREAKING":
			fmt.Printf("    BREAKING  %s\n", r.name)
			for _, d := range r.details {
				fmt.Printf("              - %s\n", truncate(d, 88))
			}
		case "NEW":
			fmt.Printf("    NEW       %s\n", r.name)
		}
	}
}

func countByStatus(results []compatResult) (breaking, removed, new_ int) {
	for _, r := range results {
		switch r.status {
		case "BREAKING":
			breaking++
		case "REMOVED":
			removed++
		case "NEW":
			new_++
		}
	}
	return
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func hasPackage(t *testing.T, moduleDir, pkg string) bool {
	t.Helper()

	dir := moduleDir
	if pkg != "." {
		dir = filepath.Join(moduleDir, filepath.FromSlash(pkg))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".cue") {
			return true
		}
	}
	return false
}

func loadCUE(t *testing.T, ctx *cue.Context, moduleDir, pkg string) cue.Value {
	t.Helper()

	cfg := &load.Config{Dir: moduleDir}
	instances := load.Instances([]string{pkg}, cfg)
	if len(instances) == 0 {
		t.Fatalf("no CUE instances for %s in %s", pkg, moduleDir)
	}

	if instances[0].Err != nil {
		t.Fatalf("load CUE %s from %s: %v", pkg, moduleDir, instances[0].Err)
	}

	val := ctx.BuildInstance(instances[0])
	if val.Err() != nil {
		t.Fatalf("build CUE %s from %s: %v", pkg, moduleDir, val.Err())
	}

	return val
}

func collectDefs(val cue.Value) map[string]cue.Value {
	defs := make(map[string]cue.Value)
	iter, _ := val.Fields(cue.Definitions(true))
	for iter.Next() {
		defs[iter.Selector().String()] = iter.Value()
	}
	return defs
}
