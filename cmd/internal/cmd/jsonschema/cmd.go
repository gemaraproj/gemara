// SPDX-License-Identifier: Apache-2.0

package jsonschema

import (
	"fmt"

	"github.com/spf13/cobra"
)

type cue2JSONSchemaOpts struct {
	schemaDir  string
	outputPath string
	cueBin     string
}

// NewCue2JSONSchemaCmd creates a cobra command for converting CUE schemas to JSON Schema.
func NewCue2JSONSchemaCmd() *cobra.Command {
	opts := cue2JSONSchemaOpts{}
	cmd := &cobra.Command{
		Use:   "cue2jsonschema",
		Short: "Convert CUE schema to JSON Schema",
		Long: `Convert CUE schema definitions to Draft 2020-12 JSON Schema format.
This command processes CUE files and generates a canonical, deterministic
JSON Schema with a flat $defs dictionary suitable for downstream SDK generation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCue2JSONSchema(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.schemaDir, "schema", "s", "../..", "Path to the CUE package directory")
	cmd.Flags().StringVarP(&opts.outputPath, "output", "o", "gemara-v1.schema.json", "Output path for JSON Schema")
	cmd.Flags().StringVar(&opts.cueBin, "cue-bin", "cue", "Path to the cue binary")
	return cmd
}

func runCue2JSONSchema(opts cue2JSONSchemaOpts) error {
	gen := NewGenerator(opts.schemaDir).WithCueBin(opts.cueBin)
	if err := gen.Write(opts.outputPath); err != nil {
		return fmt.Errorf("generating JSON Schema: %w", err)
	}
	fmt.Printf("JSON Schema generated successfully at %s\n", opts.outputPath)
	return nil
}
