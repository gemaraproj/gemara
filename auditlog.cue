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

		// Within one evidence list an id appears once: listing an item twice in the
		// same place says nothing. Every rule here skips an item with no id, which
		// the schema already reports where the id is missing.
		for i, r in results
		if r.evidence != _|_ {
			let positions = {
				for k, item in r.evidence
				if item.id != _|_ {
					(item.id): "\(k)": k
				}
			}
			for id, ks in positions
			if len(ks) > 1 {
				_evidenceIDRepeatedInList: "\(i)": (id): error("evidence id \(id) appears more than once in one evidence list")
			}
		}

		// Across the log an evidence id names exactly one evidence item. The same
		// item may appear under its id in another result's evidence, which is how
		// one artifact supports several results; two different items may not share
		// an id. Each occurrence is compared with the first occurrence of its id,
		// so a log where many results cite one item validates in linear time.
		_evidenceByID: {
			for i, r in results
			if r.evidence != _|_
			for k, item in r.evidence
			if item.id != _|_ {
				(item.id): "\(i).\(k)": item
			}
		}
		for id, occurrences in _evidenceByID {
			let items = [for o in occurrences {o}]
			for n, item in items
			if n > 0 {
				_evidenceIDReuse: "evidence id \(id) names two different evidence items in this log": true & (item == items[0])
			}
		}

		// An evidence source names where the evidence came from, so its reference-id
		// must be a mapping reference declared in this log's metadata.
		//
		// This check and the one above are assignments rather than conditions. An
		// assignment involving an incomplete item stays incomplete and waits, while a
		// condition fails on it and hides every other error in the log.
		//
		// The declared ids are field names, so a mapping reference with no id leaves
		// them incomplete and the source checks wait until the id is supplied; the
		// missing id is reported where it is missing. They are read by iterating
		// metadata's fields, because selecting mapping-references directly fails as
		// a whole when any one entry is invalid, and every sourced item would then be
		// reported as undeclared.
		_declaredReferenceIDs: {for k, refs in metadata if k == "mapping-references" for r in refs {(r.id): true}}
		for r in results
		if r.evidence != _|_
		for item in r.evidence
		if item.id != _|_
		if item.source != _|_ {
			_evidenceSourceDeclared: "evidence \(item.id) names a source that is not declared in metadata.mapping-references": true & _declaredReferenceIDs[item.source."reference-id"]
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

// Retention records the resolved policy obligations governing referenced evidence.
// Each effective instant is a snapshot for this evidence at publication.
// Conflicting obligations are recorded side by side rather than reconciled by the producer.
#Retention: {
	// obligations records each retention rule that applies to this evidence and
	// the instant it resolves to
	obligations: [#RetentionObligationRef, ...#RetentionObligationRef]
}

// RetentionObligationRef ties a resolved instant to the obligation that produced it.
#RetentionObligationRef: {
	// obligation references a retention-obligations entry in the governing Policy
	obligation: #EntryMapping

	// effective is the instant this obligation resolves to for this evidence
	effective: #Datetime
}
