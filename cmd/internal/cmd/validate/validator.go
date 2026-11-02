// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"
	"strings"
)

// Validate analyzes a slice of Gemara documents and checks both intra-document
// mapping reference declarations and cross-file foreign-key integrity.
func Validate(docs []*Document, opts Options) *ValidationResult {
	result := &ValidationResult{
		TotalFiles: len(docs),
		Passed:     true,
	}

	docByID := make(map[string]*Document)
	docsByType := make(map[string][]*Document)

	for _, d := range docs {
		if d.ID != "" {
			docByID[d.ID] = d
		}
		if d.Type != "" {
			docsByType[d.Type] = append(docsByType[d.Type], d)
		}
	}

	for _, doc := range docs {
		// Phase 1: Intra-Document Reference Integrity
		// Verify every reference-id cited in the document is declared in metadata.mapping-references
		for _, refUsage := range doc.ReferenceUsages {
			if _, ok := doc.MappingReferences[refUsage.ReferenceID]; !ok {
				severity := SeverityWarning
				if doc.Type == "MappingDocument" || opts.Strict {
					severity = SeverityError
				}
				diag := Diagnostic{
					Severity:   severity,
					Code:       "undeclared-mapping-reference",
					FilePath:   doc.FilePath,
					DocumentID: doc.ID,
					Message: fmt.Sprintf("reference-id %q used in %s is not declared in metadata.mapping-references",
						refUsage.ReferenceID, refUsage.Context),
				}
				result.Diagnostics = append(result.Diagnostics, diag)
			}
		}

		// Phase 2: Validate MappingDocument entries
		validateMappingEntries(doc, docByID, docsByType, opts, result)

		// Phase 3: Validate general EntityUsages (Policies, Evaluations, Logs, Controls)
		for _, entityUsage := range doc.EntityUsages {
			targetDoc := resolveTargetCatalog(doc, entityUsage.ReferenceID, docByID, docsByType)
			if targetDoc == nil {
				if opts.Strict {
					diag := Diagnostic{
						Severity:   SeverityWarning,
						Code:       "unresolved-catalog-reference",
						FilePath:   doc.FilePath,
						DocumentID: doc.ID,
						Message: fmt.Sprintf("referenced catalog %q cited in %s could not be resolved from loaded bundle",
							entityUsage.ReferenceID, entityUsage.Context),
					}
					result.Diagnostics = append(result.Diagnostics, diag)
				}
				continue
			}

			// Check top-level entity exists in target catalog
			if len(targetDoc.DeclaredEntities) > 0 && !targetDoc.DeclaredEntities[entityUsage.EntryID] {
				diag := Diagnostic{
					Severity:   SeverityError,
					Code:       "dangling-entry-reference",
					FilePath:   doc.FilePath,
					DocumentID: doc.ID,
					Message: fmt.Sprintf("entry-id %q not found in referenced catalog %q (%s, cited in %s)",
						entityUsage.EntryID, entityUsage.ReferenceID, targetDoc.FilePath, entityUsage.Context),
				}
				result.Diagnostics = append(result.Diagnostics, diag)
				continue
			}

			// Check sub-entry (e.g. assessment requirement within control)
			if entityUsage.SubEntryID != "" && targetDoc.SubEntities[entityUsage.EntryID] != nil {
				if !targetDoc.SubEntities[entityUsage.EntryID][entityUsage.SubEntryID] {
					diag := Diagnostic{
						Severity:   SeverityError,
						Code:       "dangling-subentry-reference",
						FilePath:   doc.FilePath,
						DocumentID: doc.ID,
						Message: fmt.Sprintf("sub-entry %q not found under %q in referenced catalog %q (cited in %s)",
							entityUsage.SubEntryID, entityUsage.EntryID, entityUsage.ReferenceID, entityUsage.Context),
					}
					result.Diagnostics = append(result.Diagnostics, diag)
				}
			}
		}
	}

	for _, d := range result.Diagnostics {
		if d.Severity == SeverityError || (opts.Strict && d.Severity == SeverityWarning) {
			result.Passed = false
			break
		}
	}

	return result
}

