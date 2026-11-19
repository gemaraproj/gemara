// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/goccy/go-yaml"
)

// Converter owns one CUE-to-OpenAPI conversion lifecycle.
type Converter struct {
	schemaDir    string
	manifestPath string
	root         string
	value        cue.Value
	prep         *prepInfo
	instance     *build.Instance
	doc          *openapi3.T
}

// NewConverter creates a converter with the supplied output metadata.
func NewConverter(manifestPath, root string) *Converter {
	return &Converter{manifestPath: manifestPath, root: root}
}

// Load resolves the CUE package location. Prep loads and rewrites the
// package because the OpenAPI pre-pass operates on its syntax files.
func (c *Converter) Load(schemaDir string) error {
	if !filepath.IsAbs(schemaDir) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		schemaDir = filepath.Join(wd, schemaDir)
	}
	c.schemaDir = schemaDir
	return nil
}

// Convert invokes the upstream OpenAPI generator and retains its document for
// Post and the write methods.
func (c *Converter) Convert(schemaDir string, options ...ConvertOption) error {
	if c.instance == nil {
		return fmt.Errorf("preprocess a CUE package before converting")
	}

	opts := defaultProjectionOpts(schemaDir)
	opts.apply(options...)

	doc, err := generateRaw(c.value, opts.title, opts.version)
	if err != nil {
		return err
	}
	c.doc = doc

	// generateRaw asks the encoder for 3.0.0; declare 3.0.3, the version the
	// released document has always carried.
	doc.OpenAPI = "3.0.3"
	// The previous generator always set this, and --root replaces it with the
	// named definition's doc comment. Keep both behaviours: the released
	// document has always carried a description.
	if doc.Info != nil {
		doc.Info.Description = "Gemara schema definitions"
		desc, err := rootDescription(c.value, c.root)
		if err != nil {
			return err
		}
		if desc != "" {
			doc.Info.Description = desc
		}
	}

	return nil
}

// WriteOpenAPI writes the fully converted and postprocessed document.
func (c *Converter) WriteOpenAPI(outputPath string) error {
	data, err := c.EncodeOpenAPI()
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

// EncodeOpenAPI marshals the fully converted and postprocessed document to YAML bytes.
func (c *Converter) EncodeOpenAPI() ([]byte, error) {
	if c.doc == nil {
		return nil, fmt.Errorf("convert a CUE package before writing OpenAPI")
	}
	data, err := yaml.Marshal(c.doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}
	return data, nil
}

// WriteManifest writes the schema-to-source-file manifest for the current document.
func (c *Converter) WriteManifest(manifestPath string) error {
	data, err := c.EncodeManifest()
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0644)
}

// EncodeManifest builds and marshals the schema-to-source-file manifest to JSON bytes.
func (c *Converter) EncodeManifest() ([]byte, error) {
	if c.doc == nil {
		return nil, fmt.Errorf("convert a CUE package before writing a manifest")
	}
	if c.doc.Components == nil || c.doc.Components.Schemas == nil {
		return nil, fmt.Errorf("generated document has no components.schemas; cannot build the manifest")
	}
	files := schemaFiles(c.value)
	for name := range files {
		if _, ok := c.doc.Components.Schemas[name]; ok {
			continue
		}
		// A hidden #_… definition is expected to be absent: collapseHelperSchemas
		// folded it into its base. Anything else missing is a generation bug,
		// and dropping it here would hide that bug behind a plausible manifest.
		if !strings.HasPrefix(name, "_") {
			return nil, fmt.Errorf("definition #%s is declared in CUE but missing from the generated "+
				"document; refusing to write a manifest that omits it", name)
		}
		delete(files, name)
	}
	data, err := json.MarshalIndent(buildManifest(files), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}
	return data, nil
}
