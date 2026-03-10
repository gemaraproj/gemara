// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// A VectorCatalog is a structured collection of documented vectors,
// serving as a centralized reference for known attack methods and exploitation pathways that may be relevant to a particular domain, framework, or security model.

#VectorCatalog: {
	// title describes the contents of this catalog
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// vectors is a list of attack vectors documented in this catalog
	vectors?: [...#Vector] @go(Vectors)
}

// A Vector represents a method, pathway, or technique through which a threat may be realized or an attack may be carried out.
#Vector: {
	// id allows this vector to be referenced by other elements
	id: string

	// title describes the vector
	title: string

	// description explains how the attack vector works
	description: string

	// applicability specifies the contexts in which this vector can manifest
	applicability?: [...string] @go(Applicability)
}
