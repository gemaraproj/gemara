package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/goccy/go-yaml"
)

type OpenAPISpec struct {
	OpenAPI    string            `yaml:"openapi"`
	Info       OpenAPIInfo       `yaml:"info"`
	Components OpenAPIComponents `yaml:"components"`
}

type OpenAPIInfo struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type OpenAPIComponents struct {
	Schemas map[string]interface{} `yaml:"schemas"`
}

type Schema struct {
	Type        string                 `yaml:"type"`
	Description string                 `yaml:"description"`
	Properties  map[string]interface{} `yaml:"properties"`
	Required    []string               `yaml:"required"`
	Pattern     string                 `yaml:"pattern"`
	Format      string                 `yaml:"format"`
	Items       interface{}            `yaml:"items"`
	Ref         string                 `yaml:"$ref"`
}

type NavPage struct {
	Title    string   `yaml:"title"`
	Filename string   `yaml:"filename"`
	Schemas  []string `yaml:"schemas"`
}

type NavConfig struct {
	Pages []NavPage `yaml:"pages"`
}

func main() {
	inputFile := flag.String("input", "openapi.yaml", "Input OpenAPI YAML file")
	outputDir := flag.String("output", "spec", "Output directory for markdown files")
	manifestPath := flag.String("manifest", "", "Path to schema-manifest.json for per-file mode")
	navPath := flag.String("nav", "", "Path to schema-nav.yml for nav-based mode")
	rootsFlag := flag.String("roots", "", "Comma-separated list of root schema names (used when -manifest and -nav are not set)")
	flag.Parse()

	if *navPath != "" {
		if err := convertFromNav(*inputFile, *outputDir, *navPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else if *manifestPath != "" {
		if err := convertPerFile(*inputFile, *outputDir, *manifestPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		roots := splitRoots(*rootsFlag)
		if len(roots) == 0 {
			fmt.Fprintf(os.Stderr, "Error: -roots is required when -manifest and -nav are not set\n")
			os.Exit(1)
		}
		if err := convertOpenAPIToMarkdown(*inputFile, *outputDir, roots); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Markdown documentation generated successfully in %s/\n", *outputDir)
}

func splitRoots(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// displayNameFor returns a human-readable title for a schema file base name.
func displayNameFor(base string) string {
	switch base {
	case "base":
		return "Base"
	case "metadata":
		return "Metadata"
	case "mapping":
		return "Mapping"
	case "layer-1":
		return "Layer 1"
	case "layer-2":
		return "Layer 2"
	case "layer-3":
		return "Layer 3"
	case "layer-5":
		return "Layer 5"
	default:
		return strings.ReplaceAll(base, "-", " ")
	}
}

func loadManifest(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return m, nil
}

func loadNavFile(path string) (*NavConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read nav file: %w", err)
	}
	var nav NavConfig
	if err := yaml.Unmarshal(data, &nav); err != nil {
		return nil, fmt.Errorf("parse nav file: %w", err)
	}
	return &nav, nil
}

func slugify(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(unicode.ToLower(r))
		} else if r == ' ' || r == '-' {
			result.WriteRune('-')
		}
	}
	return result.String()
}

func convertFromNav(inputFile, outputDir, navPath string) error {
	// Load OpenAPI spec
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI file: %w", err)
	}
	var spec OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("failed to parse OpenAPI YAML: %w", err)
	}

	// Load nav file
	nav, err := loadNavFile(navPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	version := spec.Info.Version
	if version == "" {
		version = "unknown"
	}

	// For each page in nav
	for _, page := range nav.Pages {
		// Initialize a markdown buffer
		var buf strings.Builder
		// Initialize a visited map (for tracking circular references in field sections)
		visited := make(map[string]bool)

		// For each schema name listed in the page's schemas array
		for _, schemaName := range page.Schemas {
			// Look up schema in spec.Components.Schemas
			schemaData, ok := spec.Components.Schemas[schemaName]
			if !ok {
				return fmt.Errorf("schema %q not found in OpenAPI spec (referenced in page %q)", schemaName, page.Title)
			}

			// Parse schema data into Schema struct
			schemaBytes, _ := yaml.Marshal(schemaData)
			var schema Schema
			if err := yaml.Unmarshal(schemaBytes, &schema); err != nil {
				return fmt.Errorf("failed to parse schema %q: %w", schemaName, err)
			}

			// Use isAlias() to determine schema type
			if isAlias(schema) {
				buf.WriteString(generateAliasBlock(schemaName, schema))
			} else {
				buf.WriteString(generateRootSection(schemaName, schema, spec, version, visited))
			}
		}

		// Determine output filename
		filename := page.Filename
		if filename == "" {
			filename = slugify(page.Title)
		}

		// Write page buffer to {filename}.md
		outPath := filepath.Join(outputDir, filename+".md")
		if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

func convertPerFile(inputFile, outputDir, manifestPath string) error {
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI file: %w", err)
	}
	var spec OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("failed to parse OpenAPI YAML: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	version := spec.Info.Version
	if version == "" {
		version = "unknown"
	}

	fileOrder := make([]string, 0, len(manifest))
	for k := range manifest {
		fileOrder = append(fileOrder, k)
	}
	sort.Strings(fileOrder)

	visited := make(map[string]bool)

	for _, cueFile := range fileOrder {
		schemaNames := manifest[cueFile]
		if len(schemaNames) == 0 {
			continue
		}
		base := strings.TrimSuffix(cueFile, ".cue")

		var buf strings.Builder

		for _, name := range schemaNames {
			schemaData, ok := spec.Components.Schemas[name]
			if !ok {
				continue
			}
			schemaBytes, _ := yaml.Marshal(schemaData)
			var schema Schema
			if err := yaml.Unmarshal(schemaBytes, &schema); err != nil {
				continue
			}
			if isAlias(schema) {
				buf.WriteString(generateAliasBlock(name, schema))
			} else {
				buf.WriteString(generateRootSection(name, schema, spec, version, visited))
			}
		}

		outPath := filepath.Join(outputDir, base+".md")
		if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

func generateAliasBlock(name string, schema Schema) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("## `%s`\n\n", name))
	if schema.Description != "" {
		buf.WriteString(schema.Description + "\n\n")
	}
	buf.WriteString(fmt.Sprintf("- **Type**: `%s`\n", schema.Type))
	if schema.Format != "" {
		buf.WriteString(fmt.Sprintf("- **Format**: `%s`\n", schema.Format))
	}
	if schema.Pattern != "" {
		buf.WriteString(fmt.Sprintf("- **Value**: `%s`\n", schema.Pattern))
	}
	buf.WriteString("\n---\n\n")
	return buf.String()
}

func convertOpenAPIToMarkdown(inputFile, outputDir string, roots []string) error {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI file: %v", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("failed to parse OpenAPI YAML: %v", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	rootSet := make(map[string]bool)
	for _, r := range roots {
		rootSet[r] = true
	}

	// Resolve root schemas and fail if any are missing
	rootSchemas := make(map[string]Schema)
	for _, name := range roots {
		data, exists := spec.Components.Schemas[name]
		if !exists {
			return fmt.Errorf("root schema %q not found in OpenAPI spec", name)
		}
		var s Schema
		bytes, _ := yaml.Marshal(data)
		if err := yaml.Unmarshal(bytes, &s); err != nil {
			return fmt.Errorf("failed to parse root schema %q: %v", name, err)
		}
		rootSchemas[name] = s
	}

	// Collect aliases (exclude all roots)
	var aliasTypes []string
	for schemaName, schemaData := range spec.Components.Schemas {
		if rootSet[schemaName] {
			continue
		}
		schemaBytes, _ := yaml.Marshal(schemaData)
		var schema Schema
		if err := yaml.Unmarshal(schemaBytes, &schema); err != nil {
			continue
		}
		if isAlias(schema) {
			aliasTypes = append(aliasTypes, schemaName)
		}
	}
	sortStrings(aliasTypes)

	title := spec.Info.Title
	if title == "" {
		title = "Schema"
	}
	version := spec.Info.Version
	if version == "" {
		version = "unknown"
	}

	var buf strings.Builder
	visited := make(map[string]bool)

	// H1 and optional intro
	buf.WriteString(fmt.Sprintf("# %s _(%s)_\n\n", title, version))
	if spec.Info.Description != "" {
		buf.WriteString(spec.Info.Description + "\n\n")
	}

	// Table of Contents
	buf.WriteString("**Table of Contents**\n\n")
	buf.WriteString("* \n")
	buf.WriteString("{:toc}\n\n")
	buf.WriteString("---\n\n")

	// One major section per root
	for _, name := range roots {
		schema := rootSchemas[name]
		buf.WriteString(generateRootSection(name, schema, spec, version, visited))
	}

	// Aliases section
	if len(aliasTypes) > 0 {
		buf.WriteString("\n## Aliases\n\n")
		buf.WriteString("The following aliases are used throughout the schema for consistency.\n\n")

		for _, name := range aliasTypes {
			schemaBytes, _ := yaml.Marshal(spec.Components.Schemas[name])
			var schema Schema
			yaml.Unmarshal(schemaBytes, &schema)

			buf.WriteString(fmt.Sprintf("### `%s`\n\n", strings.ToLower(name)))
			if schema.Description != "" {
				buf.WriteString(schema.Description + "\n\n")
			}
			buf.WriteString(fmt.Sprintf("- **Type**: `%s`\n", schema.Type))
			if schema.Format != "" {
				buf.WriteString(fmt.Sprintf("- **Format**: `%s`\n", schema.Format))
			}
			if schema.Pattern != "" {
				buf.WriteString(fmt.Sprintf("- **Value**: `%s`\n", schema.Pattern))
			}
			buf.WriteString("\n---\n\n")
		}
	}

	outputPath := filepath.Join(outputDir, "schema.md")
	if err := os.WriteFile(outputPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", outputPath, err)
	}

	return nil
}

func sortStrings(s []string) {
	sort.Strings(s)
}

func isAlias(schema Schema) bool {
	// Aliases are anything that is NOT an object with properties
	// This includes: string types (with or without patterns), boolean, and simple object types
	return schema.Properties == nil
}

func resolveSchemaRef(ref string, spec OpenAPISpec) (*Schema, error) {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil, fmt.Errorf("invalid ref format: %s", ref)
	}

	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")
	schemaData, exists := spec.Components.Schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", schemaName)
	}

	schemaBytes, _ := yaml.Marshal(schemaData)
	var schema Schema
	if err := yaml.Unmarshal(schemaBytes, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema %s: %v", schemaName, err)
	}

	return &schema, nil
}

func generateRootSection(rootName string, schema Schema, spec OpenAPISpec, version string, visited map[string]bool) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("## `%s`\n\n", rootName))
	if schema.Description != "" {
		buf.WriteString(schema.Description + "\n\n")
	}

	if len(schema.Required) > 0 || schema.Properties != nil {
		if len(schema.Required) > 0 {
			buf.WriteString("Required:\n\n")
			required := make([]string, len(schema.Required))
			copy(required, schema.Required)
			sort.Strings(required)
			for _, req := range required {
				buf.WriteString(fmt.Sprintf("- `%s`\n", req))
			}
		}
		if schema.Properties != nil {
			var optional []string
			for propName := range schema.Properties {
				isRequired := false
				for _, req := range schema.Required {
					if req == propName {
						isRequired = true
						break
					}
				}
				if !isRequired {
					optional = append(optional, propName)
				}
			}
			sort.Strings(optional)
			if len(optional) > 0 {
				buf.WriteString("\nOptional:\n\n")
				for _, opt := range optional {
					buf.WriteString(fmt.Sprintf("- `%s`\n", opt))
				}
			}
		}
		buf.WriteString("\n---\n\n")
	}

	if schema.Properties != nil {
		propNames := make([]string, 0, len(schema.Properties))
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propData := schema.Properties[propName]
			propBytes, _ := yaml.Marshal(propData)
			var prop Schema
			yaml.Unmarshal(propBytes, &prop)

			isRequired := false
			for _, req := range schema.Required {
				if req == propName {
					isRequired = true
					break
				}
			}
			buf.WriteString(generateFieldSection(propName, prop, spec, 3, "", visited, !isRequired))
		}
	}

	return buf.String()
}

