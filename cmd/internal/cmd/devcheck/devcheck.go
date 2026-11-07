// SPDX-License-Identifier: Apache-2.0

// Package devcheck provides a unified local verification command for developers
// to validate CUE schemas, run positive/negative fixtures, check OpenAPI golden drift,
// and execute internal CLI unit test suites.
package devcheck

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gemaraproj/gemara/internal/cmd/projection"
	"github.com/spf13/cobra"
)

// CheckResult represents the outcome and elapsed duration of an individual verification step.
type CheckResult struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Message  string
}

// devCheckOptions holds configuration flags for dev-check execution.
type devCheckOptions struct {
	schemaDir    string
	goldenPath   string
	updateGolden bool
}

// NewDevCheckCmd returns a new Cobra command for local developer verification.
func NewDevCheckCmd() *cobra.Command {
	opts := devCheckOptions{}

	cmd := &cobra.Command{
		Use:   "dev-check",
		Short: "Run comprehensive local developer test and verification flow",
		Long: `Run the unified developer verification suite before creating or pushing pull requests.
It executes:
  1. CUE module and schema structure check
  2. Complete CUE schema validation test suite (positive and negative fixtures)
  3. OpenAPI golden specification drift check (with automatic --update-golden support)
  4. Internal cmd unit tests (projection, annotate, breaking change logic)
  5. Formatted execution summary table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runDevCheck(opts)
		},
	}

	defaultSchemaDir := detectSchemaDir()

	cmd.Flags().StringVarP(&opts.schemaDir, "schema", "s", defaultSchemaDir, "Path to root schema directory")
	cmd.Flags().StringVar(&opts.goldenPath, "golden", "internal/cmd/testdata/openapi.golden.yaml", "Path to OpenAPI golden specification")
	cmd.Flags().BoolVarP(&opts.updateGolden, "update-golden", "u", false, "Automatically update openapi.golden.yaml if drift is detected")

	return cmd
}

// detectSchemaDir determines the default schema directory whether invoked from root or cmd subdirectory.
func detectSchemaDir() string {
	if _, err := os.Stat("cue.mod"); err == nil {
		return "."
	}
	if _, err := os.Stat(filepath.Join("..", "cue.mod")); err == nil {
		return ".."
	}
	return "."
}

// runDevCheck coordinates and runs each verification step in order.
func runDevCheck(opts devCheckOptions) error {
	startTime := time.Now()
	var results []CheckResult

	fmt.Println("==================================================")
	fmt.Println("🚀  Gemara Unified Local Dev Verification Suite")
	fmt.Println("==================================================")
	fmt.Println()

	// 1. Check cue.mod & root schema files
	fmt.Print("⏳ [1/4] Checking CUE module sanity and schema files... ")
	resMod := checkCueModule(opts.schemaDir)
	results = append(results, resMod)
	printResult(resMod)

	// 2. Run test package (CUE schema positive & negative validation)
	fmt.Print("⏳ [2/4] Running schema test suite (70+ fixtures)... ")
	resSchema := runGoPackageTests(filepath.Join(opts.schemaDir, "test"))
	resSchema.Name = "CUE Schema Test Fixtures (test/...)"
	results = append(results, resSchema)
	printResult(resSchema)

	// 3. OpenAPI Golden Drift Check (executed before cmd unit tests so --update-golden syncs first)
	fmt.Print("⏳ [3/4] Verifying OpenAPI golden projection sync... ")
	resGolden := checkOpenAPIGolden(opts)
	results = append(results, resGolden)
	printResult(resGolden)

	// 4. Run cmd package tests
	fmt.Print("⏳ [4/4] Running internal CLI & projection unit tests... ")
	resCmd := runGoPackageTests(filepath.Join(opts.schemaDir, "cmd"))
	resCmd.Name = "CLI & Projection Logic (cmd/...)"
	results = append(results, resCmd)
	printResult(resCmd)

	// Print Summary Table
	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("📊  Verification Summary Report")
	fmt.Println("==================================================")

	allPassed := true
	for _, r := range results {
		status := "✅ PASS"
		if !r.Passed {
			status = "❌ FAIL"
			allPassed = false
		}
		fmt.Printf("%-40s | %-7s | %6.2fs\n", r.Name, status, r.Duration.Seconds())
		if !r.Passed && r.Message != "" {
			fmt.Printf("   └── Error: %s\n", strings.TrimSpace(r.Message))
		}
	}
	fmt.Println("--------------------------------------------------")
	totalDuration := time.Since(startTime)
	if allPassed {
		fmt.Printf("🎉 All checks passed successfully in %.2fs!\n", totalDuration.Seconds())
		return nil
	}

	fmt.Printf("⚠️  Some checks failed (total %.2fs). Please review the errors above.\n", totalDuration.Seconds())
	return fmt.Errorf("local developer verification failed")
}

// printResult logs the pass/fail status and elapsed duration for a step.
func printResult(r CheckResult) {
	if r.Passed {
		fmt.Printf("PASSED (%.2fs)\n", r.Duration.Seconds())
	} else {
		fmt.Printf("FAILED (%.2fs)\n", r.Duration.Seconds())
	}
}

// checkCueModule asserts that cue.mod/module.cue and essential schema definitions exist.
func checkCueModule(schemaDir string) CheckResult {
	start := time.Now()
	moduleFile := filepath.Join(schemaDir, "cue.mod", "module.cue")
	if _, err := os.Stat(moduleFile); err != nil {
		return CheckResult{
			Name:     "CUE Module Structure",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("cannot access cue.mod/module.cue at %s: %v", moduleFile, err),
		}
	}

	// Verify required core schemas exist
	requiredSchemas := []string{"policy.cue", "controlcatalog.cue", "auditlog.cue", "collections.cue"}
	for _, schema := range requiredSchemas {
		p := filepath.Join(schemaDir, schema)
		if _, err := os.Stat(p); err != nil {
			return CheckResult{
				Name:     "CUE Module Structure",
				Passed:   false,
				Duration: time.Since(start),
				Message:  fmt.Sprintf("cannot access required schema file %s at %s: %v", schema, p, err),
			}
		}
	}

	return CheckResult{
		Name:     "CUE Module Structure",
		Passed:   true,
		Duration: time.Since(start),
	}
}

// runGoPackageTests executes go test against a specified package directory.
func runGoPackageTests(pkgDir string) CheckResult {
	start := time.Now()
	cmd := exec.Command("go", "test", "-count=1", "./...")
	cmd.Dir = pkgDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	passed := (err == nil)

	return CheckResult{
		Passed:   passed,
		Duration: time.Since(start),
		Message:  out.String(),
	}
}

// normalizeLineEndings converts Windows CRLF to standard Unix LF.
func normalizeLineEndings(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

// checkOpenAPIGolden compares freshly generated OpenAPI specification against openapi.golden.yaml.
func checkOpenAPIGolden(opts devCheckOptions) CheckResult {
	start := time.Now()
	goldenFile := opts.goldenPath
	if !filepath.IsAbs(goldenFile) {
		goldenFile = filepath.Join(opts.schemaDir, "cmd", opts.goldenPath)
	}

	// Read existing golden file
	existingGolden, err := os.ReadFile(goldenFile)
	if err != nil {
		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("failed to read golden file %s: %v", goldenFile, err),
		}
	}

	// Generate fresh projection in memory
	tmpOut := filepath.Join(os.TempDir(), fmt.Sprintf("gemara-devcheck-%d.yaml", time.Now().UnixNano()))
	defer os.Remove(tmpOut)

	c := projection.NewConverter("", "")
	if err := c.Load(opts.schemaDir); err != nil {
		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("converter load failed: %v", err),
		}
	}
	if err := c.Prep(); err != nil {
		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("converter prep failed: %v", err),
		}
	}
	if err := c.Convert(opts.schemaDir, projection.WithTitle("Gemara"), projection.WithVersion("golden")); err != nil {
		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("converter convert failed: %v", err),
		}
	}
	if err := c.Post(); err != nil {
		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("converter post failed: %v", err),
		}
	}
	if err := c.WriteOpenAPI(tmpOut); err != nil {
		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("converter write failed: %v", err),
		}
	}

	generatedData, err := os.ReadFile(tmpOut)
	if err != nil {
		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("failed reading freshly generated openapi: %v", err),
		}
	}

	// Normalize CRLF to LF for cross-platform comparison
	normExisting := normalizeLineEndings(existingGolden)
	normGenerated := normalizeLineEndings(generatedData)

	if !bytes.Equal(normExisting, normGenerated) {
		if opts.updateGolden {
			if writeErr := os.WriteFile(goldenFile, normGenerated, 0o644); writeErr != nil {
				return CheckResult{
					Name:     "OpenAPI Golden Projection Sync",
					Passed:   false,
					Duration: time.Since(start),
					Message:  fmt.Sprintf("drift detected and update failed: %v", writeErr),
				}
			}
			return CheckResult{
				Name:     "OpenAPI Golden Projection Sync (Updated)",
				Passed:   true,
				Duration: time.Since(start),
			}
		}

		return CheckResult{
			Name:     "OpenAPI Golden Projection Sync",
			Passed:   false,
			Duration: time.Since(start),
			Message:  fmt.Sprintf("OpenAPI golden file drifted from schemas. Run with '--update-golden' to sync %s", goldenFile),
		}
	}

	return CheckResult{
		Name:     "OpenAPI Golden Projection Sync",
		Passed:   true,
		Duration: time.Since(start),
	}
}
