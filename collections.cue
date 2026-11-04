// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// Catalog describes a set of topically-associated entries
#Catalog: {
	// title describes the purpose of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// groups contains a list of groups that can be referenced by entries in this catalog
	groups?: [#Group, ...#Group]

	// extends references catalogs that this catalog builds upon
	extends?: [...#ArtifactMapping] @go(Extends)

	imports?: [#MultiEntryMapping, ...#MultiEntryMapping]

	// group ids must be unique
	if groups != _|_ {
		_uniqueGroupIds: {for i, g in groups {(g.id): i}}
	}

	// when extends is set, metadata.mapping-references is required
	if extends != _|_ {
		metadata: "mapping-references": [#MappingReference, ...#MappingReference]
	}

	// when imports are listed, metadata.mapping-references is required
	if imports != _|_ {
		metadata: "mapping-references": [#MappingReference, ...#MappingReference]
	}
}

// Log describes a set of recorded entries from a measurement activity
#Log: {
	// metadata provides detailed data about this log
	metadata: #Metadata @go(Metadata)

	// target identifies the resource being evaluated
	target: #Resource @go(Target)
}

// Lifecycle represents the lifecycle state of a guideline, control, or assessment requirement
#Lifecycle: *"Active" | "Draft" | "Deprecated" | "Retired" @go(-)

// ConfidenceLevel indicates the evaluator's confidence level in an assessment result.
#ConfidenceLevel: "Undetermined" | "Low" | "Medium" | "High" @go(-)
