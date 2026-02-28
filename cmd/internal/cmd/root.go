package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gemara-docs",
	Short: "Gemara CLI tool for schema conversion and documentation generation",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newCue2OpenAPICmd())
	rootCmd.AddCommand(newOpenAPI2MDCmd())
	rootCmd.AddCommand(newLexicon2MDCmd())
	rootCmd.AddCommand(newTermLinkerCmd())
}
