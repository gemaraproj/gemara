// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gemaraproj/gemara/internal/cmd/projection"
	"github.com/spf13/cobra"
)

// breakingCheckOpts configures a breaking-change check run.
type breakingCheckOpts struct {
	schemaDir     string // CUE package directory used to generate the candidate
	basePath      string // Baseline OpenAPI file to diff against (required)
	candidatePath string // Local candidate OpenAPI file; empty means generate from schemaDir
	allowPath     string // Allowlist file of oasdiff check IDs; empty means no allowlist
}

// NewBreakingCheckCmd builds the breaking-check cobra subcommand.
func NewBreakingCheckCmd() *cobra.Command {
	opts := breakingCheckOpts{}
	cmd := &cobra.Command{
		Use:   "breaking-check",
		Short: "Detect backward-incompatible schema changes against the v1 baseline",
		Long: `Generate an OpenAPI projection of the current CUE schema and diff it against
a baseline OpenAPI file using oasdiff. The baseline is supplied via --base
(fetched from the latest v1 release by the Makefile/CI). Backward-incompatible
(ERR-level) changes fail the check unless their oasdiff check ID is listed in the
allowlist.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := runBreakingCheck(opts)
			if err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("breaking changes detected")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.schemaDir, "schema", "../..", "Path to the CUE package directory")
	cmd.Flags().StringVar(&opts.basePath, "base", "", "Path to the baseline OpenAPI file to diff against (required)")
	cmd.Flags().StringVar(&opts.allowPath, "allow", "", "Path to an allowlist file of schema-scoped oasdiff exceptions (\"<id> <schema>\", one per line)")
	return cmd
}

// runBreakingCheck diffs a candidate OpenAPI projection against the baseline at
// opts.basePath, filters allowlisted changes, prints each remaining breaking
// change, and returns exit code 1 when any remain (0 otherwise). The baseline is
// supplied by the caller (fetched from releases in CI/Makefile via gh); an empty
// basePath is an error.
func runBreakingCheck(opts breakingCheckOpts) (exitCode int, err error) {
	if opts.basePath == "" {
		return 0, fmt.Errorf("--base is required (path to the baseline OpenAPI file)")
	}

	candidatePath := opts.candidatePath
	if candidatePath == "" {
		tmpDir, mkErr := os.MkdirTemp("", "gemara-breaking-candidate-")
		if mkErr != nil {
			return 0, fmt.Errorf("creating candidate temp dir: %w", mkErr)
		}
		defer func() {
			if rerr := os.RemoveAll(tmpDir); rerr != nil && err == nil {
				err = fmt.Errorf("removing candidate temp dir: %w", rerr)
			}
		}()
		candidatePath = filepath.Join(tmpDir, "openapi.yaml")
		if cerr := convertCUEToOpenAPI(opts.schemaDir, candidatePath); cerr != nil {
			return 0, fmt.Errorf("generating candidate OpenAPI: %w", cerr)
		}
	}

	basePath := opts.basePath

	base, err := loadWrapped(basePath)
	if err != nil {
		return 0, fmt.Errorf("loading baseline %q: %w", basePath, err)
	}
	rev, err := loadWrapped(candidatePath)
	if err != nil {
		return 0, fmt.Errorf("loading candidate %q: %w", candidatePath, err)
	}

	changes, err := breakingChanges(base, rev)
	if err != nil {
		return 0, fmt.Errorf("computing breaking changes: %w", err)
	}

	allowed, err := loadAllowlist(opts.allowPath)
	if err != nil {
		return 0, err
	}

	remaining, unused := filterAllowed(changes, allowed)
	for _, c := range remaining {
		fmt.Printf("ERR[%s] %s: %s\n", c.ID, c.Schema, c.Text)
	}
	if len(unused) > 0 {
		sort.Strings(unused)
		for _, u := range unused {
			fmt.Printf("ERR[unused-allowlist-entry] %s: allowlist entry did not match any breaking change\n", u)
		}
	}
	if len(remaining) > 0 || len(unused) > 0 {
		return 1, nil
	}
	return 0, nil
}

func convertCUEToOpenAPI(schemaDir, outputPath string) error {
	c := projection.NewConverter("", "")
	if err := c.Load(schemaDir); err != nil {
		return err
	}
	if err := c.Prep(); err != nil {
		return err
	}
	if err := c.Convert(schemaDir); err != nil {
		return err
	}
	if err := c.Post(); err != nil {
		return err
	}
	if err := c.WriteOpenAPI(outputPath); err != nil {
		return err
	}
	return nil
}
