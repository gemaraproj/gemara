// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

import "list"

@go(gemara)

// PrincipleCatalog describes a set of related principles and relevant metadata
#PrincipleCatalog: {
	#Catalog
	metadata: type: "PrincipleCatalog"

	// principles is a list of unique principles defined by this catalog
	principles?: [#Principle, ...#Principle] @go(Principles)

	if principles != _|_ {
		_uniquePrinciplesIds: {for i, p in principles {(p.id): i}}
		groups: [#Group, ...#Group]
		let _validGroupIds = [for g in groups {g.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, p in principles {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(p.group)
		}
	}
}

// Principle represents a foundational value or tenet that guides governance, design, and operational decisions
#Principle: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the principle at a glance
	title: string

	// description explains the principle and its expected outcomes
	description: string

	// group references by id a catalog group that this principle belongs to
	group: string @go(Group)

	// rationale provides the context for this principle
	rationale?: string
}
