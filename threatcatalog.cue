// Schema lifecycle: experimental | stable | deprecated
@status("stable")
package gemara

import "list"

@go(gemara)

// ThreatCatalog describes a set of topically-associated threats
#ThreatCatalog: {
	#Catalog
	metadata: type: "ThreatCatalog"

	// threats is a list of threats defined by this catalog
	threats?: [#Threat, ...#Threat] @go(Threats)

	if threats != _|_ {
		_uniqueThreatIds: {for i, t in threats {(t.id): i}}
		groups: [#Group, ...#Group]
		let _validGroupIds = [for g in groups {g.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, t in threats {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(t.group)
		}
	}
}

// Threat describes a specifically-scoped opportunity for a negative impact to the organization
#Threat: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes this threat at a glance
	title: string

	// description provides a detailed explanation of an opportunity for negative impact
	description: string

	// group references by id a catalog group that this threat belongs to
	group: string @go(Group)

	// capabilities documents the relationship between this threat and a system capability
	capabilities: [#MultiEntryMapping, ...#MultiEntryMapping]

	// vectors documents the relationship between this threat and one or more vectors
	vectors?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Vectors)

	// actors describes the relevant internal or external threat actors
	actors?: [#Actor, ...#Actor]
}
