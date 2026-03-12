package validation

import "github.com/gemaraproj/gemara"

#ControlCatalog: gemara.#ControlCatalog & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	families?: _
	controls?: _
	extends?:  _
	imports?:  _

	metadata: #Metadata
	metadata: type: "ControlCatalog"

	if controls != _|_ {
		families: [_, ...gemara.#Group]
	}

	if families != _|_ {
		_uniqueFamilyIds: {for i, f in families {(f.id): i}}
	}

	if controls != _|_ {
		_uniqueControlIds: {for i, c in controls {(c.id): i}}
	}

	if families != _|_ {
		_validCtlFamilyId: or([for f in families {f.id}])
		controls?: [...{family: _validCtlFamilyId}]
	}

	if metadata."applicability-categories" != _|_ {
		_validCtlApplicabilityId: or([for c in metadata."applicability-categories" {c.id}])
		controls?: [...{
			"assessment-requirements": [...{
				applicability?: [..._validCtlApplicabilityId]
			}]
		}]
	}

	// metadata.id allows self-referencing in addition to mapping-reference IDs
	if metadata."mapping-references" != _|_ {
		_validCtlMappingRefId: metadata.id | or([for r in metadata."mapping-references" {r.id}])

		if extends != _|_ {
			extends?: [...{"reference-id": _validCtlMappingRefId}]
		}

		controls?: [...{
			guidelines?: [...{"reference-id": _validCtlMappingRefId}]
			threats?: [...{"reference-id": _validCtlMappingRefId}]
		}]

		if imports != _|_ if imports.controls != _|_ {
			imports: controls: [...{"reference-id": _validCtlMappingRefId}]
		}
	}

	// Unify each element with validation.#Control; the embedded base type only applies
	// gemara.#Control, so without this, #Control constraints here are not enforced.
	controls?: [...#Control]
}

#Control: gemara.#Control & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	state: _

	// Alias re-declares the hyphenated field for use in the comprehension below
	ARs="assessment-requirements": _
	_uniqueARIds: {for i, r in ARs {(r.id): i}}

	// Unify each element with validation.#AssessmentRequirement; the embedded base type
	// only applies gemara.#AssessmentRequirement, so without this, #AssessmentRequirement
	// constraints here are not enforced.
	"assessment-requirements": [...#AssessmentRequirement]
}

#AssessmentRequirement: gemara.#AssessmentRequirement & {
	// Re-declare embedded fields in local scope for if-guards (_=top, no extra constraint)
	state: _

	if state == "Retired" {
		recommendation?: _|_
	}
}
