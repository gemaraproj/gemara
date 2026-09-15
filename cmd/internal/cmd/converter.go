// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"github.com/goccy/go-yaml"
)

// ConvertOpts configures CUE-to-OpenAPI conversion.
type ConvertOpts struct {
	ManifestPath string // If set, write schema→file manifest JSON here
	Root         string // Optional #Name whose comment sets Info.Description
	Version      string // Override version (default: VERSION file or "unknown")
	Title        string // OpenAPI info title (default: "Gemara")
}

func readVersion(schemaDir string) string {
	path := filepath.Join(schemaDir, "VERSION")
	if data, err := os.ReadFile(path); err == nil {
		version := strings.TrimSpace(string(data))
		if version != "" {
			return version
		}
	}
	return "unknown"
}

func convertCUEToOpenAPI(schemaDir, outputPath string, opts ConvertOpts) error {
	if !filepath.IsAbs(schemaDir) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		schemaDir = filepath.Join(wd, schemaDir)
	}

	version := opts.Version
	if version == "" {
		version = readVersion(schemaDir)
	}
	title := opts.Title
	if title == "" {
		title = "Gemara"
	}

	v, prep, inst, err := loadPrepared(schemaDir)
	if err != nil {
		return err
	}

	// OpenAPI cannot express cross-field conditionals, so the pre-pass removes
	// them. Report what was dropped: a newly added conditional must announce
	// that it will not reach the generated document rather than disappearing.
	if len(prep.DroppedConditionals) > 0 {
		fmt.Fprintf(os.Stderr, "cue2openapi: %d cross-field conditional(s) cannot be expressed in OpenAPI and were omitted:\n", len(prep.DroppedConditionals))
		for _, pos := range prep.DroppedConditionals {
			fmt.Fprintf(os.Stderr, "  %s\n", pos)
		}
	}

	doc, err := generateRaw(v, title, version)
	if err != nil {
		return err
	}

	// The encoder emits only "3.0.0"/"3.1.0"; keep the version the published
	// document has always declared.
	doc["openapi"] = "3.0.3"
	// The previous generator always set this, and --root replaces it with the
	// named definition's doc comment. Keep both behaviours: the released
	// document has always carried a description.
	if info, ok := doc["info"].(map[string]any); ok {
		info["description"] = "Gemara schema definitions"
		desc, err := rootDescription(v, opts.Root)
		if err != nil {
			return err
		}
		if desc != "" {
			info["description"] = desc
		}
	}

	if err := unwrapTupleItems(doc); err != nil {
		return err
	}
	if err := openDisjunctions(doc); err != nil {
		return err
	}
	// Order matters from here: collapseHelperSchemas rewrites `$ref` in place,
	// and restoreRefDescriptions moves a `$ref` inside an `allOf` where that
	// rewrite no longer reaches it. Collapsing second would leave dangling
	// refs to the deleted helper schemas.
	if err := collapseHelperSchemas(doc); err != nil {
		return err
	}
	docs, err := fieldDocs(v)
	if err != nil {
		return err
	}
	if err := restoreRefDescriptions(doc, docs); err != nil {
		return err
	}
	if err := applyTimeFormats(doc, prep.TimeFormats); err != nil {
		return err
	}
	defaults, err := fieldDefaults(v)
	if err != nil {
		return err
	}
	if err := applyDefaults(doc, defaults); err != nil {
		return err
	}
	statuses, err := fileStatuses(inst)
	if err != nil {
		return err
	}
	applyStatus(doc, schemaAbsFiles(v), statuses)

	if err := writeDoc(doc, outputPath); err != nil {
		return err
	}
	if opts.ManifestPath != "" {
		components, _ := doc["components"].(map[string]any)
		schemas, _ := components["schemas"].(map[string]any)
		if schemas == nil {
			return fmt.Errorf("generated document has no components.schemas; cannot build the manifest")
		}
		files := schemaFiles(v)
		for name := range files {
			if _, ok := schemas[name]; ok {
				continue
			}
			// A hidden #_… definition is expected to be absent: collapseHelperSchemas
			// folded it into its base. Anything else missing is a generation bug,
			// and dropping it here would hide that bug behind a plausible manifest.
			if !strings.HasPrefix(name, "_") {
				return fmt.Errorf("definition #%s is declared in CUE but missing from the generated "+
					"document; refusing to write a manifest that omits it", name)
			}
			delete(files, name)
		}
		if err := writeManifest(buildManifest(files), opts.ManifestPath); err != nil {
			return err
		}
	}
	return nil
}

// rootDescription returns the doc comment of the named definition, which
// becomes info.description. An empty name leaves the default in place.
//
// A --root naming a definition that does not exist must be an error: the lookup
// yields an error value whose Doc() is nil, docText returns "", and the default
// description is silently kept — so `--root '#Metdata'` would produce a
// plausible, wrong document rather than a complaint about the typo.
func rootDescription(v cue.Value, root string) (string, error) {
	if root == "" {
		return "", nil
	}
	rv := v.LookupPath(cue.ParsePath(root))
	if err := rv.Err(); err != nil {
		return "", fmt.Errorf("--root %q does not resolve in the CUE package: %w", root, err)
	}
	return docText(rv), nil
}

// writeDoc marshals the document to YAML. goccy/go-yaml sorts map keys, which
// is what keeps schema and property order stable across runs.
func writeDoc(doc map[string]any, outputPath string) error {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}
	return os.WriteFile(outputPath, data, 0644)
}

func writeManifest(manifest map[string][]string, path string) error {
	keys := make([]string, 0, len(manifest))
	for k := range manifest {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string][]string, len(manifest))
	for _, k := range keys {
		ordered[k] = manifest[k]
	}
	data, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
