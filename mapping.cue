// Schema lifecycle: experimental | stable | deprecated
@status("stable")

package schemas

// ============================================================================
// Mapping Types - MappingReference, MappingEntry, MultiMapping, SingleMapping
// ============================================================================

// MappingReference represents a reference to an external document with full metadata.
#MappingReference: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the purpose of this mapping reference at a glance
	title: string

	// version is the version identifier of the artifact being mapped to
	version: string

	// description is prose regarding the artifact's purpose or content
	description?: string

	// url is the path where the artifact may be retrieved; preferrably responds with Gemara-compatible YAML/JSON
	url?: =~"^(https?|file)://[^\\s]+$"
}

// MultiMapping represents a mapping to an external reference with one or more entries.
#MultiMapping: {
	// ReferenceId should reference the corresponding MappingReference id from metadata
	"reference-id": string @go(ReferenceId)

	// entries is a list of mapping entries
	entries: [#MappingEntry, ...#MappingEntry] @go(Entries)

	// remarks is prose regarding the mapped artifact or the mapping relationship
	remarks?: string
}

// SingleMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.
#SingleMapping: {
	// reference-id is the id for a MappingReference entry in the artifact's metadata
	"reference-id"?: string @go(ReferenceId)

	// entry-id is the identifier being mapped to in the referenced artifact
	"entry-id": string @go(EntryId)

	// strength is the author's estimate of how completely the current/source material satisfies the target/reference material;
	// Range: 1-10. Zero value means not yet quantified.
	strength?: int & >=1 & <=10

	// remarks is prose describing the mapping relationship
	remarks?: string
}

// MappingEntry represents a single entry within a mapping
#MappingEntry: {
	// reference-id is the id for a MappingReference entry in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// strength is the author's estimate of how completely the current/source material satisfies the target/reference material;
	// Range: 1-10. Zero value means not yet quantified.
	strength?: int & >=1 & <=10

	// remarks is prose describing the mapping relationship
	remarks?: string
}
