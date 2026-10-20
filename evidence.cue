// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// Evidence records what was cited to support an opinion for a specific activity:
// raw data for the evaluation layer, evaluation and enforcement artifacts for the audit layer.
// At least one of payload or source MUST be present; an entry with neither is semantically incomplete.
//
// A tool attaching evidence gives it an id, says when it was collected and what
// it is, and either carries small content inline in payload or leaves the content
// where it lives and describes it in source. Evidence that came from somewhere
// names that source.
//
// A reviewer reading evidence can rely on the following. source.reference-id
// resolves to a source declared in the log. Within a log, one id names one item.
// Inline content carries no digest of its own, because it is as trustworthy as
// the log around it.
#Evidence: {
	// id identifies this evidence item within its log, so that a citation can
	// point at it. Within one log an id names exactly one item. It appears once
	// in any one evidence list. The same item may appear again under its id in
	// another evidence list of the log, which is how one artifact supports
	// several results, but two different items may not share an id. It identifies
	// the record in the log; source.reference-id identifies where the evidence
	// came from.
	id: string

	// type categorizes the kind of evidence
	type: #EvidenceType

	// collected-at is the timestamp when the evidence was gathered
	"collected-at": #Datetime @go(CollectedAt)

	// payload is the raw evidence data collected inline
	payload?: _ @go(Payload,type=any)

	// source identifies the artifact or system from which this evidence was collected
	source?: #EvidenceMapping @go(Source,optional=nillable)

	// description explains what this evidence represents
	description?: string
}

// EvidenceType categorizes the kind of evidence. It remains an open enum:
// recommended values include artifact types already known to Gemara (e.g.
// EvaluationLog, EnforcementLog) plus categories for common evidence forms.
#EvidenceType: #ArtifactType | string @go(-)

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

	// entry-id is the id of a catalog entry: a control or an assessment requirement.
	// It is used only when the evidence is a Gemara artifact, and selects what that
	// artifact recorded for the entry. In an audit that is the findings an
	// evaluation or enforcement log holds for that control or requirement. It
	// narrows where to look and never what to fetch: the whole artifact is
	// retrieved, and entry-id is not an input to digest. May be combined with
	// coordinate.
	"entry-id"?: string @go(EntryId)

	// digest is a cryptographic hash of the observed content at collection time; format: algorithm:encoded (e.g. sha256:abc123...)
	digest?: =~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$"

	// remarks is prose regarding this evidence reference
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
	payload?: _
	if payload == _|_ {
		source: #EvidenceMapping
	}

	// Evidence with an inline payload is carried by this log, and is as
	// trustworthy as the log itself. source may still name where it came from, by
	// reference-id. A digest is for evidence that lives elsewhere, so it cannot
	// accompany a payload.
	if payload != _|_ {
		source?: digest?: error("an inline payload cannot also have a source digest: an inline payload is as trustworthy as the log that carries it, and a digest is for evidence that lives elsewhere")
	}

	// entry-id selects what a Gemara artifact recorded for a catalog entry, so it
	// has nothing to select when the evidence is not a Gemara artifact.
	type: #EvidenceType
	if (type & #ArtifactType) == _|_ {
		source?: "entry-id"?: error("entry-id is only for evidence that is a Gemara artifact: set type to the artifact type, or remove entry-id")
	}
}
