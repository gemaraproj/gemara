// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

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

	// ---- Validation --------------------------------------------------------
	// A comment directly above a rule is listed under x-rules in the OpenAPI
	// projection. The comment above the actions redeclaration stays detached
	// (blank line after), so it is not added to that field's description.

	// A Blocked or Remediated action acted on something, so it names the findings
	// it acted on. An action with no effect may carry findings without an
	// exception (a shadow gate), findings with the exceptions that waived them (an
	// excepted finding), or nothing (a clean pass). An empty justification would
	// satisfy #Justification, whose every field is optional, while saying nothing,
	// so it is rejected.

	actions: [...#_ActionResultStrict]

	// Action ids are unique within the log, so an action can be referenced
	// unambiguously.
	_uniqueActionIds: {for i, a in actions {(a.id): i}}

	// Finding ids must be unique across the whole log, not just within one
	// action's justification: a reference to a finding id would otherwise be
	// ambiguous between actions. The value names the action position, so a
	// collision names both.
	_uniqueFindingIds: {
		for i, a in actions if a.justification != _|_ if a.justification.findings != _|_ for j, f in a.justification.findings {
			(f.id): "actions[\(i)].justification.findings[\(j)]"
		}
	}

	// Every finding's control and log name a declared mapping-reference. Unlike
	// other references in a log, neither may name this log itself: a finding's
	// control lives in a catalog, and its log is the evaluation log it came from.
	for i, a in actions if a.justification != _|_ if a.justification.findings != _|_ for j, f in a.justification.findings {
		let _declaredRefIds = [if metadata."mapping-references" != _|_ for r in metadata."mapping-references" {r.id}]
		_refValidation: "actions-\(i)-findings-\(j)-control": _declaredRefIds & list.Contains(f.control."reference-id")
		_refValidation: "actions-\(i)-findings-\(j)-log":     _declaredRefIds & list.Contains(f.log."reference-id")
	}
}

// ActionResult captures a performed enforcement action.
#ActionResult: {
	// id allows this entry to be referenced by other elements
	id: string

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

	// justification links the action to the findings that prompted it and any
	// applicable exceptions. A Blocked or Remediated action requires one that
	// names its findings; an action with no effect carries one only when it
	// reported findings or cited exceptions.
	justification?: #Justification @go(Justification,optional=nillable)
}

// EnforcementStep is a reference to the code that performed an enforcement action
#EnforcementStep: string @go(-)

// Justification provides the assessment data and exception references that justify an enforcement action.
#Justification: {
	// findings links the action to the assessment findings that justify it.
	findings?: [#AssessmentFinding, ...#AssessmentFinding] @go(Findings,type=[]*AssessmentFinding)

	// exceptions references the approved exceptions that authorise this action:
	// a subject that cannot conform has been granted one, so evaluation still
	// fails and enforcement stands down. Gemara does not model an exception, so
	// these name external artifacts through the log's mapping-references.
	exceptions?: [#ArtifactMapping, ...#ArtifactMapping] @go(Exceptions)
}

// AssessmentFinding is a finding an enforcement action acted on: a control
// failure surfaced by assessing it in an evaluation. It carries an id, a handle
// for referring to it within the log; a log and a control, because an
// enforcement action reacts to an evaluation and the control picks out that
// evaluation within the log; and a severity, because a finding a gate acted on
// is negative by nature. Unlike an audit finding, which may record conformity
// as readily as nonconformity, there is nothing to enforce against good news,
// and a waived Critical is not a waived Low. This is the grading of
// nonconformities that certification bodies work to under ISO/IEC 17021-1,
// which #Severity carries. Two actions may act on the same control failure, a
// block and then a remediation, so one log can hold findings with the same
// identity; they are the same finding.
#AssessmentFinding: {
	#Finding

	// id identifies this finding within the log, a handle its actions use to
	// refer to it.
	id: string

	// control names the control whose failure this finding records. It is
	// required here because it is the finding's subject, and it picks out the
	// evaluation in log.
	control: #EntryMapping

	// severity states how material the failure is. It is required because a
	// finding a gate acted on is negative by nature.
	severity: #Severity

	// log names the evaluation log this finding came from. The finding's control
	// identifies the evaluation within it, since a log evaluates each control
	// once, so no entry id is needed. There is no plan field beside it: the
	// evaluation's assessments name their own plans, and a second copy here could
	// only disagree.
	log: #ArtifactMapping
}

// Effect is what an enforcement action did to its target. It rolls up from an
// action to the log by the precedence Blocked > Remediated > Undetermined > None:
// the strongest effect dominates, because the log records what the enforcement
// layer did; how much that should worry anyone is the audit's question.
//
// An effect is not a finding's disposition, which says how the finding was
// settled. The two line up in documented ways rather than enforced ones: a
// block leaves its finding Open, since the gate prevented harm without fixing
// anything; a remediation normally leaves it Resolved; an action with no effect
// that cites an exception leaves it Waived; and one with no effect and no
// exception, a shadow gate, leaves it Open.
#Effect:
	// The action's effect could not be determined.
	"Undetermined" |
	// The action stopped the target, for example by blocking a merge.
	"Blocked" |
	// The action changed the target to correct a finding.
	"Remediated" |
	// The action did nothing to the target.
	"None" @go(-)

// #_ActionResultStrict carries the justification rules on #ActionResult. It is
// hidden from Go code generation (@go(-)), so the generated types stay shape
// only; cue2openapi lists its rules under x-rules. effect is redeclared here so
// the conditions resolve to the action rather than to the log's aggregate.
#_ActionResultStrict: {
	@go(-)
} & #ActionResult & {
	effect:         #Effect
	justification?: #Justification

	// A Blocked or Remediated action acted on something, so it names the findings
	// it acted on.
	if effect == "Blocked" || effect == "Remediated" {
		justification: #Justification

		if justification.findings == _|_ {
			justification: findings: error("a Blocked or Remediated action acted on findings: name the findings it acted on")
		}
	}

	// The error is a hidden field rather than justification: findings: error(...):
	// the rule reads justification, and writing to it from the same conditional is
	// a circular reference CUE rejects.

	// An action with no effect may leave its justification out, but a justification
	// it carries names findings or exceptions: an empty one justifies nothing.
	if effect == "None" || effect == "Undetermined" {
		if justification != _|_ if justification.findings == _|_ if justification.exceptions == _|_ {
			_emptyJustification: error("an empty justification justifies nothing: name findings or exceptions, or leave it out")
		}
	}
}
