// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// PrincipleCatalog describes a set of related principles and relevant metadata
#PrincipleCatalog: {
	#Catalog
	metadata: type: "PrincipleCatalog"

	// principles is a list of unique principles defined by this catalog
	principles?: [#Principle, ...#Principle] @go(Principles)

	// when principles are listed, groups is required, principle ids must be unique, each principle's group must be a declared group id, and each citation reference-id must be declared in metadata.mapping-references
	if principles != _|_ {
		_uniquePrinciplesIds: {for i, p in principles {(p.id): i}}
		groups: [#Group, ...#Group]
		let _validGroupIds = [for g in groups {g.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, p in principles {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(p.group)
		}

		let _validRefIds = [for k, v in metadata if k == "mapping-references" for r in v if r.id != _|_ {r.id}]
		for i, p in principles if p.citations != _|_ {
			for j, cit in p.citations {
				_citationRefValidation: "\(i)-\(j)": _validRefIds & list.Contains(cit."reference-id")
			}
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

	// citations lists the external sources this entry is drawn from
	citations?: [#ArtifactMapping, ...#ArtifactMapping] @go(Citations)
}
