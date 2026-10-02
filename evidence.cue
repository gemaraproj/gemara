// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// Evidence records what was cited to support an opinion for a specific activity:
// raw data for the evaluation layer, evaluation and enforcement artifacts for the audit layer.
#Evidence: {
	// id uniquely identifies this evidence
	id: string

	// type categorizes the kind of evidence
	type: #EvidenceType

	// collected-at is the timestamp when the evidence was gathered
	"collected-at": #Datetime @go(CollectedAt)

	// originator is the party that produced the evidence content.
	originator?: #Actor @go(Originator,optional=nillable)

	// collector is the party that gathered this evidence into the log, recorded
	// only when it differs from the log's metadata.author. Third-party evidence
	// has a real distinction to draw — a scanner produced it, an auditor pulled it
	// in — but where the log's author gathered it, saying so again would be a
	// second copy of the same fact.
	collector?: #Actor @go(Collector,optional=nillable)

	// payload is the raw evidence data collected inline
	payload?: _ @go(Payload,type=any)

	// source identifies the artifact or system from which this evidence was collected and
	// retrieval information.
	source?: #EvidenceMapping @go(Source,optional=nillable)

	// description explains what this evidence represents
	description?: string
}

// EvidenceType categorizes the kind of evidence. It remains an open enum:
// recommended values include artifact types already known to Gemara (e.g.
// EvaluationLog, EnforcementLog) plus categories for common evidence forms.
#EvidenceType: #ArtifactType | string @go(-)

// EvidenceMapping identifies source identity and describes the inline or retrieved
// representation used as evidence. reference-id identifies that source. When the
// artifact declares a matching MappingReference, it names the reusable back-matter
// entry that describes the source and may provide a location; coordinate and entry-id
// are reader hints for locating content within the referenced source. A source can
// yield multiple representations across locations and collection times, so their
// digest, size, and media type belong here rather than on the reusable MappingReference.
#EvidenceMapping: {
	// reference-id defines the evidence artifact source's identifying information.
	"reference-id": string @go(ReferenceId)

	// coordinate is the precise location within the stream identified by reference-id
	// (e.g. an API path, file path, or JSON path expression). May be combined with
	// entry-id. It is a reader hint, not an address a verifier resolves, and not an
	// input to digest.
	coordinate?: string

	// entry-id identifies a specific entry within a referenced Gemara artifact.
	// May be combined with coordinate. It is a reader hint, not an address a
	// verifier resolves, and not an input to digest.
	"entry-id"?: string @go(EntryId)

	// download-url is a retrieval-specific location for this content when it is not
	// available from the MappingReference.url.
	"download-url"?: #URL @go(DownloadUrl,type=string)

	// digest is a cryptographic hash of the full octet stream this citation represents.
	// It is a representation-scoped integrity claim, not a guarantee about the source
	// or future representations. See #Digest for what a verifier must do with it, and
	// for why absence is not assurance. It describes the cited representation rather
	// than the reusable MappingReference, which can have multiple representations.
	digest?: #Digest @go(Digest,type=string)

	// size is the length, in bytes, of that octet stream.
	size?: int & >=0

	// media-type is the IANA media type of that content, e.g. application/json,
	// text/yaml.
	"media-type"?: =~"^[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*/[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*$" @go(MediaType)

	// remarks is prose regarding this evidence reference
	remarks?: string
}

// ---- Validation ------------------------------------------------------------
// #_EvidenceStrict carries every rule on #Evidence. It is hidden (@go(-), and
// never projected), so the structures above stay shape and documentation only.

// _EvidenceStrict layers the "at least one of payload or source" rule on top of #Evidence
#_EvidenceStrict: {
	@go(-)
} & #Evidence & {
	// payload is the raw evidence data collected inline
	payload?: _

	if payload == _|_ {
		source: #EvidenceMapping
	}

	// An inline payload and a source download-url are mutually exclusive:
	// inline content has no address to retrieve it from, and content with an
	// address is referenced rather than carried. source may still name where an
	// inline payload came from, by reference-id, without a retrievable address.
	if payload != _|_ {
		source?: "download-url"?: error("an inline payload cannot also have a source download-url: inline content is carried, referenced content is addressed")
	}
}

// Digest is a cryptographic hash of a full octet stream; format: algorithm:encoded
// (e.g. sha256:<64 lowercase hex>). sha256 is the MUST-support floor for every
// conforming verifier; sha512 and blake3 are MAY. Registered algorithms are
// additionally checked for that algorithm's encoding and length; any other algorithm
// is checked for grammar only, per the OCI rule that unrecognized algorithms
// complying with the grammar pass. That is how the algorithm set extends without
// this field becoming a closed enum.
//
// The hash covers the full octet stream as delivered — never a canonical
// re-serialization, and never a sub-resource selected by coordinate or entry-id,
// which are reader hints rather than inputs to the hash. A verifier reports exactly
// one of three outcomes per citation: verified, integrity-failure, or unverifiable
// (retrieval failed, or the algorithm is unregistered or unimplemented). None of
// them may be reported as valid. Absence of digest means no integrity claim was
// made, not that the content is known unchanged.
//
#Digest: (=~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$" &
	(=~"^(?:sha256:[a-f0-9]{64}|sha512:[a-f0-9]{128}|blake3:[a-f0-9]{64})$" |
	!~"^(?:sha256|sha512|blake3):")) @go(-)
