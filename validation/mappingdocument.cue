package validation

import "github.com/gemaraproj/gemara"

#MappingDocument: gemara.#MappingDocument & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	mappings: _

	metadata: #Metadata
	metadata: type: "MappingDocument"

	// Base type allows empty mapping-references; MappingDocument requires at least one
	// because source-reference and target-reference must resolve against them.
	metadata: "mapping-references": [_, ...gemara.#MappingReference]

	_uniqueMappingIds: {for i, m in mappings {(m.id): i}}

	_validMappingDocRefId: or([for r in metadata."mapping-references" {r.id}])
	"source-reference": "reference-id": _validMappingDocRefId
	"target-reference": "reference-id": _validMappingDocRefId

	if metadata."applicability-categories" != _|_ {
		_validMappingApplicabilityId: or([for c in metadata."applicability-categories" {c.id}])
		mappings: [...{
			applicability?: [..._validMappingApplicabilityId]
		}]
	}

	// Unify each element with validation.#Mapping; the embedded base type only applies
	// gemara.#Mapping, so without this, #Mapping constraints here are not enforced.
	mappings: [...#Mapping]
}

#Mapping: gemara.#Mapping & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	relationship: _

	// no-match has no counterpart in the target artifact
	if relationship != "no-match" {
		target: gemara.#EntryReference
	}
}
