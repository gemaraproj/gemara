// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// Evidence records what was cited to support an opinion for a specific activity:
// raw data at the evaluation layer, and at the audit layer raw data as well as
// evaluation and enforcement artifacts. Gemara treats the evidence itself as
// opaque. These fields exist so that a reader can follow the chain to it.
//
// A tool attaching evidence gives it an id, says when it was collected and what
// it is, and either carries small content inline in payload or leaves the content
// where it lives and describes it in source. Evidence that came from somewhere
// names that source.
//
// A reviewer reading evidence can rely on the following. source.reference-id
// resolves to a source declared in the log. Within a log, one id names one item.
// A digest is well-formed for its algorithm. Inline content carries no digest of
// its own, because it is as trustworthy as the log around it.
#Evidence: {
	// id identifies this evidence item within its log, so that a citation can
	// point at it. Within one log an id names exactly one item: the same item may
	// be repeated under its id, but two different items may not share one. It
	// identifies the record in the log; source.reference-id identifies where the
	// evidence came from.
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

	// entry-id is the id of a catalog entry: a control or an assessment requirement.
	// It is used only when the evidence is a Gemara artifact, and selects what that
	// artifact recorded for the entry. In an audit that is the findings an
	// evaluation or enforcement log holds for that control or requirement. It
	// narrows where to look and never what to fetch: the whole artifact is
	// retrieved, and entry-id is not an input to digest. May be combined with
	// coordinate.
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
	// application/json, text/yaml. Unlike digest and size it may accompany an
	// inline payload, where it tells a tool how to read the payload.
	"media-type"?: =~"^[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*/[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*$" @go(MediaType)

	// remarks holds the author's notes about this evidence reference. What the
	// evidence is belongs in the evidence's description.
	remarks?: string
}

// ---- Validation ------------------------------------------------------------
// #_EvidenceStrict carries the rules one evidence item can check about itself. It
// is hidden (@go(-), and never projected), so the structures above stay shape and
// documentation only. Rules across a log's evidence (one id per item, declared
// sources) live on #EvaluationLog and #AuditLog.

// _EvidenceStrict layers the "at least one of payload or source" rule on top of #Evidence
#_EvidenceStrict: {
	@go(-)
} & #Evidence & {
	// payload is the raw evidence data collected inline
	payload?: _

	if payload == _|_ {
		source: #EvidenceMapping
	}

	// Evidence with an inline payload is carried by this log, and is as
	// trustworthy as the log itself. source may still name where it came from, by
	// reference-id, and its media-type, which a tool may need to read the
	// payload. download-url, digest and size are for evidence that lives
	// elsewhere, so they cannot accompany a payload.
	if payload != _|_ {
		source?: "download-url"?: error("an inline payload cannot also have a source download-url: inline content is carried, referenced content is addressed")
		source?: digest?:         error("an inline payload cannot also have a source digest: an inline payload is as trustworthy as the log that carries it, and a digest is for evidence that lives elsewhere")
		source?: size?:           error("an inline payload cannot also have a source size: size describes evidence that lives elsewhere")
	}

	// entry-id selects what a Gemara artifact recorded for a catalog entry, so it
	// has nothing to select when the evidence is not a Gemara artifact.
	type: #EvidenceType
	if (type & #ArtifactType) == _|_ {
		source?: "entry-id"?: error("entry-id is only for evidence that is a Gemara artifact: set type to the artifact type, or remove entry-id")
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
