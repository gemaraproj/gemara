package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cue2OpenAPICmd = &cobra.Command{
	Use:   "cue2openapi",
	Short: "Convert CUE schema to OpenAPI YAML",
	Long: `Convert CUE schema definitions to OpenAPI 3.0.3 YAML format.
This command processes CUE files and generates an OpenAPI specification
that can be used for API documentation and validation.`,
	RunE: runCue2OpenAPI,
}

var cue2OpenAPIFlags struct {
	schemaDir    string
	outputPath   string
	manifestPath string
	root         string
	version      string
	title        string
}

func newCue2OpenAPICmd() *cobra.Command {
	cue2OpenAPICmd.Flags().StringVarP(&cue2OpenAPIFlags.schemaDir, "schema", "s", "../..", "Path to the CUE package directory")
	cue2OpenAPICmd.Flags().StringVarP(&cue2OpenAPIFlags.outputPath, "output", "o", "openapi.yaml", "Output path for OpenAPI schema")
	cue2OpenAPICmd.Flags().StringVarP(&cue2OpenAPIFlags.manifestPath, "manifest", "m", "", "Optional path to write schema→file manifest JSON")
	cue2OpenAPICmd.Flags().StringVarP(&cue2OpenAPIFlags.root, "root", "r", "", "Optional root definition (#Name) whose comment sets spec description")
	cue2OpenAPICmd.Flags().StringVarP(&cue2OpenAPIFlags.version, "version", "v", "", "Optional version string (default: VERSION file in schema dir or \"unknown\")")
	cue2OpenAPICmd.Flags().StringVarP(&cue2OpenAPIFlags.title, "title", "t", "Gemara", "OpenAPI info title")
	return cue2OpenAPICmd
}

func runCue2OpenAPI(cmd *cobra.Command, args []string) error {
	if err := convertCUEToOpenAPI(cue2OpenAPIFlags.schemaDir, cue2OpenAPIFlags.outputPath, ConvertOpts{
		ManifestPath: cue2OpenAPIFlags.manifestPath,
		Root:         cue2OpenAPIFlags.root,
		Version:      cue2OpenAPIFlags.version,
		Title:        cue2OpenAPIFlags.title,
	}); err != nil {
		return err
	}

	fmt.Printf("OpenAPI schema generated successfully at %s\n", cue2OpenAPIFlags.outputPath)
	return nil
}
