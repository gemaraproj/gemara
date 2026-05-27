// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// Lexicon is a controlled vocabulary or glossary artifact referenced by Metadata.lexicon
#Lexicon: {
	// title describes the purpose of this lexicon at a glance
	title: string

	// metadata provides detailed data about this document
	metadata: #Metadata @go(Metadata)
	metadata: type: "Lexicon"

	// terms is one or more defined entries for linking and rendering
	terms: [#LexiconTerm, ...#LexiconTerm] @go(Terms)

	_uniqueTermIds: {for i, t in terms {(t.id): i}}
}

// LexiconTerm is a single definition within a lexicon
#LexiconTerm: {
	// id allows this entry to be referenced for anchors and tooling
	id: string

	// title is the canonical name of the defined concept
	title: string

	// definition explains the meaning of the term
	definition: string

	// synonyms lists alternative labels that should resolve to this term for linking
	synonyms?: [string, ...string] @go(Synonyms)

	// references cites external authorities supporting the definition
	references?: [#LexiconReference, ...#LexiconReference] @go(References)
}

// LexiconReference cites a source supporting a lexicon definition
#LexiconReference: {
	// citation identifies the source material in prose
	citation: string

	// url points to supporting material when available
	url?: =~"^(https?|file)://[^\\s]+$"
}
