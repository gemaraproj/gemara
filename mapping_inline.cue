// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")

package gemara

// MappingReference represents a reference to an external document with full metadata.
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

	// url is the path where the artifact may be retrieved; preferably responds with Gemara-compatible YAML/JSON.
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

// Digest is a cryptographic hash of some content in the form algorithm:encoded (e.g. sha256:abc123...)
#Digest: =~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$" @go(-)

// EvidenceMapping identifies the source from which evidence was collected.
// reference-id names the MappingReference; coordinate or entry-id gives
// specificity within it; digest pins the observed content at collection time.
#EvidenceMapping: {
	// reference-id ties this evidence to a mapping-reference in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// coordinate is the precise location within the stream identified by reference-id
	// (e.g. an API path, file path, or JSON path expression). May be combined with
	// entry-id to identify a sub-location within that entry's output.
	coordinate?: string

	// entry-id identifies a specific entry within a referenced Gemara artifact.
	// May be combined with coordinate to identify a sub-location within that entry's output.
	"entry-id"?: string @go(EntryId)

	// digest is a cryptographic hash of the observed content at collection time
	digest?: #Digest @go(Digest,type=string)

	// remarks is prose regarding this evidence reference
	remarks?: string
}
