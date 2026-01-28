package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/goccy/go-yaml"
)

type Term struct {
	Term       string   `yaml:"term"`
	Definition string   `yaml:"definition"`
	References []string `yaml:"references"`
}

type TemplateData struct {
	Table string
}

func main() {
	lexiconFile := flag.String("lexicon", "docs/lexicon.yaml", "Input lexicon YAML file")
	outputFile := flag.String("output", "docs/model/02-definitions.md", "Output markdown file")
	flag.Parse()

	// Read lexicon YAML
	data, err := os.ReadFile(*lexiconFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading lexicon file: %v\n", err)
		os.Exit(1)
	}

	var terms []Term
	if err := yaml.Unmarshal(data, &terms); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing lexicon YAML: %v\n", err)
		os.Exit(1)
	}

	// Generate table rows
	var tableRows strings.Builder
	for _, term := range terms {
		// Generate slug for anchor ID (must match termlinker logic)
		slug := termToSlug(term.Term)

		// Format term with anchor ID and bold markdown
		// Add HTML anchor tag that will be rendered by kramdown
		termName := fmt.Sprintf("<a id=\"%s\"></a>**%s**", slug, term.Term)

		// Format layer (first reference, or empty)
		layer := ""
		if len(term.References) > 0 {
			layer = term.References[0]
		}

		// Escape pipe characters in definition if any
		definition := strings.ReplaceAll(term.Definition, "|", "\\|")

		// Write table row
		tableRows.WriteString(fmt.Sprintf("| %s | %s | %s |\n", termName, definition, layer))
	}

	// Read the existing markdown file as a template
	templateContent, err := os.ReadFile(*outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading output file: %v\n", err)
		os.Exit(1)
	}

	// Parse the template
	tmpl, err := template.New("definitions").Parse(string(templateContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Execute the template with the table data
	var output bytes.Buffer
	templateData := TemplateData{
		Table: strings.TrimSpace(tableRows.String()),
	}
	if err := tmpl.Execute(&output, templateData); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	// Write the updated content
	if err := os.WriteFile(*outputFile, output.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated definitions table in %s\n", *outputFile)
}

// termToSlug converts a term to a URL-friendly slug
// This must match the logic in cmd/termlinker/main.go
func termToSlug(term string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(term)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove any other non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
