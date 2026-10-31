// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// EvaluationLog contains the results of evaluating a set of Layer 2 controls.
#EvaluationLog: {
	#Log
	metadata: type: "EvaluationLog"
	// result is the aggregate outcome across all evaluations in this log
	result: #Result
	evaluations: [#ControlEvaluation, ...#ControlEvaluation] @go(Evaluations,type=[]*ControlEvaluation)

	// Within one evidence list an id appears once: listing an item twice in the
	// same place says nothing. Every rule here skips an item with no id, which
	// the schema already reports where the id is missing.
	for i, evaluation in evaluations
	for j, assessment in evaluation."assessment-logs"
	if assessment.evidence != _|_ {
		let positions = {
			for k, item in assessment.evidence
			if item.id != _|_ {
				(item.id): "\(k)": k
			}
		}
		for id, ks in positions
		if len(ks) > 1 {
			_evidenceIDRepeatedInList: "\(i).\(j)": (id): error("evidence id \(id) appears more than once in one evidence list")
		}
	}

	// Across the log an evidence id names exactly one evidence item. The same
	// item may appear under its id in another assessment's evidence, which is how
	// one artifact supports several results; two different items may not share
	// an id. Each occurrence is compared with the first occurrence of its id,
	// so a log where many assessments cite one item validates in linear time.
	_evidenceByID: {
		for i, evaluation in evaluations
		for j, assessment in evaluation."assessment-logs"
		if assessment.evidence != _|_
		for k, item in assessment.evidence
		if item.id != _|_ {
			(item.id): "\(i).\(j).\(k)": item
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
	for evaluation in evaluations
	for assessment in evaluation."assessment-logs"
	if assessment.evidence != _|_
	for item in assessment.evidence
	if item.id != _|_
	if item.source != _|_ {
		_evidenceSourceDeclared: "evidence \(item.id) names a source that is not declared in metadata.mapping-references": true & _declaredReferenceIDs[item.source."reference-id"]
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
