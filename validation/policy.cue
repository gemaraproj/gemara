package validation

import "github.com/gemaraproj/gemara"

#Policy: gemara.#Policy & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	imports:   _
	risks?:    _
	adherence: _

	metadata: #Metadata
	metadata: type: "Policy"

	// metadata.id allows self-referencing in addition to mapping-reference IDs
	if metadata."mapping-references" != _|_ {
		_validPolicyMappingRefId: metadata.id | or([for r in metadata."mapping-references" {r.id}])

		if imports.policies != _|_ {
			imports: policies: [...{"reference-id": _validPolicyMappingRefId}]
		}

		if imports.catalogs != _|_ {
			imports: catalogs: [...{"reference-id": _validPolicyMappingRefId}]
		}

		if imports.guidance != _|_ {
			imports: guidance: [...{"reference-id": _validPolicyMappingRefId}]
		}

		if risks != _|_ if risks.mitigated != _|_ {
			risks: mitigated: [...{"reference-id": _validPolicyMappingRefId}]
		}
	}

	if adherence."assessment-plans" != _|_ {
		_uniquePlanIds: {for i, p in adherence."assessment-plans" {(p.id): i}}
	}
}
