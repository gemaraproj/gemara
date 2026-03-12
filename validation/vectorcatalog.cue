package validation

import "github.com/gemaraproj/gemara"

#VectorCatalog: gemara.#VectorCatalog & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	vectors?: _

	metadata: #Metadata
	metadata: type: "VectorCatalog"

	if vectors != _|_ {
		_uniqueVectorIds: {for i, v in vectors {(v.id): i}}
	}

	if metadata."applicability-categories" != _|_ {
		_validVcApplicabilityId: or([for c in metadata."applicability-categories" {c.id}])
		vectors?: [...{
			applicability?: [..._validVcApplicabilityId]
		}]
	}
}
