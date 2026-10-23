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
}

// MitigatedRisk represents a risk addressed by the policy
#MitigatedRisk: {
	// id allows this mitigated risk entry to be referenced by accepted risks
	id: string

	// risk references the risk being mitigated
	risk: #EntryMapping
}

// AcceptedRisk documents a risk the organization has chosen to accept,
// optionally linking it to a mitigated risk when the acceptance covers
// residual risk after partial mitigation.
#AcceptedRisk: {
	// id allows this accepted risk entry to be referenced
	id: string

	// target-id optionally links this acceptance to a mitigated risk entry
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
	"assessment-plans"?: [#AssessmentPlan, ...#AssessmentPlan] @go(AssessmentPlans)
	"enforcement-methods"?: [#AcceptedMethod & {type: #EnforcementMethodType}, ...#AcceptedMethod & {type: #EnforcementMethodType}] @go(EnforcementMethods)

	// retention-anchors declares the events, other than collected-at, from which retention obligations measure their duration
	RA="retention-anchors"?: [#RetentionAnchor, ...#RetentionAnchor] @go(RetentionAnchors)

	// retention-obligations lists the preservation and disposal requirements this policy adopts from external records schedules
	RO="retention-obligations"?: [#RetentionObligation, ...#RetentionObligation] @go(RetentionObligations)

	"non-compliance"?: string @go(NonCompliance)

	if RA != _|_ {
		// collected-at is pre-seeded so that a declaration shadowing it collides like a duplicate id
		_uniqueRetentionAnchorIds: {(#CollectedAtAnchor): -1, for i, a in RA {(a.id): i}}
	}

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
}

// AssessmentPlan defines how a specific assessment requirement is evaluated.
#AssessmentPlan: {
	id:               string
	"requirement-id": string @go(RequirementId)
	frequency:        string
	"evaluation-methods": [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)
	"evidence-requirements"?: string @go(EvidenceRequirements)
	parameters?: [#Parameter, ...#Parameter]
}

// AcceptedMethod defines a method for evaluation or enforcement.
#AcceptedMethod: {
	id:           string
	type:         #MethodType
	mode:         #ModeType
	required:     *false | bool @gemara(default=false)
	description?: string
	executor?:    #Actor
}

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

// CatalogImport defines how to import control catalogs with optional exclusions, constraints, and assessment requirement modifications.
#CatalogImport: {
	"reference-id": string @go(ReferenceId)
	exclusions?: [string, ...string]
	constraints?: [#Constraint, ...#Constraint]
	"assessment-requirement-modifications"?: [#AssessmentRequirementModifier, ...#AssessmentRequirementModifier] @go(AssessmentRequirementModifications)
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

// AssessmentRequirementModifier allows organizations to customize assessment requirements based on how an organization wants to gather evidence for the objective.
#AssessmentRequirementModifier: {
	id:                       string
	"target-id":              string   @go(TargetId)
	"modification-type":      #ModType @go(ModificationType)
	"modification-rationale": string   @go(ModificationRationale)
	// The updated text of the assessment requirement
	text?: string
	// The updated applicability of the assessment requirement
	applicability?: [string, ...string]
	// The updated recommendation for the assessment requirement
	recommendation?: string
}

// ModType defines the type of modification to the assessment requirement.
#ModType: "Add" | "Modify" | "Remove" | "Replace" | "Override" @go(-)

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
