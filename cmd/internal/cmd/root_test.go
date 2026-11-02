// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmdStructure(t *testing.T) {
	if rootCmd.Use != "gemara-tools" {
		t.Fatalf("expected rootCmd.Use to be 'gemara-tools', got %q", rootCmd.Use)
	}

	subCommands := make(map[string]bool)
	for _, sub := range rootCmd.Commands() {
		subCommands[sub.Name()] = true
	}

	expectedSubs := []string{"cue2openapi", "breaking-check"}
	for _, exp := range expectedSubs {
		if !subCommands[exp] {
			t.Errorf("expected subcommand %q to be registered under rootCmd", exp)
		}
	}
}

func TestRootCmdHelp(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error running rootCmd --help: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "gemara-tools") {
		t.Errorf("expected help output to mention 'gemara-tools', got:\n%s", out)
	}
	if !strings.Contains(out, "cue2openapi") {
		t.Errorf("expected help output to mention 'cue2openapi', got:\n%s", out)
	}
	if !strings.Contains(out, "breaking-check") {
		t.Errorf("expected help output to mention 'breaking-check', got:\n%s", out)
	}
}

func TestRootCmdUnknownFlag(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"--nonexistent-flag"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("expected error running rootCmd with unknown flag, got nil")
	}
}

func TestExecuteHelp(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	// Execute() calls rootCmd.Execute() and only calls os.Exit(1) if err != nil.
	// With --help, err is nil so Execute() completes cleanly.
	Execute()
}