func validateMappingEntries(
	doc *Document,
	docByID map[string]*Document,
	docsByType map[string][]*Document,
	opts Options,
	result *ValidationResult,
) {
	for _, m := range doc.Mappings {
		srcDoc := resolveTargetCatalog(doc, m.SourceRef, docByID, docsByType)
		tgtDoc := resolveTargetCatalog(doc, m.TargetRef, docByID, docsByType)

		if srcDoc == nil && tgtDoc == nil {
			if opts.Strict {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity:   SeverityWarning,
					Code:       "unresolved-catalog-reference",
					FilePath:   doc.FilePath,
					DocumentID: doc.ID,
					Message: fmt.Sprintf("mapping %q references unresolved catalogs (%q, %q)",
						m.ID, m.SourceRef, m.TargetRef),
				})
			}
			continue
		}

		// Check if mapping is inverted (e.g. subsumed-by where target entry is in source catalog)
		isInverted := false
		if srcDoc != nil && len(srcDoc.DeclaredEntities) > 0 {
			if !srcDoc.DeclaredEntities[m.Source] {
				// Check if any target exists in srcDoc instead
				for _, t := range m.Targets {
					if srcDoc.DeclaredEntities[t] {
						isInverted = true
						break
					}
				}
			}
		}

		if isInverted {
			// Inverted mapping: targets come from srcDoc, source comes from tgtDoc
			if srcDoc != nil && len(srcDoc.DeclaredEntities) > 0 {
				for _, t := range m.Targets {
					if !srcDoc.DeclaredEntities[t] {
						result.Diagnostics = append(result.Diagnostics, Diagnostic{
							Severity:   SeverityError,
							Code:       "dangling-entry-reference",
							FilePath:   doc.FilePath,
							DocumentID: doc.ID,
							Message: fmt.Sprintf("entry-id %q not found in source catalog %q (mapping %q)",
								t, m.SourceRef, m.ID),
						})
					}
				}
			}
			if tgtDoc != nil && len(tgtDoc.DeclaredEntities) > 0 {
				if !tgtDoc.DeclaredEntities[m.Source] {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity:   SeverityError,
						Code:       "dangling-entry-reference",
						FilePath:   doc.FilePath,
						DocumentID: doc.ID,
						Message: fmt.Sprintf("entry-id %q not found in target catalog %q (mapping %q)",
							m.Source, m.TargetRef, m.ID),
					})
				}
			}
		} else {
			// Normal mapping: source comes from srcDoc, targets come from tgtDoc
			if srcDoc != nil && len(srcDoc.DeclaredEntities) > 0 {
				if !srcDoc.DeclaredEntities[m.Source] {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity:   SeverityError,
						Code:       "dangling-entry-reference",
						FilePath:   doc.FilePath,
						DocumentID: doc.ID,
						Message: fmt.Sprintf("entry-id %q not found in referenced catalog %q (%s, mapping %q)",
							m.Source, m.SourceRef, srcDoc.FilePath, m.ID),
					})
				}
			}
			if tgtDoc != nil && len(tgtDoc.DeclaredEntities) > 0 {
				for _, t := range m.Targets {
					if !tgtDoc.DeclaredEntities[t] {
						result.Diagnostics = append(result.Diagnostics, Diagnostic{
							Severity:   SeverityError,
							Code:       "dangling-entry-reference",
							FilePath:   doc.FilePath,
							DocumentID: doc.ID,
							Message: fmt.Sprintf("target entry-id %q not found in referenced catalog %q (%s, mapping %q)",
								t, m.TargetRef, tgtDoc.FilePath, m.ID),
						})
					}
				}
			}
		}
	}
}

// resolveTargetCatalog attempts to locate the referenced document within the loaded bundle.
func resolveTargetCatalog(
	sourceDoc *Document,
	refID string,
	docByID map[string]*Document,
	docsByType map[string][]*Document,
) *Document {
	// 1. Exact match on target document ID
	if doc, ok := docByID[refID]; ok && doc != sourceDoc {
		return doc
	}

	// 2. Match via mapping reference details
	if mr, ok := sourceDoc.MappingReferences[refID]; ok {
		// Try matching mr.Title with targetDoc.Title
		for _, doc := range docByID {
			if doc == sourceDoc {
				continue
			}
			if mr.Title != "" && doc.Title == mr.Title {
				return doc
			}
		}

		// Try matching mr.URL with targetDoc.ID or file name
		for _, doc := range docByID {
			if doc == sourceDoc {
				continue
			}
			if mr.URL != "" && (strings.Contains(doc.FilePath, refID) || strings.Contains(doc.ID, refID)) {
				return doc
			}
		}
	}

	// 3. Fallback: fuzzy ID or filename match
	lowerRef := strings.ToLower(refID)
	for _, doc := range docByID {
		if doc == sourceDoc {
			continue
		}
		if strings.Contains(strings.ToLower(doc.ID), lowerRef) ||
			strings.Contains(strings.ToLower(doc.FilePath), lowerRef) {
			return doc
		}
	}

	return nil
}
