package validation

import "github.com/gemaraproj/gemara"

#ThreatCatalog: gemara.#ThreatCatalog & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	threats?:      _
	capabilities?: _
	extends?:      _
	imports?:      _

	metadata: #Metadata
	metadata: type: "ThreatCatalog"

	if threats != _|_ {
		_uniqueThreatIds: {for i, t in threats {(t.id): i}}
	}

	if capabilities != _|_ {
		_uniqueCapabilityIds: {for i, c in capabilities {(c.id): i}}
	}

	// metadata.id allows self-referencing in addition to mapping-reference IDs
	if metadata."mapping-references" != _|_ {
		_validThreatMappingRefId: metadata.id | or([for r in metadata."mapping-references" {r.id}])

		threats?: [...{
			capabilities: [...{"reference-id": _validThreatMappingRefId}]
			vectors?: [...{"reference-id": _validThreatMappingRefId}]
		}]

		if extends != _|_ {
			extends?: [...{"reference-id": _validThreatMappingRefId}]
		}

		if imports != _|_ if imports.threats != _|_ {
			imports: threats: [...{"reference-id": _validThreatMappingRefId}]
		}

		if imports != _|_ if imports.capabilities != _|_ {
			imports: capabilities: [...{"reference-id": _validThreatMappingRefId}]
		}
	}
}
