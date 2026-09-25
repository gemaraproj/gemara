// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

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

// Adherence defines evaluation methods, assessment plans, enforcement methods, and non-compliance notifications.
#Adherence: {
	"evaluation-methods"?: [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)
	"assessment-plans"?: [#AssessmentPlan, ...#AssessmentPlan] @go(AssessmentPlans)
	"enforcement-methods"?: [#AcceptedMethod & {type: #EnforcementMethodType}, ...#AcceptedMethod & {type: #EnforcementMethodType}] @go(EnforcementMethods)
	"non-compliance"?: string @go(NonCompliance)
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
// Under every option, methods that agree resolve to the result they agree on.
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
//     so the order is always defined. Two latest logs with the same start that disagree
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
