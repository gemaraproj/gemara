// SPDX-License-Identifier: Apache-2.0

// Throwaway prototype: evidence-weighted catalog links and explicit coverage gaps.
@gemara(status="experimental")
package gemara

@go(gemara)

// MappingTargetMetadata carries the author's assessment of a target relationship.
// Absence means unreported; none of these attributes has a default.
#MappingTargetMetadata: {
	// strength is the author's estimate of how completely the source satisfies this target; range 1-10
	strength?:           int & >=1 & <=10 @go(Strength)
	"confidence-level"?: #ConfidenceLevel @go(ConfidenceLevel)
	// applicability constrains the contexts in which this target mapping holds
	applicability?: [string, ...string] @go(Applicability)
	// rationale explains why this relationship exists for this target
	rationale?: string
	// remarks is general prose regarding this target mapping
	remarks?: string
}

// CatalogEntryMapping extends an entry pointer with node-to-target semantics.
#CatalogEntryMapping: {
	#EntryMapping
	#MappingTargetMetadata
	relationship: *"relates-to" | #RelationshipType @gemara(default="relates-to") @go(Relationship)
}

// CatalogMappingTarget preserves the legacy grouped inline entry identifier.
// Here reference-id names an entry; the parent group names its artifact.
#CatalogMappingTarget: {
	#ArtifactMapping
	#MappingTargetMetadata
}

// CatalogMultiEntryMapping is one mapping assertion on the containing node.
// Separate assertions retain separate ids, even for identical source/target pairs.
#CatalogMultiEntryMapping: {
	#ArtifactMapping
	id?:          string
	relationship: *"relates-to" | #RelationshipType @gemara(default="relates-to") @go(Relationship)
	entries?: [#CatalogMappingTarget, ...#CatalogMappingTarget] @go(Entries)

	if relationship != "no-match" {
		entries: [#CatalogMappingTarget, ...#CatalogMappingTarget]
	}
}
