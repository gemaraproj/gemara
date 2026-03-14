// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// EnforcementLog records actions taken in response to noncompliance findings from Layer 5 evaluations.
#EnforcementLog: {
	// metadata provides detailed data about this log
	metadata: #Metadata @go(Metadata)
	// disposition is the aggregate enforcement disposition across all actions in this log
	disposition: #Disposition
	// target references the resource enforcement was performed on
	target: #Resource @go(Target)
	// actions is the list of enforcement actions performed
	actions: [#ActionLog, ...#ActionLog] @go(Actions,type=[]*ActionLog)
	// Enforce that Clear dispositions only contain Passed assessment results
	actions: [...{
		if disposition == "Clear" {
			justification: assessments: [...{result: "Passed"}]
		}
	}]
}

// ActionLog captures a performed enforcement action.
#ActionLog: {
	// disposition is the enforcement action taken
	disposition: #Disposition @go(Disposition)

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

// Disposition enumerates the possible enforcement outcomes.
#Disposition:
	// Findings existed and actions were taken.
	"Enforced" |
	// Findings existed but were accepted without action.
	"Tolerated" |
	// No findings, nothing to act on.
	"Clear" @go(-)
