// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
@if(!stable)
package schemas

@go(gemara)

// Policy represents a policy document with metadata, contacts, scope, imports, implementation plan, risks, and adherence requirements.
#Policy: {
	title:                  string
	metadata:               #Metadata
	contacts:               #Contacts
	scope:                  #Scope
	imports:                #Imports
	"implementation-plan"?: #ImplementationPlan @go(ImplementationPlan)
	risks?:                 #Risks
	adherence:              #Adherence
}

// Contacts defines RACI roles for policy compliance and notification.
#Contacts: {
	// responsible is the person or group responsible for implementing controls for technical requirements
	responsible: [...#Contact]
	// accountable is the person or group accountable for evaluating and enforcing the efficacy of technical controls
	accountable: [...#Contact]
	// consulted is an optional person or group who may be consulted for more information about the technical requirements 
	consulted?: [...#Contact]
	// informed is an optional person or group who must receive updates about compliance with this policy 
	informed?: [...#Contact]
}

// Scope defines what is included and excluded from policy applicability.
#Scope: {
	in:   #Dimensions
	out?: #Dimensions
}

// Dimensions specify the applicability criteria for a policy
#Dimensions: {
	// technologies is an optional list of technology categories or services
	technologies?: [...string]
	// geopolitical is an optional list of geopolitical regions
	geopolitical?: [...string]
	// sensitivity is an optional list of data classification levels
	sensitivity?: [...string]
	// users is an optional list of user roles
	users?: [...string]
	groups?: [...string]
}

// Imports defines external policies, controls, and guidelines required by this policy.
#Imports: {
	policies?: [...string]
	catalogs?: [...#CatalogImport]
	guidance?: [...#GuidanceImport]
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
	mitigated?: [...#MultiMapping]
	// Accepted risks require rationale (justification) and may include scope. Controls addressing these risks are implicitly identified through threat mappings.
	accepted?: [...#AcceptedRisk]
}

// RiskMapping maps a risk to a reference and optionally includes scope and justification.
#AcceptedRisk: {
	risk: #SingleMapping
	// Scope and justification are only required for accepted risks (e.g., risk is accepted for TLP:Green and TLP:Clear because they contain non-sensitive data)
	scope?:         #Scope
	justification?: string
}

// Adherence defines evaluation methods, assessment plans, enforcement methods, and non-compliance notifications.
#Adherence: {
	"evaluation-methods"?: [...#AcceptedMethod] @go(EvaluationMethods)
	"assessment-plans"?: [...#AssessmentPlan] @go(AssessmentPlans)
	"enforcement-methods"?: [...#AcceptedMethod] @go(EnforcementMethods)
	"non-compliance"?: string @go(NonCompliance)
}

// AssessmentPlan defines how a specific assessment requirement is evaluated.
#AssessmentPlan: {
	id:               string
	"requirement-id": string @go(RequirementId)
	frequency:        string
	"evaluation-methods": [...#AcceptedMethod] @go(EvaluationMethods)
	"evidence-requirements"?: string @go(EvidenceRequirements)
	parameters?: [...#Parameter]
}

// AcceptedMethod defines a method for evaluation or enforcement.
#AcceptedMethod: {
	type:         #MethodType | string
	description?: string
	executor?:    #Actor
}

#MethodType: "manual" | "behavioral" | "automated" | "autoremediation" | "gate"

// Parameter defines a configurable parameter for assessment or enforcement activities.
#Parameter: {
	id:          string
	label:       string
	description: string
	"accepted-values"?: [...string] @go(AcceptedValues)
}

// GuidanceImport defines how to import guidance documents with optional exclusions and constraints.
#GuidanceImport: {
	"reference-id": string @go(ReferenceId)
	exclusions?: [...string]
	// Constraints allow policy authors to define ad hoc minimum requirements (e.g., "review at least annually").
	constraints?: [...#Constraint]
}

// CatalogImport defines how to import control catalogs with optional exclusions, constraints, and assessment requirement modifications.
#CatalogImport: {
	"reference-id": string @go(ReferenceId)
	exclusions?: [...string]
	constraints?: [...#Constraint]
	"assessment-requirement-modifications"?: [...#AssessmentRequirementModifier] @go(AssessmentRequirementModifications)
}

// Constraint defines a prescriptive requirement that applies to a specific guidance or control.
#Constraint: {
	// Unique ID for this constraint to enable Layer 5/6 tracking
	id: string
	// Links to the specific Guidance or Control being constrained
	"target-id": string @go(TargetId)
	// The prescriptive requirement/constraint text
	"text": string
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
	applicability?: [...string]
	// The updated recommendation for the assessment requirement
	recommendation?: string
}

// ModType defines the type of modification to the assessment requirement.
#ModType: "add" | "modify" | "remove" | "replace" | "override"