func generateFieldSection(fieldName string, fieldSchema Schema, spec OpenAPISpec, headingLevel int, prefix string, visited map[string]bool, isOptional bool) string {
	var buf strings.Builder

	// Build field path
	fieldPath := fieldName
	if prefix != "" {
		fieldPath = prefix + "." + fieldName
	}

	// Generate heading
	heading := strings.Repeat("#", headingLevel)
	optionalText := ""
	if isOptional {
		optionalText = " (optional)"
	}
	buf.WriteString(fmt.Sprintf("%s `%s`%s\n\n", heading, fieldPath, optionalText))

	// Handle $ref - resolve and recurse
	if fieldSchema.Ref != "" {
		refType := strings.TrimPrefix(fieldSchema.Ref, "#/components/schemas/")

		// Check if it's an alias - if so, just show the type reference
		refSchema, err := resolveSchemaRef(fieldSchema.Ref, spec)
		if err == nil && isAlias(*refSchema) {
			// Just show description and type reference
			if fieldSchema.Description != "" {
				buf.WriteString(fieldSchema.Description + "\n\n")
			}
			buf.WriteString(fmt.Sprintf("- **Type**: [%s]\n", refType))
			buf.WriteString("\n---\n\n")
			return buf.String()
		}

		// Prevent infinite recursion
		if visited[refType] {
			if fieldSchema.Description != "" {
				buf.WriteString(fieldSchema.Description + "\n\n")
			}
			buf.WriteString(fmt.Sprintf("- **Type**: [%s]\n\n", refType))
			buf.WriteString("---\n\n")
			return buf.String()
		}

		visited[refType] = true
		defer delete(visited, refType)

		// Resolve the referenced schema
		refSchema, err = resolveSchemaRef(fieldSchema.Ref, spec)
		if err != nil {
			buf.WriteString(fmt.Sprintf("Error resolving reference: %v\n\n", err))
			return buf.String()
		}

		// Show description (from field or referenced type)
		description := fieldSchema.Description
		if description == "" {
			description = refSchema.Description
		}
		if description != "" {
			buf.WriteString(description + "\n\n")
		}

		// Show type reference (refType already declared above)
		buf.WriteString(fmt.Sprintf("- **Type**: [%s]\n", refType))

		// Show required vs optional for the referenced type
		if len(refSchema.Required) > 0 || refSchema.Properties != nil {
			buf.WriteString("\n")
			if len(refSchema.Required) > 0 {
				buf.WriteString(fmt.Sprintf("Required if `%s` is present:\n\n", fieldPath))
				required := make([]string, len(refSchema.Required))
				copy(required, refSchema.Required)
				sort.Strings(required)
				for _, req := range required {
					buf.WriteString(fmt.Sprintf("- `%s`\n", req))
				}
			}
			if refSchema.Properties != nil {
				var optional []string
				for propName := range refSchema.Properties {
					isRequired := false
					for _, req := range refSchema.Required {
						if req == propName {
							isRequired = true
							break
						}
					}
					if !isRequired {
						optional = append(optional, propName)
					}
				}
				sort.Strings(optional)
				if len(optional) > 0 {
					buf.WriteString("\nOptional:\n\n")
					for _, opt := range optional {
						buf.WriteString(fmt.Sprintf("- `%s`\n", opt))
					}
				}
			}
			buf.WriteString("\n---\n\n")
		}

		// Recursively generate nested fields
		if refSchema.Properties != nil {
			// Sort property names for deterministic output
			propNames := make([]string, 0, len(refSchema.Properties))
			for propName := range refSchema.Properties {
				propNames = append(propNames, propName)
			}
			sort.Strings(propNames)

			for _, propName := range propNames {
				propData := refSchema.Properties[propName]
				propBytes, _ := yaml.Marshal(propData)
				var prop Schema
				yaml.Unmarshal(propBytes, &prop)

				propIsRequired := false
				for _, req := range refSchema.Required {
					if req == propName {
						propIsRequired = true
						break
					}
				}

				buf.WriteString(generateFieldSection(propName, prop, spec, headingLevel+1, fieldPath, visited, !propIsRequired))
			}
		}
	} else if fieldSchema.Type == "object" && fieldSchema.Properties != nil {
		// Inline object with properties - recurse into it
		if fieldSchema.Description != "" {
			buf.WriteString(fieldSchema.Description + "\n\n")
		}

		// Show required vs optional for the inline object
		if len(fieldSchema.Required) > 0 || fieldSchema.Properties != nil {
			if len(fieldSchema.Required) > 0 {
				buf.WriteString(fmt.Sprintf("Required if `%s` is present:\n\n", fieldPath))
				required := make([]string, len(fieldSchema.Required))
				copy(required, fieldSchema.Required)
				sort.Strings(required)
				for _, req := range required {
					buf.WriteString(fmt.Sprintf("- `%s`\n", req))
				}
			}
			if fieldSchema.Properties != nil {
				var optional []string
				for propName := range fieldSchema.Properties {
					isRequired := false
					for _, req := range fieldSchema.Required {
						if req == propName {
							isRequired = true
							break
						}
					}
					if !isRequired {
						optional = append(optional, propName)
					}
				}
				sort.Strings(optional)
				if len(optional) > 0 {
					buf.WriteString("\nOptional:\n\n")
					for _, opt := range optional {
						buf.WriteString(fmt.Sprintf("- `%s`\n", opt))
					}
				}
			}
			buf.WriteString("\n---\n\n")
		}

		// Recursively generate nested fields
		// Sort property names for deterministic output
		propNames := make([]string, 0, len(fieldSchema.Properties))
		for propName := range fieldSchema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propData := fieldSchema.Properties[propName]
			propBytes, _ := yaml.Marshal(propData)
			var prop Schema
			yaml.Unmarshal(propBytes, &prop)

			propIsRequired := false
			for _, req := range fieldSchema.Required {
				if req == propName {
					propIsRequired = true
					break
				}
			}

			buf.WriteString(generateFieldSection(propName, prop, spec, headingLevel+1, fieldPath, visited, !propIsRequired))
		}
	} else {
		// Simple field (no $ref, no inline object)
		if fieldSchema.Description != "" {
			buf.WriteString(fieldSchema.Description + "\n\n")
		}

		if fieldSchema.Type != "" {
			buf.WriteString(fmt.Sprintf("- **Type**: `%s`\n", fieldSchema.Type))
		}

		if fieldSchema.Pattern != "" {
			buf.WriteString(fmt.Sprintf("- **Matches Pattern**: `%s`\n", fieldSchema.Pattern))
		}

		// Handle array items
		if fieldSchema.Type == "array" && fieldSchema.Items != nil {
			itemsBytes, _ := yaml.Marshal(fieldSchema.Items)
			var itemsSchema Schema
			if err := yaml.Unmarshal(itemsBytes, &itemsSchema); err == nil {
				if itemsSchema.Ref != "" {
					refType := strings.TrimPrefix(itemsSchema.Ref, "#/components/schemas/")
					buf.WriteString(fmt.Sprintf("- **Items**: [%s]\n", refType))
				} else if itemsSchema.Type != "" {
					buf.WriteString(fmt.Sprintf("- **Items**: `%s`\n", itemsSchema.Type))
				}
			}
		}

		buf.WriteString("\n---\n\n")
	}

	return buf.String()
}

