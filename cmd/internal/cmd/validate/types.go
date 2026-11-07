// SPDX-License-Identifier: Apache-2.0

package validate

// Severity indicates the seriousness of a validation diagnostic.
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

// Diagnostic records an issue detected during semantic or referential integrity validation.
type Diagnostic struct {
	Severity   Severity `json:"severity"`
	Code       string   `json:"code"`
	FilePath   string   `json:"filePath"`
	DocumentID string   `json:"documentId,omitempty"`
	Message    string   `json:"message"`
}

// Options configures the validator engine and output format.
type Options struct {
	Strict    bool   // When true, warnings are treated as failures
	Format    string // "text" or "json"
	Quiet     bool   // Suppress progress messages, output errors only
	SchemaDir string // Path to CUE schema directory (optional; enables AST walk discovery)
}

// MappingReference describes an external document cited in metadata.mapping-references.
type MappingReference struct {
	ID      string `yaml:"id" json:"id"`
	Title   string `yaml:"title" json:"title"`
	URL     string `yaml:"url,omitempty" json:"url,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// ReferenceUsage tracks where a reference-id was used in a document.
type ReferenceUsage struct {
	ReferenceID string
	Context     string // e.g. "source-reference", "control", "imports.catalogs"
}

// EntityUsage tracks where an entry-id within a referenced catalog was used.
type EntityUsage struct {
	ReferenceID string
	EntryID     string
	SubEntryID  string // e.g. assessment requirement id within a control
	Context     string // e.g. "mappings.source", "evaluations.control", "policy.mitigated-risk"
}

// MappingEntry represents a single mapping row in a MappingDocument.
type MappingEntry struct {
	ID           string
	Source       string
	Relationship string
	Targets      []string
	SourceRef    string
	TargetRef    string
}

// Document captures extracted metadata, declared identifiers, and references for a single file.
type Document struct {
	FilePath          string
	Title             string
	ID                string
	Type              string
	Version           string
	MappingReferences map[string]MappingReference

	// DeclaredEntities holds the top-level entity IDs declared in this document
	// (e.g. control IDs, threat IDs, capability IDs, risk IDs, guideline IDs).
	DeclaredEntities map[string]bool

	// SubEntities holds child IDs for each parent entity
	// (e.g. control ID -> set of assessment-requirement IDs).
	SubEntities map[string]map[string]bool

	// ReferenceUsages contains all places in this document where a reference-id is cited.
	ReferenceUsages []ReferenceUsage

	// EntityUsages contains all places where an entry-id within an external catalog is cited.
	EntityUsages []EntityUsage

	// Mappings contains mapping entries for MappingDocuments.
	Mappings []MappingEntry
}

// HasEntity checks whether an identifier is declared in this document as either
// a top-level entity or a nested sub-entity (e.g. an assessment requirement).
func (doc *Document) HasEntity(id string) bool {
	if doc.DeclaredEntities[id] {
		return true
	}
	for _, subs := range doc.SubEntities {
		if subs[id] {
			return true
		}
	}
	return false
}

// AddReferenceUsage records a reference usage if not already present.
func (doc *Document) AddReferenceUsage(refID, context string) {
	if refID == "" {
		return
	}
	for _, u := range doc.ReferenceUsages {
		if u.ReferenceID == refID && u.Context == context {
			return
		}
	}
	doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
		ReferenceID: refID,
		Context:     context,
	})
}

// AddEntityUsage records an entity usage if not already present.
func (doc *Document) AddEntityUsage(refID, entryID, subEntryID, context string) {
	if refID == "" || (entryID == "" && subEntryID == "") {
		return
	}
	for _, u := range doc.EntityUsages {
		if u.ReferenceID == refID && u.EntryID == entryID && u.SubEntryID == subEntryID && u.Context == context {
			return
		}
	}
	doc.EntityUsages = append(doc.EntityUsages, EntityUsage{
		ReferenceID: refID,
		EntryID:     entryID,
		SubEntryID:  subEntryID,
		Context:     context,
	})
}

// ValidationResult summarizes the outcome of validating a document bundle.
type ValidationResult struct {
	TotalFiles  int          `json:"totalFiles"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	Passed      bool         `json:"passed"`
}
