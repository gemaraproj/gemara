// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// AuditLog records results from an audit performed against a target resource
#AuditLog: {
	#Log
	metadata: type: "AuditLog"

	// owner defines the RACI roles responsible for managing the audit
	owner?: #RACI @go(Owner)

	// summary provides the high-level conclusion
	summary: string

	// criteria defines the acceptable state for the audited resource
	criteria: [#ArtifactMapping, ...#ArtifactMapping]

	// results records audit results against the criteria
	results: [#AuditResult, ...#AuditResult] @go(Results,type=[]*AuditResult)

	if results != _|_ {
		_uniqueResultIds: {for i, r in results {(r.id): i}}
	}
}

// ResultType classifies the nature of an audit result
#ResultType: "Gap" | "Finding" | "Observation" | "Strength" @go(-)

// AuditResult records a single result with supporting evidence and recommendations.
#AuditResult: {
	// id uniquely identifies this result
	id: string

	// title describes this result at a glance
	title: string

	// type classifies the nature of this result
	type: #ResultType

	// description explains the result in detail
	description: string

	// criteria-reference maps this result to specific criteria entries
	"criteria-reference": #MultiEntryMapping @go(CriteriaReference)

	// evidence records the data sources that support this result
	evidence?: [#Evidence, ...#Evidence] @go(Evidence)

	// recommendations records corrective actions for this result
	recommendations?: [#Recommendation, ...#Recommendation] @go(Recommendations)
}

// Recommendation provides a corrective action for an audit result
#Recommendation: {
	// id uniquely identifies this recommendation
	id?: string

	// text describes the recommended corrective action
	text: string

	// required indicates whether this recommendation is a mandatory corrective action
	required: *false | bool
}

// Evidence records what was cited to support an opinion for a specific activity:
// raw data for the evaluation layer, evaluation and enforcement artifacts for the audit layer.
#Evidence: {
	// id uniquely identifies this evidence
	id: string

	// type categorizes the kind of evidence
	type: #EvidenceType

	// collected-at is the timestamp when the evidence was gathered
	"collected-at": #Datetime @go(CollectedAt)

	// payload is the raw evidence data collected
	payload?: _ @go(Payload,type=any)

	// digest is a hash of the evidence content at collection time for integrity verification.
	// Enables verification that mutable evidence (S3 objects, HTTP URLs, API responses)
	// has not changed since the tool observed it. Not needed for content-addressable systems
	// (OCI, git) or inline payloads.

	// description explains what this evidence represents
	description?: string
}

// EvidenceType categorizes the kind of evidence. It remains an open enum:
// recommended values include artifact types already known to Gemara (e.g.
// EvaluationLog, EnforcementLog) plus categories for common evidence forms.
#EvidenceType: #ArtifactType | string @go(-)
