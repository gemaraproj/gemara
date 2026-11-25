// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

import "list"

@go(gemara)

// CapabilityCatalog describes a collection of system capabilities
#CapabilityCatalog: {
	#Catalog
	metadata: type: "CapabilityCatalog"

	// capabilities is a list of capabilities defined by this catalog
	capabilities?: [#Capability, ...#Capability] @go(Capabilities)

	// when capabilities are listed, groups is required, capability ids must be unique, each capability's group must be a declared group id, and each citation reference-id must be declared in metadata.mapping-references
	if capabilities != _|_ {
		_uniqueCapabilityIds: {for i, c in capabilities {(c.id): i}}
		groups: [#Group, ...#Group]
		let _validGroupIds = [for g in groups {g.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, c in capabilities {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(c.group)
		}

		let _validRefIds = [for k, v in metadata if k == "mapping-references" for r in v if r.id != _|_ {r.id}]
		for i, c in capabilities if c.citations != _|_ {
			for j, cit in c.citations {
				_citationRefValidation: "\(i)-\(j)": _validRefIds & list.Contains(cit."reference-id")
			}
		}
	}
}

// Capability describes a system capability such as a feature, component or object.
#Capability: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes this capability at a glance
	title: string

	// description provides a detailed overview of this capability
	description: string

	// group references by id a catalog group that this capability belongs to
	group: string @go(Group)

	// citations lists the external sources this entry is drawn from
	citations?: [#ArtifactMapping, ...#ArtifactMapping] @go(Citations)
}
