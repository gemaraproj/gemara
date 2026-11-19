// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

@go(gemara)

// EnforcementLog records actions taken in response to noncompliance findings from Layer 5 evaluations.
#EnforcementLog: {
	#Log
	metadata: type: "EnforcementLog"
	// effect is the aggregate across this log's actions, by the precedence
	// documented on #Effect.
	effect: #Effect
	// actions is the list of enforcement actions performed
	actions: [#ActionResult, ...#ActionResult] @go(Actions,type=[]*ActionResult)
}

// ActionResult captures a performed enforcement action.
#ActionResult: {
	// effect is what this action did to the target.
	effect: #Effect @go(Effect)

	// method references the specific AcceptedMethod entry within the Policy being enforced
	method: #EntryMapping @go(Method)

	// message provides additional context about the action
	message?: string @go(Message,type=*string)

	// start is the timestamp when the enforcement action began
	start: #Datetime

	// end is the timestamp when the enforcement action concluded
	end?: #Datetime

	// steps references the code paths or addresses that carried out this enforcement action
	steps: [#EnforcementStep, ...#EnforcementStep]

	// justification links the action to its assessment findings and any applicable exceptions
	justification: #Justification @go(Justification)
}

// EnforcementStep is a reference to the code that performed an enforcement action
#EnforcementStep: string @go(-)

// Justification provides the assessment data and exception references that justify an enforcement action.
#Justification: {
	// assessments links the action to one or more Assessment Findings
	assessments: [#AssessmentFinding, ...#AssessmentFinding] @go(Assessments)

	// exceptions references approved Policy exceptions that authorize the action
	exceptions?: [#ArtifactMapping, ...#ArtifactMapping] @go(Exceptions)
}

// AssessmentFinding maps an enforcement action to its originating assessment data across Layer 2, Layer 3, and Layer 5.
#AssessmentFinding: {
	// result is the assessment outcome that triggered the enforcement action
	result: #Result

	// requirement maps to the Layer 2 assessment requirement that was evaluated
	requirement?: #EntryMapping @go(Requirement)

	// plan maps to the Policy assessment plan that was executed
	plan?: #EntryMapping @go(Plan)

	// log maps to the EvaluationLog entry containing the finding
	log: #EntryMapping @go(Log)
}

// Effect is what an enforcement action did to its target. It rolls up from an
// action to the log by the precedence Blocked > Remediated > Undetermined > None:
// the strongest effect dominates, because the log records what the enforcement
// layer did; how much that should worry anyone is the audit's question.
#Effect:
	// The action's effect could not be determined.
	"Undetermined" |
	// The action stopped the target, for example by blocking a merge.
	"Blocked" |
	// The action changed the target to correct a finding.
	"Remediated" |
	// The action did nothing to the target.
	"None" @go(-)
