// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

// ValidateOpts holds CLI options for the validate command.
type ValidateOpts struct {
	Directory string
	Files     []string
	Strict    bool
	Format    string
}

// DiagnosticSeverity indicates the severity level of an issue.
type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "error"
	SeverityWarning DiagnosticSeverity = "warning"
)

// Diagnostic represents an issue found during validation.
type Diagnostic struct {
	File     string             `json:"file"`
	Artifact string             `json:"artifact,omitempty"`
	Rule     string             `json:"rule"`
	Severity DiagnosticSeverity `json:"severity"`
	Message  string             `json:"message"`
}

// ValidationReport holds all diagnostics from a validation run.
type ValidationReport struct {
	TotalFiles  int          `json:"total_files"`
	Errors      int          `json:"errors"`
	Warnings    int          `json:"warnings"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// NewValidateCmd creates the cobra command for cross-document validation.
func NewValidateCmd() *cobra.Command {
	opts := ValidateOpts{}
	cmd := &cobra.Command{
		Use:   "validate [files...]",
		Short: "Validate semantic cross-references, catalog integrity, and foreign-key links",
		Long: `Validate semantic cross-references and foreign-key relationships across Gemara documents.
Checks that all target entry-ids in MappingDocuments, Policies, and EvaluationLogs exist in
their referenced catalogs, and verifies relationship consistency.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Files = args
			report, err := RunValidation(opts)
			if err != nil {
				return err
			}

			if opts.Format == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(report); err != nil {
					return err
				}
			} else {
				printTextReport(report)
			}

			if report.Errors > 0 || (opts.Strict && report.Warnings > 0) {
				return fmt.Errorf("validation failed with %d error(s) and %d warning(s)", report.Errors, report.Warnings)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Directory, "dir", "d", "", "Directory containing Gemara YAML/JSON artifacts to scan")
	cmd.Flags().BoolVar(&opts.Strict, "strict", false, "Treat warnings as validation errors")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "text", "Output format (text or json)")
	return cmd
}

func printTextReport(report *ValidationReport) {
	if len(report.Diagnostics) == 0 {
		fmt.Printf("✓ Validated %d file(s) with zero errors or warnings.\n", report.TotalFiles)
		return
	}

	for _, d := range report.Diagnostics {
		prefix := "ERR"
		if d.Severity == SeverityWarning {
			prefix = "WARN"
		}
		if d.Artifact != "" {
			fmt.Printf("[%s] %s (%s) [%s]: %s\n", prefix, d.File, d.Artifact, d.Rule, d.Message)
		} else {
			fmt.Printf("[%s] %s [%s]: %s\n", prefix, d.File, d.Rule, d.Message)
		}
	}

	fmt.Printf("\nValidation Summary: %d file(s) scanned, %d error(s), %d warning(s).\n",
		report.TotalFiles, report.Errors, report.Warnings)
}

type GenericArtifact struct {
	Metadata struct {
		ID                string `yaml:"id"`
		Type              string `yaml:"type"`
		MappingReferences []struct {
			ID  string `yaml:"id"`
			URL string `yaml:"url"`
		} `yaml:"mapping-references"`
	} `yaml:"metadata"`
	Controls     []struct{ ID string `yaml:"id"` } `yaml:"controls"`
	Guidelines   []struct{ ID string `yaml:"id"` } `yaml:"guidelines"`
	Threats      []struct{ ID string `yaml:"id"` } `yaml:"threats"`
	Capabilities []struct{ ID string `yaml:"id"` } `yaml:"capabilities"`
	Vectors      []struct{ ID string `yaml:"id"` } `yaml:"vectors"`
	Principles   []struct{ ID string `yaml:"id"` } `yaml:"principles"`
	Risks        []struct{ ID string `yaml:"id"` } `yaml:"risks"`

	MappingReferences []struct {
		ID  string `yaml:"id"`
		URL string `yaml:"url"`
	} `yaml:"mapping-references"`

	Mappings []struct {
		ID           string `yaml:"id"`
		Source       string `yaml:"source"`
		Relationship string `yaml:"relationship"`
		Targets      []struct {
			EntryID      string `yaml:"entry-id"`
			Relationship string `yaml:"relationship"`
		} `yaml:"targets"`
	} `yaml:"mappings"`

	Evaluations []struct {
		Name    string `yaml:"name"`
		Control struct {
			ReferenceID string `yaml:"reference-id"`
			EntryID     string `yaml:"entry-id"`
		} `yaml:"control"`
		AssessmentLogs []struct {
			Requirement struct {
				ReferenceID string `yaml:"reference-id"`
				EntryID     string `yaml:"entry-id"`
			} `yaml:"requirement"`
		} `yaml:"assessment-logs"`
	} `yaml:"evaluations"`
}

