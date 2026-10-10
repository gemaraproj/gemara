// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

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
		let _validCriteriaIds = [for c in criteria {c."reference-id"}]

		// Unify the valid ID list with a list.Contains constraint to require each result scores against declared criteria
		for i, r in results {
			_criteriaValidation: "\(i)": _validCriteriaIds & list.Contains(r."criteria-reference"."reference-id")
		}
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
	evidence?: [#_EvidenceStrict, ...#_EvidenceStrict]

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
	required: *false | bool @gemara(default=false)
}
