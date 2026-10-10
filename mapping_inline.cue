// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")

package gemara

// Fields constrained by #URL and #Digest (evidence.cue) pin their Go projection to string. A named
// Go type here buys no validation — Go has no constructor to enforce the pattern —
// and would cost every consumer a conversion at every use site. The constraint is
// the value of the definition; the Go name is not.

// URL is a URI with a scheme. Any scheme is accepted (e.g. https, file, oci, s3,
// arn), so content hosted outside http(s) can be addressed.
#URL: =~"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]+$" @go(-)

// MappingReference is a reusable back-matter entry describing an external source or
// authority. It provides source identity and may provide a location, but does not
// itself represent a retrieved representation.
#MappingReference: {
	// id identifies this mapping reference within the artifact and, when url
	// is absent, the referenced artifact's metadata.id.
	id: string

	// title describes the purpose of this mapping reference at a glance
	title: string

	// version is the version identifier of the artifact being mapped to
	version: string

	// description is prose regarding the artifact's purpose or content
	description?: string

	// url is an optional location from which a representation of the referenced source
	// may be retrieved; preferably responds with Gemara-compatible YAML/JSON.
	// Any URI scheme is accepted (e.g. https, file, oci, s3, arn) so evidence can be
	// addressed wherever it actually lives.
	url?: =~"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]+$"
}

// ArtifactMapping represents a mapping to an external artifact or artifact entry
#ArtifactMapping: {
	// reference-id identifies an element from a MappingReference in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// remarks is prose regarding the mapped artifact or the mapping relationship
	remarks?: string
}

// MultiEntryMapping represents a mapping to an external reference with one or more entries.
#MultiEntryMapping: {
	// top-level reference to the MappingReference entry
	#ArtifactMapping

	// entries is a list of mapping entries
	entries: [#ArtifactMapping, ...#ArtifactMapping] @go(Entries)
}

// EntryMapping represents how a specific entry maps to a MappingReference.
#EntryMapping: {
	// reference-id is the id for a MappingReference entry in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// entry-id is the identifier being mapped to in the referenced artifact
	"entry-id": string @go(EntryId)

	// remarks is prose describing the mapping relationship
	remarks?: string
}
