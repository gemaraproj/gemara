// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// Policy represents a policy document with metadata, contacts, scope, imports, implementation plan, risks, and adherence requirements.
#Policy: {
	title:    string
	metadata: #Metadata
	metadata: type: "Policy"
	contacts:               #RACI
	scope:                  #Scope
	imports:                #Imports
	"implementation-plan"?: #ImplementationPlan @go(ImplementationPlan)
	risks?:                 #Risks
	adherence:              #Adherence
}

// Scope defines what is included and excluded from policy applicability.
#Scope: {
	in:   #Dimensions
	out?: #Dimensions
}

// Dimensions specify the applicability criteria for a policy
#Dimensions: {
	// technologies is an optional list of technology categories or services
	technologies?: [string, ...string]
	// geopolitical is an optional list of geopolitical regions
	geopolitical?: [string, ...string]
	// sensitivity is an optional list of data classification levels
	sensitivity?: [string, ...string]
	// users is an optional list of user roles
	users?: [string, ...string]
	groups?: [string, ...string]
}

// Imports defines external policies, controls, and guidelines required by this policy.
#Imports: {
	policies?: [#ArtifactMapping, ...#ArtifactMapping]
	catalogs?: [#CatalogImport, ...#CatalogImport]
	guidance?: [#GuidanceImport, ...#GuidanceImport]
}

// ImplementationPlan defines when and how the policy becomes active.
#ImplementationPlan: {
	"notification-process"?: string                 @go(NotificationProcess)
	"evaluation-timeline":   #ImplementationDetails @go(EvaluationTimeline)
	"enforcement-timeline":  #ImplementationDetails @go(EnforcementTimeline)
}

// ImplementationDetails specifies the timeline for policy implementation.
#ImplementationDetails: {
	start: #Datetime
	end?:  #Datetime
	notes: string
}

// Risks defines mitigated and accepted risks addressed by this policy.
#Risks: {
	// Mitigated risks only need reference-id and risk-id (no justification required)
	mitigated?: [#MitigatedRisk, ...#MitigatedRisk]
	// Accepted risks require rationale (justification) and may include scope. Controls addressing these risks are implicitly identified through threat mappings.
	accepted?: [#AcceptedRisk, ...#AcceptedRisk]

	// when mitigated risks are listed, each mitigated risk id must be unique
	if mitigated != _|_ {
		_uniqueMitigatedRiskIds: {for i, m in mitigated {(m.id): i}}
	}

	// when accepted risks are listed, each accepted risk id must be unique, and when target-id is set, it must match an entry in risks.mitigated
	if accepted != _|_ {
		_uniqueAcceptedRiskIds: {for i, a in accepted {(a.id): i}}

		let _validMitigatedIds = [if mitigated != _|_ for m in mitigated {m.id}]
		for i, a in accepted {
			if a."target-id" != _|_ {
				_targetValidation: "\(i)": _validMitigatedIds & list.Contains(a."target-id")
			}
		}
	}
}

// MitigatedRisk represents a risk addressed by the policy
#MitigatedRisk: {
	// id uniquely identifies this mitigated risk entry so it can be referenced by accepted risks
	id: string

	// risk references the risk being mitigated
	risk: #EntryMapping
}

// AcceptedRisk documents a risk the organization has chosen to accept,
// optionally linking it to a mitigated risk when the acceptance covers
// residual risk after partial mitigation.
#AcceptedRisk: {
	// id uniquely identifies this accepted risk entry
	id: string

	// target-id optionally links this acceptance to a mitigated risk entry; when set it must match the id of an entry in risks.mitigated
	"target-id"?: string

	// risk references the risk being accepted
	risk: #EntryMapping

	// scope defines where the risk acceptance applies
	scope?: #Scope

	// justification explains why the risk is accepted
	justification?: string
}

