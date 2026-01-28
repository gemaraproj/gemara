// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
@if(!stable)
package schemas

@go(gemara)

// GuidanceCatalog represents a concerted documentation effort to help bring about an optimal future without foreknowledge of the implementation details
#GuidanceCatalog: {
	// title describes the contents of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// type categorizes this document based on the intent of its contents
	type: #GuidanceType @go(GuidanceType)

	// front-matter provides introductory text for the document to be used during rendering
	"front-matter"?: string @go(FrontMatter) @yaml("front-matter,omitempty")

	// families contains a list of guidance families that can be referenced by guidance
	families?: [...#Family] @go(Families)

	// guidelines is a list of unique guidelines defined by this catalog
	guidelines?: [...#Guideline] @go(Guidelines)

	// exemptions provides information about situations where this guidance is not applicable
	exemptions?: [...#Exemption] @go(Exemptions)

	// guidelines that extend other guidelines must be in the same family as the extended guideline
	_validateExtensions: {
		for guideline in guidelines if guideline.extends != _|_ {
			if (guideline.extends."reference-id" == "" || guideline.extends."reference-id" == _|_) {
				for extended in guidelines if extended.id == guideline.extends."entry-id" {
					guideline.family == extended.family
				}
			}
		}
	}
}

// GuidanceType restricts the possible types that a catalog may be listed as
#GuidanceType: "Standard" | "Regulation" | "Best Practice" | "Framework"

// Exemption describes a single scenario where the catalog is not applicable
#Exemption: {
	// description identifies who or what is exempt from the full guidance
	description: string

	// reason explains why the exemption is granted
	reason: string

	// redirect points to alternative guidelines or controls that should be followed instead
	redirect?: #MultiEntryMapping @go(Redirect,optional=nillable)
}

// Guideline provides explanatory context and recommendations for designing optimal outcomes 
#Guideline: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the contents of this guideline
	title: string

	// objective is a unified statement of intent, which may encompass multiple situationally applicable statements
	objective: string

	// family provides an id to the family that this guideline belongs to
	family: string @go(Family)

	// recommendations is a list of non-binding suggestions to aid in evaluation or enforcement of the guideline
	recommendations?: [...string]

	// extends is an id for a guideline which this guideline adds to, in this document or elsewhere
	extends?: #EntryMapping @go(Extends,optional=nillable)

	// applicability specifies the contexts in which this guideline applies
	applicability?: [...string] @go(Applicability)

	// rationale provides the context for this guideline
	rationale?: #Rationale @go(Rationale,optional=nillable)

	// statements is a list of structural sub-requirements within a guideline
	statements?: [...#Statement] @go(Statements)

	// guideline-mappings documents the relationship between this guideline and external guidelines
	"guideline-mappings"?: [...#MultiEntryMapping] @go(GuidelineMappings) @yaml("guideline-mappings,omitempty")

	// principle-mappings documents the relationship between this guideline and one or more principles
	"principle-mappings"?: [...#MultiEntryMapping] @go(PrincipleMappings) @yaml("principle-mappings,omitempty")

	// vector-mappings documents the relationship between this guideline and one or more vectors
	"vector-mappings"?: [...#MultiEntryMapping] @go(VectorMappings) @yaml("vector-mappings,omitempty")

	// see-also lists related guideline IDs within the same GuidanceCatalog
	"see-also"?: [...string] @go(SeeAlso) @yaml("see-also,omitempty")
}

// Statement represents a structural sub-requirement within a guideline;
// They do not increase strictness and all statements within a guideline apply together
#Statement: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the contents of this statement
	title?: string

	// text is the body of this statement
	text: string

	// recommendations is a list of non-binding suggestions to aid in evaluation or enforcement of the statement
	recommendations?: [...string]
}

// Rationale provides a structured way to communicate a guideline author's intent
#Rationale: {
	// importance is an explanation of why this guideline matters
	importance: string

	// goals is a list of outcomes this guideline seeks to achieve
	goals: [...string]
}
