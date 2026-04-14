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
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modconfig"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/module"
	"golang.org/x/mod/semver"
)

const modulePath = "github.com/gemaraproj/gemara"

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

	localSchema, err := loadLocalSchemaRelaxed(schemaDir)
	if err != nil {
		t.Fatalf("failed to load local schema: %v", err)
	}

	stableDefs, err := collectStableDefs(schemaDir)
	if err != nil {
		t.Fatalf("failed to collect stable definitions: %v", err)
	}
	t.Logf("found %d stable definitions to check", len(stableDefs))

	for _, defPath := range stableDefs {
		defPath := defPath
		t.Run(defPath, func(t *testing.T) {
			newDef := localSchema.LookupPath(cue.ParsePath(defPath))
			if newDef.Err() != nil {
				t.Fatalf("new schema: lookup %s: %v", defPath, newDef.Err())
			}

			oldDef := oldSchema.LookupPath(cue.ParsePath(defPath))
			if oldDef.Err() != nil {
				t.Logf("definition %s not found in released version (new addition)", defPath)
				return
			}

			if err := newDef.Subsume(oldDef, cue.Raw(), cue.Schema()); err != nil {
				t.Errorf("breaking change detected in %s:\n%v", defPath, err)
			}
		})
	}
}

// loadLocalSchemaRelaxed loads the local CUE schema with builtin validators
// and hidden constraint fields relaxed so that CUE's Subsume does not produce
// false positives when comparing values from different load contexts
// (filesystem vs OCI registry).
func loadLocalSchemaRelaxed(schemaDir string) (cue.Value, error) {
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return cue.Value{}, fmt.Errorf("read schema dir: %w", err)
	}

	overlay := make(map[string]load.Source)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".cue") {
			continue
		}
		absPath := filepath.Join(schemaDir, entry.Name())
		original, err := os.ReadFile(absPath)
		if err != nil {
			return cue.Value{}, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		relaxed := relaxForSubsume(string(original))
		if relaxed != string(original) {
			overlay[absPath] = load.FromString(relaxed)
		}
	}

	cfg := &load.Config{
		Dir:     schemaDir,
		Overlay: overlay,
	}
	instances := load.Instances([]string{"."}, cfg)
	if len(instances) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE instances returned")
	}
	if err := instances[0].Err; err != nil {
		return cue.Value{}, fmt.Errorf("loading local schema: %w", err)
	}
	val := schemaCtx.BuildInstance(instances[0])
	if err := val.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("building local schema: %w", err)
	}
	return val, nil
}

// relaxForSubsume strips builtin validators and hidden constraint fields
// that cause cross-context Subsume false positives. The time.Format validator
// and list.Contains-based group validation both fail when compared across
// independently loaded CUE instances.
func relaxForSubsume(content string) string {
	crossContextNoise := []string{
		"_validGroupIds",
		"_groupValidation",
		"_validApplicabilityIds",
		"_applicabilityValidation",
		"// Unify the valid ID list with a list.Contains constraint",
	}

	var lines []string
	for _, line := range strings.Split(content, "\n") {
		skip := false
		for _, p := range crossContextNoise {
			if strings.Contains(line, p) {
				skip = true
				break
			}
		}
		if !skip {
			lines = append(lines, line)
		}
	}
	result := strings.Join(lines, "\n")

	result = strings.Replace(result,
		`#Datetime: time.Format("2006-01-02T15:04:05Z07:00")`,
		`#Datetime: string`, 1)

	if !strings.Contains(result, "time.") {
		result = strings.Replace(result, `import "time"`, "", 1)
	}
	if !strings.Contains(result, "list.") {
		result = strings.Replace(result, `import "list"`, "", 1)
	}

	return result
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
