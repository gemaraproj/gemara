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
	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// Under a cited plan an AssessmentLog is one method's execution against one
	// requirement, so (this control evaluation, requirement, plan, method, executor)
	// is its key, where the plan and the method are each their reference-id and
	// entry-id together: two policies may each have a plan, or a method, with the
	// same entry-id. This is what lets a log hold several results for one
	// requirement — a plan requiring an automated probe and a manual review
	// produces two — and lets a citation name exactly one of them. Two records
	// that agree on every part of the key are rejected.
	// The executor is part of the key because a plan's method may name the executor
	// meant to run it, but the key uses the executor that did, and the two can
	// differ: one requirement run by the same method on two instances is two results.
	// A record with no plan is not keyed. It has no method for a citation to name,
	// so nothing could tell a second record from the first, and a producer with no
	// policy is not required to supply anything that would.

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

// ExecutionContext records how an assessment was configured: which method it
// ran. What the assessment is, its result, timing, steps, and the executor that
// produced it, stays at the top level of #AssessmentLog. Who performed the
// execution is executor, not part of its configuration. It is one field rather
// than several so that a stable type does not gain a permanent top-level field per
// configuration input; further inputs land here additively, where a top-level field
// could never be moved.
#ExecutionContext: {
	// method names the accepted method this assessment executed. Under a cited plan it
	// is one of that plan's evaluation-methods, and its reference-id is the plan's.
	// Without a plan its reference-id names the policy, and its entry-id one of that
	// policy's own evaluation-methods, which are the fallback where no plan is bound.
	method?: #EntryMapping & {"entry-id": string & strings.MinRunes(1)} @go(Method,optional=nillable)
}

// AssessmentLog contains the results of executing a single assessment procedure for a control requirement.
#AssessmentLog: {
	// Requirement should map to the assessment requirement for this assessment.
	requirement: #EntryMapping
	// Plan maps to the policy assessment plan being executed.
	plan?: #EntryMapping @go(Plan,optional=nillable)
	// execution-context records how this assessment was configured. A cited plan obliges a
	// method of that plan: it lists the methods it accepts, so a record of running under
	// it has to say which one ran.
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

	// A cited plan obliges a method of that plan: the plan lists the methods it accepts,
	// and a record of running under it that does not say which one ran leaves two
	// results against the same plan indistinguishable. The method's reference-id is
	// the plan's, the way a requirement's is its control's. Stated on the field it
	// requires — forbidding a field instead puts bottom in that field's type, and the
	// Go projection degrades it to an unusable `any`.
	//
	// There is no rule the other way. A method without a plan is the ordinary case for
	// a producer running the policy's fallback method, and a producer that knows of no
	// policy omits it.

	if plan != _|_ {"execution-context": method: {"reference-id": plan."reference-id", "entry-id": string}}

}

#AssessmentStep: string @go(-)

#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)
