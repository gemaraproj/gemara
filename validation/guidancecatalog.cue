package validation

import "github.com/gemaraproj/gemara"

#GuidanceCatalog: gemara.#GuidanceCatalog & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	families?:   _
	guidelines?: _

	metadata: #Metadata
	metadata: type: "GuidanceCatalog"

	if guidelines != _|_ {
		families: [_, ...gemara.#Group]
	}

	if families != _|_ {
		_uniqueFamilyIds: {for i, f in families {(f.id): i}}
	}

	if guidelines != _|_ {
		_uniqueGuidelineIds: {for i, g in guidelines {(g.id): i}}
	}

	// Local extensions (no reference-id) must share the extended guideline's family
	if guidelines != _|_ {
		for guideline in guidelines if guideline.extends != _|_ {
			if (guideline.extends."reference-id" == "" || guideline.extends."reference-id" == _|_) {
				for extended in guidelines if extended.id == guideline.extends."entry-id" {
					guideline.family == extended.family
				}
			}
		}
	}

	if families != _|_ {
		_validFamilyId: or([for f in families {f.id}])
		guidelines?: [...{family: _validFamilyId}]
	}

	if metadata."applicability-categories" != _|_ {
		_validGdApplicabilityId: or([for c in metadata."applicability-categories" {c.id}])
		guidelines?: [...{
			applicability?: [..._validGdApplicabilityId]
		}]
	}

	if guidelines != _|_ {
		_validGuidelineId: or([for g in guidelines {g.id}])
		guidelines?: [...{
			"see-also"?: [..._validGuidelineId]
		}]
	}

	// metadata.id allows self-referencing in addition to mapping-reference IDs
	if metadata."mapping-references" != _|_ {
		_validGdMappingRefId: metadata.id | or([for r in metadata."mapping-references" {r.id}])
		guidelines?: [...{
			principles?: [...{"reference-id": _validGdMappingRefId}]
			vectors?: [...{"reference-id": _validGdMappingRefId}]
		}]
		exemptions?: [...{
			redirect?: "reference-id": _validGdMappingRefId
		}]
	}

	// Unify each element with validation.#Guideline; the embedded base type only applies
	// gemara.#Guideline, so without this, #Guideline constraints here are not enforced.
	guidelines?: [...#Guideline]
}

#Guideline: gemara.#Guideline & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	state:       _
	statements?: _

	if state == "Retired" {
		recommendations?: _|_
	}

	if statements != _|_ {
		_uniqueStatementIds: {for i, s in statements {(s.id): i}}
	}
}
