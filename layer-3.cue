// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// A RiskCatalog is a structured collection of documented risks that may affect an organization,
// system, or service. It provides a centralized reference for risks that can be mapped to threats
// and referenced by policies when documenting how those risks are mitigated or accepted.
#RiskCatalog: {

	// title describes the contents of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// categories is a list of risk categories used to classify risks
	categories?: [#RiskCategory, ...#RiskCategory] @go(Categories)

	// risks is a list of risks defined by this catalog
	risks?: [#Risk, ...#Risk] @go(Risks)

	// Constraints
	if risks != _|_ {
		categories: [_, ...#RiskCategory]
	}
}

// RiskCategory describes a grouping of risks and defines appetite boundaries
#RiskCategory: {
	#Group

	// appetite defines the acceptable level of risk for this category
	appetite: #RiskAppetite @go(Appetite)

	// max-severity defines the highest allowed severity within this category
	"max-severity"?: #Severity @go(MaxSeverity) @yaml("max-severity,omitempty")
}

// Severity defines the allowed impact levels for a risk
#Severity: "Low" | "Medium" | "High" | "Critical" @go(-)

// RiskAppetite defines the acceptable level of exposure for a risk category
#RiskAppetite: "Zero" | "Low" | "Moderate" | "High" @go(-)

// A Risk represents the potential for negative impact resulting from one or more threats.
#Risk: {
	// id allows this risk to be referenced by other elements
	id: string

	// title describes the risk
	title: string

	// description explains the risk scenario
	description: string

	// category references by id a catalog risk category that this risk belongs to
	category: string @go(Category)

	// severity describes the impact level
	severity: #Severity @go(Severity)

	// owner defines the RACI roles responsible for managing this risk
	owner?: #RACI @go(Owner)

	// impact describes the business or operational impact
	impact?: string

	// threats link this risk to Layer 2 threats
	"threats"?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Threats)
}

// Policy represents a policy document with metadata, contacts, scope, imports, implementation plan, risks, and adherence requirements.
#Policy: {
	title:                  string
	metadata:               #Metadata
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
	"evaluation-methods"?: [#AcceptedMethod, ...#AcceptedMethod] @go(EvaluationMethods)
	"assessment-plans"?: [#AssessmentPlan, ...#AssessmentPlan] @go(AssessmentPlans)
	"enforcement-methods"?: [#AcceptedMethod, ...#AcceptedMethod] @go(EnforcementMethods)
	"non-compliance"?: string @go(NonCompliance)
}

// AssessmentPlan defines how a specific assessment requirement is evaluated.
#AssessmentPlan: {
	id:               string
	"requirement-id": string @go(RequirementId)
	frequency:        string
	"evaluation-methods": [#AcceptedMethod, ...#AcceptedMethod] @go(EvaluationMethods)
	"evidence-requirements"?: string @go(EvidenceRequirements)
	parameters?: [#Parameter, ...#Parameter]
}

// AcceptedMethod defines a method for evaluation or enforcement.
#AcceptedMethod: {
	type:         #MethodType
	description?: string
	executor?:    #Actor
}

#MethodType: "Manual" | "Behavioral" | "Automated" | "Autoremediation" | "Gate" @go(-)

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
