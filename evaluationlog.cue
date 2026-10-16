// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

import "list"

@go(gemara)

// EvaluationLog contains the results of evaluating a set of Layer 2 controls.
#EvaluationLog: {
	#Log
	metadata: type: "EvaluationLog"
	// result is the aggregate outcome across all evaluations in this log
	result: #Result
	evaluations: [#ControlEvaluation, ...#ControlEvaluation] @go(Evaluations,type=[]*ControlEvaluation)

	// Within one evidence list an id appears once: listing an item twice in the
	// same place says nothing.
	for i, evaluation in evaluations
	for j, assessment in evaluation."assessment-logs"
	if assessment.evidence != _|_ {
		_uniqueEvidenceIDsInList: "\(i).\(j)": {for k, item in assessment.evidence {(item.id): k}}
	}

	// Across the log an evidence id names exactly one evidence item. The same
	// item may appear under its id in another assessment's evidence, which is how
	// one artifact supports several results; two different items may not share
	// an id.
	_evidenceByID: {
		for i, evaluation in evaluations
		for j, assessment in evaluation."assessment-logs"
		if assessment.evidence != _|_
		for k, item in assessment.evidence {
			(item.id): "\(i).\(j).\(k)": item
		}
	}
	for id, occurrences in _evidenceByID
	for p, first in occurrences
	for q, second in occurrences
	if p < q
	if first != second {
		_evidenceIDReuse: (id): error("evidence id \(id) names two different evidence items in this log")
	}

	// An evidence source names where the evidence came from, so its reference-id
	// must be a mapping reference declared in this log's metadata. The check is an
	// assignment rather than a condition, because an assignment tolerates an
	// item that is incomplete and a condition fails on one, hiding what is missing.
	_declaredReferenceIDs: [if metadata."mapping-references" != _|_ for r in metadata."mapping-references" {r.id}]
	for evaluation in evaluations
	for assessment in evaluation."assessment-logs"
	if assessment.evidence != _|_
	for item in assessment.evidence
	if item.source != _|_ {
		_evidenceSourceDeclared: "evidence \(item.id) names a source that is not declared in metadata.mapping-references": true & list.Contains(_declaredReferenceIDs, item.source."reference-id")
	}

	// Evidence that omits digest must be located by a url that fixes its bytes
	// (see EvidenceMapping.digest).
	_ociDigestCheck: _ociDigestRule & {
		references: [if metadata."mapping-references" != _|_ for r in metadata."mapping-references" {r}]
		evidence: [for e in _evidenceByID for item in e {item}]
	}
}

// ControlEvaluation contains the results of evaluating a single Layer 5 control.
#ControlEvaluation: {
	name:    string
	result:  #Result
	message: string
	control: #EntryMapping
	"assessment-logs": [#AssessmentLog, ...#AssessmentLog] @gemara(projectable=false) @go(AssessmentLogs,type=[]*AssessmentLog)
	// Enforce that control reference and the assessments' references match
	// This formulation uses the control's reference if the assessment doesn't include a reference
	"assessment-logs": [...{
		requirement: "reference-id": (control."reference-id")
	}]
	// Require start timestamp on assessments that actually executed
	"assessment-logs": [#_AssessmentLogStrict, ...#_AssessmentLogStrict]
}

// _AssessmentLogStrict layers the "start required unless unexecuted" rule on top of #AssessmentLog
#_AssessmentLogStrict: {
	@go(-)
} & #AssessmentLog & {
	result: #Result
	if result != "Not Run" && result != "Unknown" && result != "Not Applicable" {
		start: #Datetime
	}
}

// AssessmentLog contains the results of executing a single assessment procedure for a control requirement.
#AssessmentLog: {
	// Requirement should map to the assessment requirement for this assessment.
	requirement: #EntryMapping
	// Plan maps to the policy assessment plan being executed.
	plan?: #EntryMapping @go(Plan,optional=nillable)
	// Description provides a summary of the assessment procedure.
	description: string
	// Result is the overall outcome of the assessment procedure, matching the result of the last step that was run.
	result: #Result
	// Message provides additional context about the assessment result.
	message: string
	// Applicability is elevated from the Layer 2 Assessment Requirement to aid in execution and reporting.
	applicability: [string, ...string] @go(Applicability,type=[]string)
	// Steps are sequential actions taken as part of the assessment, which may halt the assessment if a failure occurs.
	steps: [#AssessmentStep, ...#AssessmentStep]
	// Steps-executed is the number of steps that were executed as part of the assessment.
	"steps-executed"?: int @go(StepsExecuted)
	// Start is the timestamp when the assessment began.
	// Assessments that never executed have no start time to record.
	start?: #Datetime

	// End is the timestamp when the assessment concluded.
	end?: #Datetime
	// Recommendation provides guidance on how to address a failed assessment.
	recommendation?: string
	// ConfidenceLevel indicates the evaluator's confidence level in this specific assessment result.
	"confidence-level"?: #ConfidenceLevel @go(ConfidenceLevel)
	// Evidence records the raw data cited to support this assessment's opinion.
	evidence?: [#Evidence, ...#Evidence] @go(Evidence)
	evidence?: [#_EvidenceStrict, ...#_EvidenceStrict]
}

#AssessmentStep: string @go(-)

#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)