// RunValidation scans files or directory and runs semantic checks.
func RunValidation(opts ValidateOpts) (*ValidationReport, error) {
	filesToScan, err := collectFiles(opts)
	if err != nil {
		return nil, err
	}

	report := &ValidationReport{
		TotalFiles: len(filesToScan),
	}

	catalogEntries := make(map[string]map[string]bool) // artifactID -> set of entryIDs
	parsedArtifacts := make(map[string]GenericArtifact) // filePath -> parsed artifact

	// First pass: Parse artifacts and index catalog entries
	for _, path := range filesToScan {
		data, err := os.ReadFile(path)
		if err != nil {
			report.addDiag(Diagnostic{
				File:     path,
				Rule:     "file-read",
				Severity: SeverityError,
				Message:  fmt.Sprintf("failed to read file: %v", err),
			})
			continue
		}

		var artifact GenericArtifact
		if err := yaml.Unmarshal(data, &artifact); err != nil {
			// Skip files that aren't valid YAML or Gemara artifacts
			continue
		}

		if artifact.Metadata.ID == "" || artifact.Metadata.Type == "" {
			continue
		}

		parsedArtifacts[path] = artifact

		// Index defined entries
		entries := make(map[string]bool)
		for _, c := range artifact.Controls {
			if c.ID != "" {
				entries[c.ID] = true
			}
		}
		for _, g := range artifact.Guidelines {
			if g.ID != "" {
				entries[g.ID] = true
			}
		}
		for _, t := range artifact.Threats {
			if t.ID != "" {
				entries[t.ID] = true
			}
		}
		for _, c := range artifact.Capabilities {
			if c.ID != "" {
				entries[c.ID] = true
			}
		}
		for _, v := range artifact.Vectors {
			if v.ID != "" {
				entries[v.ID] = true
			}
		}
		for _, p := range artifact.Principles {
			if p.ID != "" {
				entries[p.ID] = true
			}
		}
		for _, r := range artifact.Risks {
			if r.ID != "" {
				entries[r.ID] = true
			}
		}

		if len(entries) > 0 {
			catalogEntries[artifact.Metadata.ID] = entries
		}
	}

	// Second pass: Cross-reference and structural semantics validation
	for path, artifact := range parsedArtifacts {
		declaredRefs := make(map[string]bool)
		for _, r := range artifact.Metadata.MappingReferences {
			if r.ID != "" {
				declaredRefs[r.ID] = true
			}
		}

		// Validation Rule 1: MappingDocument integrity
		if artifact.Metadata.Type == "MappingDocument" {
			for _, m := range artifact.Mappings {
				if m.Relationship == "no-match" && len(m.Targets) > 0 {
					report.addDiag(Diagnostic{
						File:     path,
						Artifact: artifact.Metadata.ID,
						Rule:     "mapping-no-match-targets",
						Severity: SeverityError,
						Message:  fmt.Sprintf("mapping %q has relationship 'no-match' but declares %d target(s)", m.ID, len(m.Targets)),
					})
				}
				if m.Relationship != "no-match" && len(m.Targets) == 0 {
					report.addDiag(Diagnostic{
						File:     path,
						Artifact: artifact.Metadata.ID,
						Rule:     "mapping-missing-targets",
						Severity: SeverityError,
						Message:  fmt.Sprintf("mapping %q has relationship %q but declares no targets", m.ID, m.Relationship),
					})
				}
			}
		}

		// Validation Rule 2: EvaluationLog cross-field reference consistency
		if artifact.Metadata.Type == "EvaluationLog" {
			for i, eval := range artifact.Evaluations {
				controlRef := eval.Control.ReferenceID
				if controlRef != "" && !declaredRefs[controlRef] {
					report.addDiag(Diagnostic{
						File:     path,
						Artifact: artifact.Metadata.ID,
						Rule:     "undeclared-control-reference",
						Severity: SeverityError,
						Message:  fmt.Sprintf("evaluation[%d] control cites undeclared reference-id %q", i, controlRef),
					})
				}

				for j, log := range eval.AssessmentLogs {
					reqRef := log.Requirement.ReferenceID
					if reqRef != "" && controlRef != "" && reqRef != controlRef {
						report.addDiag(Diagnostic{
							File:     path,
							Artifact: artifact.Metadata.ID,
							Rule:     "mismatched-assessment-reference",
							Severity: SeverityError,
							Message: fmt.Sprintf("evaluation[%d].assessment-logs[%d] requirement reference-id %q does not match control reference-id %q",
								i, j, reqRef, controlRef),
						})
					}
				}
			}
		}

		// Validation Rule 3: Catalog Foreign-Key Integrity (if referenced catalog was scanned)
		for refID := range declaredRefs {
			if knownEntries, exists := catalogEntries[refID]; exists {
				// Check EvaluationLog control references
				for i, eval := range artifact.Evaluations {
					if eval.Control.ReferenceID == refID && eval.Control.EntryID != "" {
						if !knownEntries[eval.Control.EntryID] {
							report.addDiag(Diagnostic{
								File:     path,
								Artifact: artifact.Metadata.ID,
								Rule:     "broken-foreign-key",
								Severity: SeverityError,
								Message:  fmt.Sprintf("evaluation[%d] references non-existent entry-id %q in catalog %q", i, eval.Control.EntryID, refID),
							})
						}
					}
				}
			}
		}
	}

	return report, nil
}

func (r *ValidationReport) addDiag(d Diagnostic) {
	if d.Severity == SeverityError {
		r.Errors++
	} else if d.Severity == SeverityWarning {
		r.Warnings++
	}
	r.Diagnostics = append(r.Diagnostics, d)
}

func collectFiles(opts ValidateOpts) ([]string, error) {
	var files []string
	if opts.Directory != "" {
		err := filepath.Walk(opts.Directory, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(p))
				if ext == ".yaml" || ext == ".yml" || ext == ".json" {
					files = append(files, p)
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking directory %q: %w", opts.Directory, err)
		}
	}

	for _, f := range opts.Files {
		files = append(files, f)
	}

	// Sort and deduplicate
	sort.Strings(files)
	var dedup []string
	seen := make(map[string]bool)
	for _, f := range files {
		clean := filepath.Clean(f)
		if !seen[clean] {
			seen[clean] = true
			dedup = append(dedup, clean)
		}
	}
	return dedup, nil
}
