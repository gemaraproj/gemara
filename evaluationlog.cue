// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

import "strings"

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
	// A comment directly above a rule is listed under x-rules in the OpenAPI
	// projection.

	// Under a cited plan, an AssessmentLog is keyed by (this control evaluation,
	// requirement, plan, method, executor). The plan and the method are each their
	// reference-id and entry-id together, because two policies may each have a
	// plan or a method with the same entry-id; the executor is its id. The key lets
	// a log hold several results for one requirement: a plan with an automated
	// probe and a manual review produces two. A method does not fix who runs it, so
	// one method run on two instances is also two results, even when the plan's
	// method names the executor meant to run it. Two records that agree on every
	// part of the key are rejected, so a log holds one run of a method by an
	// executor. Records with no plan are not keyed, so this rule does not affect
	// logs that cite no plans.
	for i, e in evaluations {
		_uniqueAssessments: "\(i)": {
			for j, a in e."assessment-logs"
			if a.plan != _|_ {
				let _methodRef = [if a["execution-context"] != _|_ if a["execution-context"].method != _|_ {"present:\(a["execution-context"].method."reference-id")"}, "absent"][0]
				let _method = [if a["execution-context"] != _|_ if a["execution-context"].method != _|_ {"present:\(a["execution-context"].method."entry-id")"}, "absent"][0]
				let _executor = [if a.executor != _|_ {"present:\(a.executor.id)"}, "absent"][0]
				"\(a.requirement."entry-id")": "\(a.plan."reference-id")": "\(a.plan."entry-id")": "\(_methodRef)": "\(_method)": "\(_executor)": j
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

// ExecutionContext records what an assessment took from a policy: the accepted
// method it ran. The executor, the result and everything else about the
// assessment stay at the top level of AssessmentLog, because they mean something
// without a policy. Policy inputs share this one field so that a stable type does
// not gain a top-level field for each of them, and a field placed at the top
// level could not later be grouped with the rest. Further policy inputs, such as
// the values given for a plan's parameters, belong here too.
#ExecutionContext: {
	// method names the accepted method this assessment ran. Under a cited plan it
	// is one of that plan's evaluation-methods, and its reference-id is the plan's.
	// Without a plan, its reference-id names a policy and its entry-id one of that
	// policy's own evaluation-methods.
	method?: #EntryMapping & {"entry-id": string & strings.MinRunes(1)} @go(Method,optional=nillable)
}

// AssessmentLog contains the results of executing a single assessment procedure for a control requirement.
// An assessment log MUST include start unless its result is Not Run, Unknown or Not Applicable.
#AssessmentLog: {
	// Requirement should map to the assessment requirement for this assessment.
	requirement: #EntryMapping
	// Plan maps to the policy assessment plan being executed.
	plan?: #EntryMapping @go(Plan,optional=nillable)
	// execution-context records what this assessment took from a policy. A record
	// that cites a plan names here which of the plan's methods it ran.
	"execution-context"?: #ExecutionContext @go(ExecutionContext,optional=nillable)
	// executor is the actor that performed this assessment: the tool instance,
	// service or person that produced this result. When it is omitted, the log's
	// metadata.author performed it. A plan's accepted method may name the executor
	// meant to run it; this field names the one that did. A record whose executor
	// differs from the method's is still valid, and an audit reports the mismatch.
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
	// A comment directly above a rule is listed under x-rules in the OpenAPI
	// projection.

	// The rule below compares the method's reference-id with `==` rather than
	// unifying it: unifying would fill an omitted reference-id in from the plan,
	// so the document a producer wrote would lack a field every non-CUE consumer
	// requires. It requires a method when a plan is present rather than forbidding
	// the plan when the method is absent: a conditional that forbids a field puts
	// bottom in that field's type, and the Go projection turns the field into an
	// unusable `any`.

	// A record that cites a plan names the method it ran, so that the record can
	// be checked against the methods the plan accepts. The method's reference-id
	// must equal the plan's. Unlike a requirement's reference-id, which is taken
	// from its control when left out, it must be written. A method without a plan
	// is allowed: it names one of a policy's own evaluation-methods, and a producer
	// that knows of no policy omits it.
	if plan != _|_ {"execution-context": method: {"reference-id": ==plan."reference-id", "entry-id": string}}

}

#AssessmentStep: string @go(-)

#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)
