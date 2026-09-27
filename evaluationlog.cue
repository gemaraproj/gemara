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

	// Items are grouped by id, keyed by where each sits in the log, and each
	// occurrence is compared with the first occurrence of its id, so a log where
	// many assessments cite one item validates in linear time.

	// Across the log an evidence id names exactly one evidence item. The same
	// item may appear under its id in another assessment's evidence, which is how
	// one artifact supports several results; two different items may not share
	// an id.
	for id, occurrences in {
		for i, evaluation in evaluations
		for j, assessment in evaluation."assessment-logs"
		if assessment.evidence != _|_
		for k, item in assessment.evidence
		if item.id != _|_ {
			(item.id): "evaluations.\(i).assessment-logs.\(j).evidence.\(k)": item
		}
	} {
		let at = [for o, _ in occurrences {o}]
		let items = [for o in occurrences {o}]
		for n, item in items
		if n > 0 {
			_evidenceIDReuse: "evidence id \(id) names two different evidence items in this log: \(at[n]) differs from \(at[0])": true & (item == items[0])
		}
	}

	// The source check below and the reuse check above are assignments rather
	// than conditions. An assignment involving an incomplete item stays
	// incomplete and waits, while a condition fails on it and hides every other
	// error in the log.
	//
	// The declared ids are field names, so a mapping reference with no id leaves
	// them incomplete and the source checks wait until the id is supplied; the
	// missing id is reported where it is missing. They are read by iterating
	// metadata's fields, because selecting mapping-references directly fails as
	// a whole when any one entry is invalid, and every sourced item would then be
	// reported as undeclared.

	// The mapping-reference ids declared in this log's metadata, which an
	// evidence source's reference-id must name.
	_declaredReferenceIDs: {for k, refs in metadata if k == "mapping-references" for r in refs {(r.id): true}}

	// An evidence source names where the evidence came from, so its reference-id
	// must be a mapping reference declared in this log's metadata.
	for evaluation in evaluations
	for assessment in evaluation."assessment-logs"
	if assessment.evidence != _|_
	for item in assessment.evidence
	if item.id != _|_
	if item.source != _|_ {
		_evidenceSourceDeclared: "evidence \(item.id) names a source that is not declared in metadata.mapping-references": true & _declaredReferenceIDs[item.source."reference-id"]
	}

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// An AssessmentLog is one method's execution against one requirement under one
	// plan, so (this control evaluation, requirement, plan-id, method-id) is its key.
	// This is what lets a log hold several results for one requirement — a plan
	// requiring an automated probe and a manual review produces two — and what stops
	// it holding two that nothing tells apart. An assessment citing no plan has no
	// method to select, so the requirement alone must be unique within its control
	// evaluation: attribution by executor does not make two such records distinct.

	for i, e in evaluations {
		_uniqueAssessments: "\(i)": {
			for j, a in e."assessment-logs" {
				let _plan = [if a.plan != _|_ {a.plan."entry-id"}, "-"][0]
				let _method = [if a["execution-context"] != _|_ if a["execution-context"]["method-id"] != _|_ {a["execution-context"]."method-id"}, "-"][0]
				"\(a.requirement."entry-id")/\(_plan)/\(_method)": j
			}
		}
	}

}

// ControlEvaluation contains the results of evaluating a single Layer 5 control.
#ControlEvaluation: {
	name:    string
	result:  #Result
	message: string
	control: #EntryMapping
	"assessment-logs": [#AssessmentLog, ...#AssessmentLog] @go(AssessmentLogs,type=[]*AssessmentLog)
	"assessment-logs": [...{
		// each assessment log's requirement.reference-id must match the control's reference-id, and is taken from the control when left out
		requirement: "reference-id": (control."reference-id")
	}]
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

// ExecutionContext records how an assessment was configured: which method it
// ran, and the values it ran with. What the assessment is, its result, timing, steps,
// and the executor that produced it, stays at the top level of #AssessmentLog.
// Who performed the execution is executor, not part of its configuration. It is one
// field rather than several so that a stable type does not gain a permanent
// top-level field per configuration input; further inputs land here additively, where
// a top-level field could never be moved.
//
#ExecutionContext: {
	// method-id names the accepted method this assessment executed, where the policy
	// prescribed one: a method of the cited plan, or one of the policy's own
	// evaluation-methods, which are the fallback where no plan is bound.
	"method-id"?: string @go(MethodId)
}

// AssessmentLog contains the results of executing a single assessment procedure for a control requirement.
// An assessment log MUST include start unless its result is Not Run, Unknown or Not Applicable.
#AssessmentLog: {
	// Requirement should map to the assessment requirement for this assessment.
	requirement: #EntryMapping
	// Plan maps to the policy assessment plan being executed.
	plan?: #EntryMapping @go(Plan,optional=nillable)
	// execution-context records how this assessment was configured. A cited plan obliges a
	// method: the plan lists the methods it accepts, so a record of running under it
	// has to say which one ran.
	"execution-context"?: #ExecutionContext @go(ExecutionContext,optional=nillable)
	// executor is the actor that performed this assessment: the tool instance,
	// service or person that produced this result when it is not the author of the log.
	executor?: #Actor @go(Executor,optional=nillable)
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

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// A cited plan obliges a method: the plan lists the methods it accepts, and a
	// record of running under it that does not say which one ran leaves two results
	// against the same plan indistinguishable. Stated on the field it requires —
	// forbidding a field instead puts bottom in that field's type, and the Go
	// projection degrades it to an unusable `any`.
	//
	// There is no rule the other way. Settings without a plan is the ordinary case
	// for a producer running the policy's fallback method, or for one that knows of
	// no policy at all.

	if plan != _|_ {"execution-context": "method-id": string}

}

#AssessmentStep: string @go(-)

#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)
