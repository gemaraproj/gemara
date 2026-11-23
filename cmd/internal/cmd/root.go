// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/gemaraproj/gemara/internal/cmd/devcheck"
	"github.com/gemaraproj/gemara/internal/cmd/diff"
	"github.com/gemaraproj/gemara/internal/cmd/projection"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gemara-tools",
	Short: "Gemara CLI tool for schema conversion, validation, and diff",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(projection.NewCue2OpenAPICmd())
	rootCmd.AddCommand(diff.NewBreakingCheckCmd())
	rootCmd.AddCommand(devcheck.NewDevCheckCmd())
}
