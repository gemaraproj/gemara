package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

type Schema struct {
	Type        string                 `yaml:"type"`
	Description string                 `yaml:"description"`
	Properties  map[string]interface{} `yaml:"properties"`
	Required    []string               `yaml:"required"`
	Pattern     string                 `yaml:"pattern"`
	Format      string                 `yaml:"format"`
	Items       interface{}            `yaml:"items"`
	Ref         string                 `yaml:"$ref"`
	XStatus     string                 `yaml:"x-status"`
}

type NavPage struct {
	Title    string   `yaml:"title"`
	Filename string   `yaml:"filename"`
	Schemas  []string `yaml:"schemas"`
}

type NavConfig struct {
	Pages []NavPage `yaml:"pages"`
}

var openAPI2MDCmd = &cobra.Command{
	Use:   "openapi2md",
	Short: "Convert OpenAPI YAML to Markdown documentation",
	Long: `Convert OpenAPI 3.0.3 YAML specifications to Markdown documentation.
Supports three modes:
  - Navigation-based: Uses a nav.yml file to organize schemas into pages
  - Manifest-based: Uses a manifest.json to map CUE files to schemas
  - Roots-based: Uses a comma-separated list of root schema names`,
	RunE: runOpenAPI2MD,
}

var openAPI2MDFlags struct {
	inputFile    string
	outputDir    string
	manifestPath string
	navPath      string
	rootsFlag    string
}

func newOpenAPI2MDCmd() *cobra.Command {
	openAPI2MDCmd.Flags().StringVarP(&openAPI2MDFlags.inputFile, "input", "i", "openapi.yaml", "Input OpenAPI YAML file")
	openAPI2MDCmd.Flags().StringVarP(&openAPI2MDFlags.outputDir, "output", "o", "spec", "Output directory for markdown files")
	openAPI2MDCmd.Flags().StringVarP(&openAPI2MDFlags.manifestPath, "manifest", "m", "", "Path to schema-manifest.json for per-file mode")
	openAPI2MDCmd.Flags().StringVarP(&openAPI2MDFlags.navPath, "nav", "n", "", "Path to schema-nav.yml for nav-based mode")
	openAPI2MDCmd.Flags().StringVarP(&openAPI2MDFlags.rootsFlag, "roots", "r", "", "Comma-separated list of root schema names (used when -manifest and -nav are not set)")
	return openAPI2MDCmd
}

func runOpenAPI2MD(cmd *cobra.Command, args []string) error {
	if openAPI2MDFlags.navPath != "" {
		if err := convertFromNav(openAPI2MDFlags.inputFile, openAPI2MDFlags.outputDir, openAPI2MDFlags.navPath); err != nil {
			return err
		}
	} else if openAPI2MDFlags.manifestPath != "" {
		if err := convertPerFile(openAPI2MDFlags.inputFile, openAPI2MDFlags.outputDir, openAPI2MDFlags.manifestPath); err != nil {
			return err
		}
	} else {
		roots := splitRoots(openAPI2MDFlags.rootsFlag)
		if len(roots) == 0 {
			return fmt.Errorf("Error: -roots is required when -manifest and -nav are not set")
		}
		if err := convertOpenAPIToMarkdown(openAPI2MDFlags.inputFile, openAPI2MDFlags.outputDir, roots); err != nil {
			return err
		}
	}

	fmt.Printf("Markdown documentation generated successfully in %s/\n", openAPI2MDFlags.outputDir)
	return nil
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

	// Build schema-to-filename map for generating links
	schemaToFile := make(map[string]string)
	for _, page := range nav.Pages {
		filename := page.Filename
		if filename == "" {
			filename = slugify(page.Title)
		}
		for _, schemaName := range page.Schemas {
			schemaToFile[schemaName] = filename
		}
	}

	// For each page in nav
	for _, page := range nav.Pages {
		var buf strings.Builder

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
				buf.WriteString(generateAliasBlock(schemaName, schema, false))
			} else {
				buf.WriteString(generateRootSection(schemaName, schema, spec, schemaToFile))
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

	fileOrder := make([]string, 0, len(manifest))
	for k := range manifest {
		fileOrder = append(fileOrder, k)
	}
	sort.Strings(fileOrder)

	// Empty map since we don't have nav file info in this mode
	schemaToFile := make(map[string]string)

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
				buf.WriteString(generateAliasBlock(name, schema, false))
			} else {
				buf.WriteString(generateRootSection(name, schema, spec, schemaToFile))
			}
		}

		outPath := filepath.Join(outputDir, base+".md")
		if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

