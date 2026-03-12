// Package validation is a policy overlay for authoring Gemara artifacts.
// It layers referential integrity, uniqueness, and cross-field consistency
// checks on top of the base schema definitions in the gemara package.
package validation

import "github.com/gemaraproj/gemara"

#Metadata: gemara.#Metadata & {
	// Alias re-declares the hyphenated field for use in the comprehension below
	MR="mapping-references"?: _
	if MR != _|_ {
		_uniqueRefIds: {for i, r in MR {(r.id): i}}
	}

	// Alias re-declares the hyphenated field for use in the comprehension below
	AC="applicability-categories"?: _
	if AC != _|_ {
		_uniqueCatIds: {for i, c in AC {(c.id): i}}
	}
}
