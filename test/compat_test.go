// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modconfig"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/module"
	"golang.org/x/mod/semver"
)

const modulePath = "github.com/gemaraproj/gemara"

var builtinValidatorNoise = []string{
	"time.Format",
	"time.Time",
	"strings.MinRunes",
}

var skipDefs = map[string]bool{
	"#Datetime": true,
}

func TestNoBreakingChanges(t *testing.T) {
	ctx := context.Background()

	resolver, err := modconfig.NewResolver(&modconfig.Config{
		CUERegistry: modconfig.DefaultRegistry,
	})
	if err != nil {
		t.Fatalf("failed to create resolver: %v", err)
	}
	regClient := modregistry.NewClientWithResolver(resolver)

	includePrerelease := os.Getenv("GEMARA_COMPAT_PRERELEASE") == "true"

	latestVer, err := latestVersion(ctx, regClient, modulePath, includePrerelease)
	if err != nil {
		t.Logf("no suitable release found | skipping compatibility check: %v", err)
		t.Skip()
	}
	t.Logf("comparing against released version: %s", latestVer)

	reg, err := modconfig.NewRegistry(&modconfig.Config{
		CUERegistry: modconfig.DefaultRegistry,
	})
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}

	oldSchema, err := loadModuleFromRegistry(reg, latestVer)
	if err != nil {
		t.Fatalf("failed to load released module: %v", err)
	}

	schemaDir, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("failed to resolve schema directory: %v", err)
	}

	stableDefs, err := collectStableDefs(schemaDir)
	if err != nil {
		t.Fatalf("failed to collect stable definitions: %v", err)
	}
	t.Logf("found %d stable definitions to check", len(stableDefs))

	for _, defPath := range stableDefs {
		defPath := defPath
		if skipDefs[defPath] {
			t.Logf("skipping %s (excluded from compatibility check)", defPath)
			continue
		}
		t.Run(defPath, func(t *testing.T) {
			newDef := schemaValue.LookupPath(cue.ParsePath(defPath))
			if newDef.Err() != nil {
				t.Fatalf("new schema: lookup %s: %v", defPath, newDef.Err())
			}

			oldDef := oldSchema.LookupPath(cue.ParsePath(defPath))
			if oldDef.Err() != nil {
				t.Logf("definition %s not found in released version (new addition)", defPath)
				return
			}

			t.Logf("validating %s: checking new definition subsumes old definition", defPath)
			if err := newDef.Subsume(oldDef, cue.Raw(), cue.Schema()); err != nil {
				var realErrors []string
				for _, e := range cueerrors.Errors(err) {
					msg := e.Error()
					if !matchesAny(msg, builtinValidatorNoise) {
						realErrors = append(realErrors, msg)
					}
				}
				if len(realErrors) > 0 {
					t.Errorf("breaking change detected in %s:\n%s", defPath, strings.Join(realErrors, "\n"))
				} else {
					t.Logf("%s: OK — no breaking changes (suppressed builtin validator noise)", defPath)
				}
			} else {
				t.Logf("%s: OK — no breaking changes", defPath)
			}
		})
	}
}

func collectStableDefs(schemaDir string) ([]string, error) {
	var stableDefs []string

	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("read schema dir: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".cue") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		content := string(data)
		if !strings.Contains(content, `@status("stable")`) {
			continue
		}
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") && strings.Contains(line, ":") {
				def := strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
				stableDefs = append(stableDefs, def)
			}
		}
	}
	return stableDefs, nil
}

func matchesAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func latestVersion(ctx context.Context, client *modregistry.Client, modPath string, includePrerelease bool) (module.Version, error) {
	versions, err := client.ModuleVersions(ctx, modPath+"@v1")
	if err != nil {
		return module.Version{}, fmt.Errorf("listing versions for %s: %w", modPath, err)
	}
	for i := len(versions) - 1; i >= 0; i-- {
		v := versions[i]
		if includePrerelease || semver.Prerelease(v) == "" {
			return module.NewVersion(modPath, v)
		}
	}
	if includePrerelease {
		return module.Version{}, fmt.Errorf("no versions found for %s", modPath)
	}
	return module.Version{}, fmt.Errorf("no stable release found for %s (set GEMARA_COMPAT_PRERELEASE=true to include pre-releases)", modPath)
}

func loadModuleFromRegistry(reg modconfig.Registry, ver module.Version) (cue.Value, error) {
	instances := load.Instances([]string{ver.String()}, &load.Config{
		Registry: reg,
	})
	if len(instances) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE instances returned for %v", ver)
	}
	if err := instances[0].Err; err != nil {
		return cue.Value{}, fmt.Errorf("loading module %v: %w", ver, err)
	}
	val := schemaCtx.BuildInstance(instances[0])
	if err := val.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("building schema for %v: %w", ver, err)
	}
	return val, nil
}