// Adherence defines evaluation methods, assessment plans, enforcement methods, retention obligations, and non-compliance notifications.
#Adherence: {
	"evaluation-methods"?: [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)
	AP="assessment-plans"?: [#AssessmentPlan, ...#AssessmentPlan] @go(AssessmentPlans)
	"enforcement-methods"?: [#AcceptedMethod & {type: #EnforcementMethodType}, ...#AcceptedMethod & {type: #EnforcementMethodType}] @go(EnforcementMethods)

	// retention-anchors declares the events, other than collected-at, from which retention obligations measure their duration
	RA="retention-anchors"?: [#RetentionAnchor, ...#RetentionAnchor] @go(RetentionAnchors)

	// retention-obligations lists the preservation and disposal requirements this policy adopts from external records schedules
	RO="retention-obligations"?: [#RetentionObligation, ...#RetentionObligation] @go(RetentionObligations)

	"non-compliance"?: string @go(NonCompliance)

	// retention anchor ids must be unique, and none may redeclare collected-at
	if RA != _|_ {
		// collected-at is pre-seeded so that a declaration shadowing it collides like a duplicate id
		_uniqueRetentionAnchorIds: {(#CollectedAtAnchor): -1, for i, a in RA {(a.id): i}}
	}

	// retention obligation ids must be unique, each obligation's anchor must be collected-at or a declared retention anchor, and an obligation measured from collected-at must state a duration
	if RO != _|_ {
		_uniqueRetentionObligationIds: {for i, o in RO {(o.id): i}}

		let _declaredAnchors = [if RA != _|_ for a in RA {a.id}]
		let _validAnchorIds = list.Concat([[#CollectedAtAnchor], _declaredAnchors])

		// Unify the valid anchor list with a list.Contains constraint to require each obligation to measure from collected-at or a declared anchor
		for i, o in RO {
			_anchorValidation: "\(i)": _validAnchorIds & list.Contains(o.anchor)

			// An obligation measured from collected-at must state its period; only a declared
			// anchor may stand without one, meaning the obligation runs to the anchor itself
			if o.anchor == #CollectedAtAnchor if o.duration == _|_ {
				_collectedAtWithoutDuration: "\(o.id)": error("retention obligation \(o.id) is measured from collected-at and must state a duration")
			}
		}
	}

	// An assessment log names the plan it ran under by id, so a plan id must name one
	// plan within a policy.
	if AP != _|_ {
		_uniquePlanIds: {for i, p in AP {"\(p.id)": i}}
	}
}

// AssessmentPlan defines how a specific assessment requirement is evaluated.
#AssessmentPlan: {
	id:               string
	"requirement-id": string @go(RequirementId)
	frequency:        string
	EM="evaluation-methods": [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)
	"evidence-requirements"?: string @go(EvidenceRequirements)
	parameters?: [#Parameter, ...#Parameter]

	// conflict-resolution states how disagreeing results are resolved for this requirement.
	// A plan naming more than one evaluation method MUST state one, because a multi-source
	// evaluation with no stated resolution rule produces a result whose value depends on
	// which log a consumer happened to read first. Stating it here rather than leaving it
	// to each implementation is the point: a default that lives in an implementation is one
	// every implementer picks differently.
	CR="conflict-resolution"?: #ConflictResolution @go(ConflictResolution)

	// A plan with more than one evaluation method must state conflict-resolution.
	matchN(>=1, [
		{"conflict-resolution"!: #ConflictResolution},
		{"evaluation-methods": [_]},
	])

	// highest-rank needs a total order, so every method on the plan must carry a rank.
	// The list names any method that does not.
	if CR != _|_ if CR == "highest-rank" {
		_methodsWithoutRank: [for m in EM if m.rank == _|_ {m.id}] & []
	}

	// unanimous is decided by the required methods, so the plan must mark at least one.
	// Unanimity over no methods would pass every requirement without evidence.
	if CR != _|_ if CR == "unanimous" {
		_requiredMethods: [for m in EM if m.required {m.id}] & [_, ...]
	}

	// Method ranks within one assessment plan must be unique, so that a rank-based
	// conflict resolution has a total order to work with.
	_uniqueRanks: {for i, m in EM if m.rank != _|_ {"\(m.rank)": i}}

	// A log that names the method that produced it does so by id (an assessment log has
	// no such field yet, see gemaraproj/gemara#485), so a method id must name one method
	// within its plan.
	_uniqueMethodIds: {for i, m in EM {"\(m.id)": i}}
}

// AcceptedMethod defines a method for evaluation or enforcement.
#AcceptedMethod: {
	id:           string
	type:         #MethodType
	mode:         #ModeType
	required:     *false | bool @gemara(default=false)
	description?: string
	executor?:    #Actor

	// rank orders this method against the others on the same assessment plan; lower is
	// higher precedence. It is required on every method when the plan's conflict-resolution
	// is highest-rank, and optional otherwise, because a plan with one method needs none.
	rank?: int & >=1
}

// ConflictResolution states how disagreeing evaluation results for one requirement are
// resolved. Methods conflict when their assessment logs for the requirement carry different
// results. Only executed results take part: a log whose result is "Not Run", "Unknown" or
// "Not Applicable" (the results that need no start time) is recorded but does not decide.
// Under every option, methods that agree resolve to the result they agree on, and when no
// method that decides has an executed result (for unanimous, no required method), the
// resolved result is "Needs Review", so missing evidence is never read as a pass.
//
// highest-rank and unanimous decide by method, so each method counts once, by its own
// result: the result of its executed log with the latest start, or "Needs Review" when
// the logs at that start disagree. Both strategies need an assessment
// log to say which method produced it. A log has no such field yet (see
// gemaraproj/gemara#485), so these two cannot be applied from log data alone until it
// has; most-recent and escalate do not depend on it.
//
//   - highest-rank: the result from the method with the lowest rank wins. Every method on
//     the plan must carry a rank, and ranks are unique.
//   - unanimous: only required methods decide. If they all returned the same result, that
//     is the resolved result; if any two disagree, the resolved result is "Needs Review".
//     Optional methods are recorded but cannot block. The plan must mark at least one
//     method as required.
//   - most-recent: the result from the assessment log with the latest start wins. The order
//     is by the assessment's start rather than the log's metadata.date, because start is
//     when the evidence was observed, while metadata.date moves whenever a log is
//     republished without being re-run. start is required on every executed assessment log,
//     so the order is always defined. Logs at the latest start that disagree
//     resolve to "Needs Review".
//   - escalate: no automatic resolution. Any conflict resolves to "Needs Review" for a
//     person to settle.
#ConflictResolution: "highest-rank" | "unanimous" | "most-recent" | "escalate" @go(-)

#ModeType:              "Manual" | "Automated"                           @go(-)
#MethodType:            "Behavioral" | "Intent" | "Remediation" | "Gate" @go(-)
#EvaluationMethodType:  "Intent" | "Behavioral"                          @go(-)
#EnforcementMethodType: "Gate" | "Remediation"                           @go(-)

// Parameter defines a configurable parameter for assessment or enforcement activities.
#Parameter: {
	id:          string
	label:       string
	description: string
	"accepted-values"?: [string, ...string] @go(AcceptedValues)
}

// GuidanceImport defines how to import guidance documents with optional exclusions and constraints.
#GuidanceImport: {
	"reference-id": string @go(ReferenceId)
	exclusions?: [string, ...string]
	// Constraints allow policy authors to define ad hoc minimum requirements (e.g., "review at least annually").
	constraints?: [#Constraint, ...#Constraint]
}

// CatalogImport defines how to import control catalogs with optional exclusions and constraints.
#CatalogImport: {
	"reference-id": string @go(ReferenceId)
	// exclusions lists the control and assessment requirement IDs from the imported catalog that this policy leaves out
	exclusions?: [string, ...string]
	constraints?: [#Constraint, ...#Constraint]
	// applicability selects the applicability groups to adopt from this imported catalog. Each ID must match
	// an entry in the catalog's metadata."applicability-groups". A requirement is in play when any of its
	// applicability IDs is selected.
	"applicability"?: [string, ...string] @go(Applicability)
}

// Constraint defines a prescriptive requirement that applies to a specific guidance or control.
#Constraint: {
	// Unique ID for this constraint to enable Layer 5/6 tracking
	id: string
	// Links to the specific Guidance or Control being constrained
	"target-id": string @go(TargetId)
	// The prescriptive requirement/constraint text
	text: string
}

// RetentionDuration is a records-retention period: an ISO 8601 duration limited to date
// components (e.g. P7Y, P18M, P1Y2M3W4D). Records schedules do not express retention below
// day precision, so a sub-day period such as PT1H is rejected. Because of that it is not a
// general-purpose duration, and frequency or freshness windows need a type of their own.
#RetentionDuration: =~"^P(?:\\d+Y(?:\\d+M)?(?:\\d+W)?(?:\\d+D)?|\\d+M(?:\\d+W)?(?:\\d+D)?|\\d+W(?:\\d+D)?|\\d+D)$" @go(RetentionDuration)

// CollectedAtAnchor is the retention anchor Gemara resolves from the evidence itself,
// so it is understood without a declaration and a policy may not redeclare it.
#CollectedAtAnchor: "collected-at" @go(-)

// RetentionAnchor declares an event from which a retention duration is measured.
// Policies declare the anchors they use rather than drawing on a fixed vocabulary,
// because records schedules anchor to events Gemara cannot enumerate.
#RetentionAnchor: {
	// id is referenced by a retention obligation's anchor field; it must not be collected-at
	id: string

	// title names the anchoring event at a glance
	title: string

	// description explains when the event occurs and who determines it
	description: string

	// defined-by cites the schedule or record that defines the event
	"defined-by"?: #EntryMapping @go(DefinedBy,optional=nillable)
}

// RetentionKind states whether an obligation sets a floor or a ceiling.
#RetentionKind: "minimum" | "maximum" @go(-)

// RetentionObligation is a typed preservation or disposal requirement adopted
// from an external records schedule.
#RetentionObligation: {
	// id allows evidence citations to reference this obligation
	id: string

	// kind states whether this obligation sets a floor or a ceiling
	kind: #RetentionKind

	// implements cites the entry in the external schedule that this obligation puts into
	// effect, in the sense the implements relationship has in a mapping document
	implements: #EntryMapping

	// anchor names the event from which duration is measured. collected-at is
	// resolved by Gemara from the evidence itself; any other value must match a
	// declared retention-anchors id, and duration is then advisory, with the
	// citation's effective instant being authoritative.
	anchor: #CollectedAtAnchor | string

	// duration is the period measured from the anchor. Omit it when the obligation runs to
	// the anchor itself, such as keeping records until a support period ends, rather than
	// writing a zero period. A zero period stays valid, since a ceiling of P0D from
	// collected-at means the evidence is disposed of when it is collected. An obligation
	// measured from collected-at always states a duration.
	duration?: #RetentionDuration

	// applies-to optionally narrows this obligation to specific assessment requirements
	"applies-to"?: [string, ...string] @go(AppliesTo)
}