func generateAliasBlock(name string, schema Schema, subheading bool) string {
	var buf strings.Builder
	level := "##"
	if subheading {
		level = "###"
	}
	buf.WriteString(fmt.Sprintf("%s `%s`\n\n", level, name))
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
		return fmt.Errorf("failed to read OpenAPI file: %w", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("failed to parse OpenAPI YAML: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
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
			return fmt.Errorf("failed to parse root schema %q: %w", name, err)
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
	sort.Strings(aliasTypes)

	title := spec.Info.Title
	if title == "" {
		title = "Schema"
	}
	version := spec.Info.Version
	if version == "" {
		version = "unknown"
	}

	var buf strings.Builder
	// Empty map since we don't have nav file info in this mode
	schemaToFile := make(map[string]string)

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
		buf.WriteString(generateRootSection(name, schema, spec, schemaToFile))
	}

	// Aliases section
	if len(aliasTypes) > 0 {
		buf.WriteString("\n## Aliases\n\n")
		buf.WriteString("The following aliases are used throughout the schema for consistency.\n\n")

		for _, name := range aliasTypes {
			schemaBytes, _ := yaml.Marshal(spec.Components.Schemas[name])
			var schema Schema
			if err := yaml.Unmarshal(schemaBytes, &schema); err != nil {
				continue
			}
			buf.WriteString(generateAliasBlock(name, schema, true))
		}
	}

	outputPath := filepath.Join(outputDir, "schema.md")
	if err := os.WriteFile(outputPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", outputPath, err)
	}

	return nil
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

// getSchemaStatus extracts the gemara-status extension value from a schema.
func getSchemaStatus(schema Schema) string {
	if schema.XStatus != "" {
		return schema.XStatus
	}
	return ""
}

// formatStatusBadge returns a markdown badge for the status.
func formatStatusBadge(status string) string {
	switch status {
	case "experimental":
		return "<span class=\"badge badge-experimental\">Experimental</span>"
	case "stable":
		return "<span class=\"badge badge-stable\">Stable</span>"
	case "deprecated":
		return "<span class=\"badge badge-deprecated\">Deprecated</span>"
	default:
		return ""
	}
}

// formatFieldInline formats a field's information and returns (fieldLine, description)
// fieldLine format: `field` **type** _Required_ or `field` **type**
// description is returned separately
func formatFieldInline(fieldName string, fieldSchema Schema, spec OpenAPISpec, prefix string, isRequired bool, schemaToFile map[string]string) (string, string) {
	// Field name with full path
	fieldPath := fieldName
	if prefix != "" {
		fieldPath = prefix + "." + fieldName
	}

	// Type
	typeStr := formatFieldType(fieldSchema, spec, schemaToFile)

	// Build field line: `field` **type** _Required_ or `field` **type**
	var fieldLineParts []string
	fieldLineParts = append(fieldLineParts, fmt.Sprintf("`%s`", fieldPath))
	if typeStr != "" {
		fieldLineParts = append(fieldLineParts, fmt.Sprintf("**%s**", typeStr))
	}
	if isRequired {
		fieldLineParts = append(fieldLineParts, "_Required_")
	}
	fieldLine := strings.Join(fieldLineParts, " ")

	// Description
	description := fieldSchema.Description
	if fieldSchema.Ref != "" {
		refSchema, err := resolveSchemaRef(fieldSchema.Ref, spec)
		if err == nil {
			if description == "" {
				description = refSchema.Description
			}
		}
	}

	return fieldLine, description
}

// formatFieldType returns the type string for a field with markdown links for custom types
func formatFieldType(fieldSchema Schema, spec OpenAPISpec, schemaToFile map[string]string) string {
	if fieldSchema.Ref != "" {
		refType := strings.TrimPrefix(fieldSchema.Ref, "#/components/schemas/")
		// Check if this is a custom type that should be linked
		if filename, exists := schemaToFile[refType]; exists {
			// Create markdown link: [TypeName](filename#typename) - no .md extension for Jekyll
			anchor := strings.ToLower(refType)
			return fmt.Sprintf("[%s](%s#%s)", refType, filename, anchor)
		}
		// Return just the type name if not found in schema map
		return refType
	}

	if fieldSchema.Type != "" {
		typeStr := fieldSchema.Type

		// Handle array items - format as array[Type]
		if fieldSchema.Type == "array" && fieldSchema.Items != nil {
			itemsBytes, _ := yaml.Marshal(fieldSchema.Items)
			var itemsSchema Schema
			if err := yaml.Unmarshal(itemsBytes, &itemsSchema); err == nil {
				var itemType string
				var itemTypeLink string
				if itemsSchema.Ref != "" {
					refType := strings.TrimPrefix(itemsSchema.Ref, "#/components/schemas/")
					// Check if this is a custom type that should be linked
					if filename, exists := schemaToFile[refType]; exists {
						anchor := strings.ToLower(refType)
						itemTypeLink = fmt.Sprintf("[%s](%s#%s)", refType, filename, anchor)
					} else {
						itemTypeLink = refType
					}
					itemType = itemTypeLink
				} else if itemsSchema.Type != "" {
					itemType = itemsSchema.Type
				}
				if itemType != "" {
					typeStr = fmt.Sprintf("array[%s]", itemType)
				}
			}
		}

		return typeStr
	}

	return ""
}

// formatFieldWithNested formats a field inline (nested expansion disabled).
func formatFieldWithNested(fieldName string, fieldSchema Schema, spec OpenAPISpec, isRequired bool, schemaToFile map[string]string) string {
	var buf strings.Builder
	fieldLine, description := formatFieldInline(fieldName, fieldSchema, spec, "", isRequired, schemaToFile)
	buf.WriteString(fieldLine + "\n\n")
	if description != "" {
		buf.WriteString(description + "\n")
	}
	return buf.String()
}

func generateRootSection(rootName string, schema Schema, spec OpenAPISpec, schemaToFile map[string]string) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("## `%s`\n\n", rootName))
	if schema.Description != "" {
		buf.WriteString(schema.Description + "\n\n")
	}

	if schema.Properties != nil {
		propNames := make([]string, 0, len(schema.Properties))
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		// Add status badge if present
		if status := getSchemaStatus(schema); status != "" {
			buf.WriteString(fmt.Sprintf(" %s", formatStatusBadge(status)))
		}
		buf.WriteString("\n\n")

		if schema.Description != "" {
			buf.WriteString(schema.Description + "\n\n")
		}

		// Output all fields in order (required first, then optional)
		// Sort by required status, then by name
		type fieldInfo struct {
			name     string
			schema   Schema
			required bool
		}
		var fields []fieldInfo

		for _, propName := range propNames {
			isRequired := false
			for _, req := range schema.Required {
				if req == propName {
					isRequired = true
					break
				}
			}

			propData := schema.Properties[propName]
			propBytes, _ := yaml.Marshal(propData)
			var prop Schema
			if err := yaml.Unmarshal(propBytes, &prop); err != nil {
				continue
			}
			fields = append(fields, fieldInfo{
				name:     propName,
				schema:   prop,
				required: isRequired,
			})
		}

		// Sort: required first, then by name
		sort.Slice(fields, func(i, j int) bool {
			if fields[i].required != fields[j].required {
				return fields[i].required // required fields come first
			}
			return fields[i].name < fields[j].name
		})

		// Output all fields
		for _, field := range fields {
			buf.WriteString(formatFieldWithNested(field.name, field.schema, spec, field.required, schemaToFile))
			buf.WriteString("\n")
		}
	}

	return buf.String()
}
