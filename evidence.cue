// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// Evidence records what was cited to support an opinion for a specific activity:
// raw data for the evaluation layer, evaluation and enforcement artifacts for the audit layer.
// At least one of payload or source MUST be present; an entry with neither is semantically incomplete.
#Evidence: {
	// id uniquely identifies this evidence
	id: string

	// type categorizes the kind of evidence
	type: #EvidenceType

	// collected-at is the timestamp when the evidence was gathered
	"collected-at": #Datetime @go(CollectedAt)

	// payload is the raw evidence data collected inline
	payload?: _ @go(Payload,type=any)

	// source identifies the artifact or system from which this evidence was collected
	source?: #EvidenceMapping @go(Source)

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

	// entry-id identifies a specific entry within a referenced Gemara artifact.
	// May be combined with coordinate to identify a sub-location within that entry's output.
	"entry-id"?: string @go(EntryId)

	// digest is a cryptographic hash of the observed content at collection time; format: algorithm:encoded (e.g. sha256:abc123...)
	digest?: =~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$"

	// remarks is prose regarding this evidence reference
	remarks?: string
}

// _EvidenceStrict layers the "at least one of payload or source" rule on top of #Evidence
#_EvidenceStrict: {
	@go(-)
} & #Evidence & {
	payload?: _
	if payload == _|_ {
		source: #EvidenceMapping
	}
}
