// Schema lifecycle: experimental | stable | deprecated
@status("stable")
package schemas

// Metadata represents common metadata fields shared across all layers
#Metadata: {
	// id allows this entry to be referenced by other elements
	id: string

	// version is the version identifier of this artifact
	version?: string

	// date is the publication or effective date of this artifact
	date?: #Date @go(Date)

	// description provides a high-level summary of the artifact's purpose and scope
	description: string

	// author is the person or group primarily responsible for this artifact
	author: #Actor

	// mapping-references is a list of external documents referenced within this artifact
	"mapping-references"?: [...#MappingReference] @go(MappingReferences) @yaml("mapping-references,omitempty")

	// applicability-categories is a list of categories used to classify within this artifact to specify scope
	"applicability-categories"?: [...#Category] @go(ApplicabilityCategories) @yaml("applicability-categories,omitempty")

	// draft indicates whether this artifact is a pre-release version; open to modification
	draft?: bool

	// lexicon is a URI pointing to a controlled vocabulary or glossary relevant to this artifact
	lexicon?: string
}
