// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@status("stable")
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

	if groups != _|_ {
		_uniqueGroupIds: {for i, g in groups {(g.id): i}}
	}

	if extends != _|_ {
		metadata: "mapping-references": [#MappingReference, ...#MappingReference]
	}

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
