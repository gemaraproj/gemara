// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// Evidence records what was cited to support an opinion for a specific activity:
// raw data at the evaluation layer, and at the audit layer raw data as well as
// evaluation and enforcement artifacts. Gemara treats the evidence itself as
// opaque. These fields exist so that a reader can follow the chain to it.
#Evidence: {
	// id uniquely identifies this evidence
	id: string

	// type says what kind of thing this evidence is, where media-type says how to
	// read it and description says what this item shows. See #EvidenceType for
	// what to put here.
	type: #EvidenceType

	// collected-at is the timestamp when the evidence was gathered
	"collected-at": #Datetime @go(CollectedAt)

	// payload is the raw evidence data collected inline. Gemara does not interpret
	// it; source.media-type tells a tool how to read it.
	payload?: _ @go(Payload,type=any)

	// source identifies the artifact or system from which this evidence was collected and
	// retrieval information.
	source?: #EvidenceMapping @go(Source,optional=nillable)

	// description says what this evidence is. Notes from the author belong in
	// source.remarks.
	description?: string
}

// EvidenceType says what kind of thing a piece of evidence is. It is open, and
// takes one of three forms:
//
//   - A Gemara artifact type (e.g. EvaluationLog, EnforcementLog), when the evidence
//     is a Gemara artifact. Tools rely on this to validate the content and follow
//     the chain into it.
//   - A URI naming the published format the content follows, when there is one
//     (e.g. an SBOM or provenance format), so a tool can choose how to interpret it.
//   - Otherwise a short plain label (e.g. api-response). This is a label for a
//     reader only; nothing depends on its value.
#EvidenceType: #ArtifactType | string @go(-)

// EvidenceMapping identifies source identity and describes the representation used
// as evidence. reference-id identifies that source: it names the matching
// MappingReference declared in the artifact's metadata, the reusable back-matter entry
// that describes the source and may provide a location. coordinate and entry-id are
// reader hints for locating content within the referenced source. A source can yield
// multiple representations across locations and collection times, so their digest,
// size, and media type belong here rather than on the reusable MappingReference.
//
// digest, size and media-type describe the evidence content that was collected.
// download-url says where that exact content can be fetched again. Without it, the
// content is what the MappingReference's url returns when that url is the evidence
// itself, such as a log or a report. When the reference is a system such as an API,
// a reader has only the coordinate and entry-id hints to find the content, and may
// not be able to check a digest.
#EvidenceMapping: {
	// reference-id identifies where this evidence came from. It is the id of a
	// MappingReference declared in the artifact's metadata.
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

	// download-url is where this exact evidence content can be fetched. Give it
	// whenever the MappingReference's url is not itself that content, for example
	// when the reference is an API or a store and the evidence is one response or
	// file from it.
	"download-url"?: #URL @go(DownloadUrl,type=string)

	// digest is a cryptographic hash of the evidence content, as a
	// representation-scoped integrity claim: it says what the content was when
	// cited, not what the source holds now. It covers the full octet stream as
	// delivered — never a canonical re-serialization, and never a sub-resource
	// selected by coordinate or entry-id. Absence means no integrity claim was made,
	// not that the content is known unchanged, and a digest that could not be
	// checked says no more than an absent one.
	digest?: #Digest @go(Digest,type=string)

	// size is the length, in bytes, of that octet stream.
	size?: int & >=0

	// media-type is the IANA media type of the evidence content, e.g.
	// application/json, text/yaml. With an inline payload it
	// tells a tool how to read the payload.
	"media-type"?: =~"^[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*/[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*$" @go(MediaType)

	// remarks holds the author's notes about this evidence reference. What the
	// evidence is belongs in the evidence's description.
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
}

// Digest is the format of a cryptographic hash: algorithm:encoded, e.g.
// sha256:<64 lowercase hex>. sha256 is recommended. sha256, sha512 and blake3 are
// checked for that algorithm's encoding and length; any other algorithm is checked
// for grammar only, so a misspelt algorithm name passes as an unknown one. It is one
// digest rather than a set, because a set holding one matching and one mismatching
// digest has two defensible readings and this schema could not pick one.
#Digest: (=~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$" &
	(=~"^(?:sha256:[a-f0-9]{64}|sha512:[a-f0-9]{128}|blake3:[a-f0-9]{64})$" |
	!~"^(?:sha256|sha512|blake3):")) @go(-)
