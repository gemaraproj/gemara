// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewValidateCmd creates the Cobra command for semantic validation of Gemara documents.
func NewValidateCmd() *cobra.Command {
	var opts Options

	cmd := &cobra.Command{
		Use:   "validate [paths...]",
		Short: "Validate semantic references and foreign-key integrity across Gemara documents",
		Long: `Validate parses Gemara artifact documents (catalogs, mappings, policies, and logs),
constructs an index of declared entities and mapping references, and verifies:
  1. Intra-document reference integrity: all referenced-ids must be declared in metadata.mapping-references.
  2. Cross-file foreign-key integrity: all cited entry-ids must exist in the referenced target catalogs.`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Format == "json" {
				cmd.SilenceErrors = true
			}
			return RunValidate(cmd.OutOrStdout(), cmd.ErrOrStderr(), args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Strict, "strict", false, "Treat warnings (such as unresolvable external catalogs) as errors")
	cmd.Flags().StringVar(&opts.Format, "format", "text", "Output format (text or json)")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress informational messages and only output errors")

	return cmd
}

// RunValidate executes document loading, validation, and diagnostic output.
func RunValidate(stdout, stderr io.Writer, paths []string, opts Options) error {
	docs, err := LoadPaths(paths)
	if err != nil {
		return fmt.Errorf("failed to load documents: %w", err)
	}

	if len(docs) == 0 {
		return fmt.Errorf("no Gemara YAML or JSON documents found in provided paths")
	}

	if !opts.Quiet && opts.Format == "text" {
		fmt.Fprintf(stdout, "Loaded %d Gemara document(s). Validating referential integrity...\n", len(docs))
	}

	result := Validate(docs, opts)

	if opts.Format == "json" {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return err
		}
	} else {
		for _, diag := range result.Diagnostics {
			w := stdout
			if diag.Severity == SeverityError {
				w = stderr
			}
			fmt.Fprintf(w, "[%s] %s: %s (%s)\n", diag.Severity, diag.FilePath, diag.Message, diag.Code)
		}

		if result.Passed {
			if !opts.Quiet {
				fmt.Fprintf(stdout, "SUCCESS: %d document(s) validated cleanly. No broken references found.\n", len(docs))
			}
		} else {
			errCount := 0
			for _, d := range result.Diagnostics {
				if d.Severity == SeverityError || (opts.Strict && d.Severity == SeverityWarning) {
					errCount++
				}
			}
			return fmt.Errorf("validation failed with %d error(s)", errCount)
		}
	}

	if !result.Passed {
		return fmt.Errorf("validation failed")
	}

	return nil
}