func generateSchemaMarkdown(name string, schema Schema, spec OpenAPISpec, version string) string {
	var buf strings.Builder

	// Title
	buf.WriteString(fmt.Sprintf("# `%s` _(%s)_\n\n", strings.ToLower(name), version))

	// Description
	if schema.Description != "" {
		buf.WriteString(schema.Description + "\n\n")
	}

	// Required vs Optional
	if len(schema.Required) > 0 || schema.Properties != nil {
		if len(schema.Required) > 0 {
			buf.WriteString(fmt.Sprintf("Required if `%s` is present:\n\n", strings.ToLower(name)))
			required := make([]string, len(schema.Required))
			copy(required, schema.Required)
			sort.Strings(required)
			for _, req := range required {
				buf.WriteString(fmt.Sprintf("- `%s`\n", req))
			}
		}
		if schema.Properties != nil {
			var optional []string
			for propName := range schema.Properties {
				isRequired := false
				for _, req := range schema.Required {
					if req == propName {
						isRequired = true
						break
					}
				}
				if !isRequired {
					optional = append(optional, propName)
				}
			}
			sort.Strings(optional)
			if len(optional) > 0 {
				buf.WriteString("\nOptional:\n\n")
				for _, opt := range optional {
					buf.WriteString(fmt.Sprintf("- `%s`\n", opt))
				}
			}
		}
		buf.WriteString("\n---\n\n")
	}

	// Properties
	if schema.Properties != nil {
		// Sort property names for deterministic output
		propNames := make([]string, 0, len(schema.Properties))
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propData := schema.Properties[propName]
			propBytes, _ := yaml.Marshal(propData)
			var prop Schema
			yaml.Unmarshal(propBytes, &prop)

			buf.WriteString(fmt.Sprintf("## `%s.%s", strings.ToLower(name), propName))
			isRequired := false
			for _, req := range schema.Required {
				if req == propName {
					isRequired = true
					break
				}
			}
			if !isRequired {
				buf.WriteString(" (optional)")
			}
			buf.WriteString("`\n\n")

			if prop.Description != "" {
				buf.WriteString(fmt.Sprintf("- **Description**: %s\n", prop.Description))
			}

			if prop.Ref != "" {
				refType := strings.TrimPrefix(prop.Ref, "#/components/schemas/")
				buf.WriteString(fmt.Sprintf("- **Type**: [%s]\n", refType))
			} else if prop.Type != "" {
				buf.WriteString(fmt.Sprintf("- **Type**: `%s`\n", prop.Type))
			}

			if prop.Pattern != "" {
				buf.WriteString(fmt.Sprintf("- **Matches Pattern**: `%s`\n", prop.Pattern))
			}

			buf.WriteString("\n---\n\n")
		}
	}

	return buf.String()
}
