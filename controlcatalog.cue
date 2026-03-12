// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// ControlCatalog describes a set of related controls and relevant metadata
#ControlCatalog: {
	// title describes the contents of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// extends references control catalogs that this catalog builds upon
	extends?: [...#ArtifactMapping] @go(Extends)

	// families contains a list of control families that can be referenced by controls
	families?: [...#Group] @go(Families)

	// controls is a list of unique controls defined by this catalog
	controls?: [...#Control] @go(Controls)

	// imports contains controls from other sources which are included as part of this document
	imports?: #ControlCatalogImports @go(Imports)
}

// ControlCatalogImports defines imported entries for a control catalog
#ControlCatalogImports: {
	// controls is a list of controls from another source
	controls?: [...#MultiEntryMapping] @go(Controls)
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

	// guidelines documents relationships between this control and Layer 1 guideline artifacts
	guidelines?: [...#MultiEntryMapping] @go(Guidelines)

	// threats documents relationships between this control and Layer 2 threat artifacts
	threats?: [...#MultiEntryMapping] @go(Threats)

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
}
