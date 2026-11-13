// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// GuidanceCatalog represents a concerted documentation effort to help bring about an optimal future without foreknowledge of the implementation details
#GuidanceCatalog: {
	#Catalog
	metadata: type: "GuidanceCatalog"

	// type categorizes this document based on the intent of its contents
	type: #GuidanceType @go(GuidanceType)

	// front-matter provides introductory text for the document to be used during rendering
	"front-matter"?: string @go(FrontMatter) @yaml("front-matter,omitempty")

	// guidelines is a list of unique guidelines defined by this catalog
	guidelines?: [#Guideline, ...#Guideline] @gemara(projectable=false) @go(Guidelines)

	// exemptions provides information about situations where this guidance is not applicable
	exemptions?: [#Exemption, ...#Exemption] @go(Exemptions)

	if exemptions != _|_ {
		let _validRefIds = list.Concat([
			[metadata.id],
			[if metadata."mapping-references" != _|_ for r in metadata."mapping-references" {r.id}],
		])
		for i, ex in exemptions if ex.redirect != _|_ {
			_exemptionRedirectRefValidation: "\(i)": _validRefIds & list.Contains(ex.redirect."reference-id")
		}
	}

	if guidelines != _|_ {
		_uniqueGuidelineIds: {for i, g in guidelines {(g.id): i}}
		groups: [#Group, ...#Group]
		let _validGroupIds = [for g in groups {g.id}]
		let _validRefIds = list.Concat([
			[metadata.id],
			[if metadata."mapping-references" != _|_ for r in metadata."mapping-references" {r.id}],
		])

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, g in guidelines {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(g.group)
			if g.extends != _|_ {
				_guidelineExtendsRefValidation: "\(i)": _validRefIds & list.Contains(g.extends."reference-id")
			}
			if g."replaced-by" != _|_ {
				_guidelineReplacedByRefValidation: "\(i)": _validRefIds & list.Contains(g."replaced-by"."reference-id")
			}
			if g.principles != _|_ {
				for j, p in g.principles {
					_principleRefValidation: "\(i)-\(j)": _validRefIds & list.Contains(p."reference-id")
				}
			}
			if g.vectors != _|_ {
				for j, v in g.vectors {
					_vectorRefValidation: "\(i)-\(j)": _validRefIds & list.Contains(v."reference-id")
				}
			}
		}
		if metadata."applicability-groups" != _|_ {
			let _validApplicabilityIds = [for ag in metadata."applicability-groups" {ag.id}]
			for i, g in guidelines if g.applicability != _|_ {
				for j, a in g.applicability {
					_applicabilityValidation: "\(i)-\(j)": _validApplicabilityIds & list.Contains(a)
				}
			}
		}
	}
}

// GuidanceType restricts the possible types that a catalog may be listed as
#GuidanceType: "Standard" | "Regulation" | "Best Practice" | "Framework" @go(-)

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

	// group provides an id to the group that this guideline belongs to
	group: string @go(Group)

	// recommendations is a list of non-binding suggestions to aid in evaluation or enforcement of the guideline
	recommendations?: [string, ...string]

	// extends is an id for a guideline which this guideline adds to, in this document or elsewhere
	extends?: #EntryMapping @go(Extends,optional=nillable)

	// applicability specifies the contexts in which this guideline applies
	applicability?: [string, ...string] @go(Applicability)

	// rationale provides the context for this guideline
	rationale?: #Rationale @go(Rationale,optional=nillable)

	// statements is a list of structural sub-requirements within a guideline
	statements?: [#Statement, ...#Statement] @go(Statements)

	// principles documents the relationship between this guideline and one or more principles
	"principles"?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Principles) @yaml("principles,omitempty")

	// vector-mappings documents the relationship between this guideline and one or more vectors
	"vectors"?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Vectors) @yaml("vectors,omitempty")

	// see-also lists related guideline IDs within the same GuidanceCatalog
	"see-also"?: [string, ...string] @go(SeeAlso) @yaml("see-also,omitempty")

	// state is the lifecycle state of this guideline
	state: #Lifecycle @gemara(default="Active") @go(State) @yaml("state,omitempty")

	// replaced-by references the guideline that supersedes this one when deprecated or retired
	"replaced-by"?: #EntryMapping @go(ReplacedBy,optional=nillable) @yaml("replaced-by,omitempty")

	// retired guidelines must not have recommendations
	if state == "Retired" {
		recommendations?: _|_
	}
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
	recommendations?: [string, ...string]
}

// Rationale provides a structured way to communicate a guideline author's intent
#Rationale: {
	// importance is an explanation of why this guideline matters
	importance: string

	// goals is a list of outcomes this guideline seeks to achieve
	goals: [string, ...string]
}
