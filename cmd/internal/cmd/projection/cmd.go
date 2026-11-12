// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type cue2OpenAPIOpts struct {
	schemaDir    string
	outputPath   string
	manifestPath string
	root         string
	version      string
	title        string
}

func NewCue2OpenAPICmd() *cobra.Command {
	opts := cue2OpenAPIOpts{}
	cmd := &cobra.Command{
		Use:   "cue2openapi",
		Short: "Convert CUE schema to OpenAPI YAML",
		Long: `Convert CUE schema definitions to OpenAPI 3.0.3 YAML format.
This command processes CUE files and generates an OpenAPI specification
that can be used for API documentation and validation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCue2OpenAPI(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.schemaDir, "schema", "s", "../..", "Path to the CUE package directory")
	cmd.Flags().StringVarP(&opts.outputPath, "output", "o", "openapi.yaml", "Output path for OpenAPI schema")
	cmd.Flags().StringVarP(&opts.manifestPath, "manifest", "m", "", "Optional path to write schema→file manifest JSON")
	cmd.Flags().StringVarP(&opts.root, "root", "r", "", "Optional root definition (#Name) whose comment sets spec description")
	cmd.Flags().StringVarP(&opts.version, "version", "v", "", "Optional version string (default: VERSION file in schema dir or \"unknown\")")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "Gemara", "OpenAPI info title")
	return cmd
}

func samePath(p1, p2 string) (bool, error) {
	if p1 == p2 {
		return true, nil
	}

	clean1 := filepath.Clean(p1)
	clean2 := filepath.Clean(p2)
	if clean1 == clean2 {
		return true, nil
	}

	abs1, err1 := filepath.Abs(p1)
	abs2, err2 := filepath.Abs(p2)
	if err1 == nil && err2 == nil && abs1 == abs2 {
		return true, nil
	}

	fi1, statErr1 := os.Stat(p1)
	fi2, statErr2 := os.Stat(p2)
	if statErr1 == nil && statErr2 == nil && os.SameFile(fi1, fi2) {
		return true, nil
	}

	return false, nil
}

func runCue2OpenAPI(opts cue2OpenAPIOpts) error {
	if opts.manifestPath != "" {
		same, err := samePath(opts.outputPath, opts.manifestPath)
		if err != nil {
			return fmt.Errorf("comparing output and manifest paths: %w", err)
		}
		if same {
			return errors.New("output and manifest paths must not refer to the same file")
		}
	}

	c := NewConverter(opts.manifestPath, opts.root)
	if err := c.Load(opts.schemaDir); err != nil {
		return err
	}
	if err := c.Prep(); err != nil {
		return err
	}
	options := []ConvertOption{WithTitle(opts.title)}
	if opts.version != "" {
		options = append(options, WithVersion(opts.version))
	}
	if err := c.Convert(opts.schemaDir, options...); err != nil {
		return err
	}
	if err := c.Post(); err != nil {
		return err
	}

	for _, input := range c.InputFiles() {
		same, err := samePath(opts.outputPath, input)
		if err != nil {
			return fmt.Errorf("checking output path against input file %s: %w", input, err)
		}
		if same {
			return fmt.Errorf("output path %s would overwrite input file %s", opts.outputPath, input)
		}

		if opts.manifestPath != "" {
			sameManifest, err := samePath(opts.manifestPath, input)
			if err != nil {
				return fmt.Errorf("checking manifest path against input file %s: %w", input, err)
			}
			if sameManifest {
				return fmt.Errorf("manifest path %s would overwrite input file %s", opts.manifestPath, input)
			}
		}
	}

	if err := c.WriteOpenAPI(opts.outputPath); err != nil {
		return err
	}
	if opts.manifestPath != "" {
		if err := c.WriteManifest(opts.manifestPath); err != nil {
			return err
		}
	}

	fmt.Printf("OpenAPI schema generated successfully at %s\n", opts.outputPath)
	return nil
}
