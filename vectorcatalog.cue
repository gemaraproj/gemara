// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

import "list"

@go(gemara)

// A VectorCatalog is a structured collection of documented vectors,
// serving as a centralized reference for known attack methods and exploitation pathways that may be relevant to a particular domain, framework, or security model.

#VectorCatalog: {
	#Catalog
	metadata: type: "VectorCatalog"

	// vectors is a list of attack vectors documented in this catalog
	vectors?: [#Vector, ...#Vector] @go(Vectors)

	if vectors != _|_ {
		_uniqueVectorIds: {for i, v in vectors {(v.id): i}}
		groups: [#Group, ...#Group]
		let _validGroupIds = [for g in groups {g.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, v in vectors {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(v.group)
		}
		if metadata."applicability-groups" != _|_ {
			let _validApplicabilityIds = [for ag in metadata."applicability-groups" {ag.id}]
			for i, v in vectors if v.applicability != _|_ {
				for j, a in v.applicability {
					_applicabilityValidation: "\(i)-\(j)": _validApplicabilityIds & list.Contains(a)
				}
			}
		}
	}
}

// A Vector represents a method, pathway, or technique through which a threat may be realized or an attack may be carried out.
#Vector: {
	// id allows this vector to be referenced by other elements
	id: string

	// title describes the vector
	title: string

	// description explains how the attack vector works
	description: string

	// group references by id a catalog group that this vector belongs to
	group: string @go(Group)

	// applicability specifies the contexts in which this vector can manifest
	applicability?: [string, ...string] @go(Applicability)
}
