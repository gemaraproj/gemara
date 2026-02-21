// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
@if(!stable)
package gemara

@go(gemara)

// ControlCatalog describes a set of related controls and relevant metadata
#ControlCatalog: {
	// title describes the contents of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	"metadata": #Metadata @go(Metadata)

	// families contains a list of control families that can be referenced by controls
	families?: [...#Group] @go(Families)

	// controls is a list of unique controls defined by this catalog
	controls?: [...#Control] @go(Controls)

	// imported-controls is a list of controls from another source which are included as part of this document
	"imported-controls"?: [...#MultiEntryMapping] @go(ImportedControls)

	if controls != _|_ {
		families: [_, ...#Group]
	}
}

// Control describes a safeguard or countermeasure with a clear objective and assessment requirements
#Control: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the purpose of this control at a glance
	title: string

	// objective is a unified statement of intent, which may encompass multiple situationally applicable requirements
	objective: string

	// family references by id a catalog control family that this control belongs to
	family: string @go(Family)

	// assessment-requirements is a list of requirements that must be verified to confirm the control objective has been met
	"assessment-requirements": [...#AssessmentRequirement] @go(AssessmentRequirements)

	// guideline-mappings documents relationships betwen this control and Layer 1 guideline artifacts
	"guideline-mappings"?: [...#MultiEntryMapping] @go(GuidelineMappings)

	// threat-mappings documents relationships betwen this control and Layer 2 threat artifacts
	"threat-mappings"?: [...#MultiEntryMapping] @go(ThreatMappings)

	// state is the lifecycle state of this control
	state: #Lifecycle @go(State) @yaml("state,omitempty")

	// replaced-by references the control that supersedes this one when deprecated or retired
	"replaced-by"?: #EntryMapping @go(ReplacedBy,optional=nillable) @yaml("replaced-by,omitempty")
}

// AssessmentRequirement describes a tightly scoped, verifiable condition that must be satisfied and confirmed by an evaluator
#AssessmentRequirement: {
	// id allows this entry to be referenced by other elements
	id: string

	// text is the body of the requirement, typically written as a MUST condition
	text: string

	// applicability is a list of strings describing the situations where this text functions as a requirement for its parent control
	applicability: [...string]

	// recommendation provides readers with non-binding suggestions to aid in evaluation or enforcement of the requirement
	recommendation?: string

	// state is the lifecycle state of this assessment requirement
	state: #Lifecycle @go(State) @yaml("state,omitempty")

	// replaced-by references the assessment requirement that supersedes this one when deprecated or retired
	"replaced-by"?: #EntryMapping @go(ReplacedBy,optional=nillable) @yaml("replaced-by,omitempty")

	// retired assessment requirements must not have a recommendation
	if state == "retired" {
		recommendation?: _|_
	}
}

// ThreatCatalog describes a set of topically-associated threats
#ThreatCatalog: {
	// title describes the purpose of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	"metadata": #Metadata @go(Metadata)

	// threats is a list of threats defined by this catalog
	threats?: [...#Threat] @go(Threats)

	// capabilities is a list of capabilities that make up the system being assessed
	capabilities?: [...#Capability] @go(Capabilities)

	// imported-threats is a list of threats from another source which are included as part of this document
	"imported-threats"?: [...#MultiEntryMapping] @go(ImportedThreats)

	// imported-capabilities is a list of capabilities from another source which are included as part of this document
	"imported-capabilities"?: [...#MultiEntryMapping] @go(ImportedCapabilities)
}

// Threat describes a specifically-scoped opportunity for a negative impact to the organization
#Threat: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes this threat at a glance
	title: string

	// description provides a detailed explanation of an opportunity for negative impact 
	description: string

	// capabilities documents the relationship between this threat and a system capability
	capabilities: [...#MultiEntryMapping]

	// actors describes the relevant internal or external threat actors
	actors?: [...#Actor]

	// external-mappings documents relationships between this threat and any other artifacts
	"external-mappings"?: [...#MultiEntryMapping] @go(ExternalMappings)
}

// Capability describes a system capability such as a feature, component or object.
#Capability: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes this capability at a glance
	title: string

	// description provides a detailed overview of this capability
	description: string
}
