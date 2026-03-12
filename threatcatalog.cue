// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// ThreatCatalog describes a set of topically-associated threats
#ThreatCatalog: {
	// title describes the purpose of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// extends references threat catalogs that this catalog builds upon
	extends?: [...#ArtifactMapping] @go(Extends)

	// threats is a list of threats defined by this catalog
	threats?: [...#Threat] @go(Threats)

	// capabilities is a list of capabilities that make up the system being assessed
	capabilities?: [...#Capability] @go(Capabilities)

	// imports contains threats and capabilities from other sources which are included as part of this document
	imports?: #ThreatCatalogImports @go(Imports)
}

// ThreatCatalogImports defines imported entries for a threat catalog
#ThreatCatalogImports: {
	// threats is a list of threats from another source
	threats?: [...#MultiEntryMapping] @go(Threats)

	// capabilities is a list of capabilities from another source
	capabilities?: [...#MultiEntryMapping] @go(Capabilities)
}

// Threat describes a specifically-scoped opportunity for a negative impact to the organization
#Threat: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes this threat at a glance
	title: string

	// description provides a detailed explanation of an opportunity for negative impact 
	description: string

	// capabilities documents the relationship between this threat and a system capability
	capabilities: [...#MultiEntryMapping]

	// vectors documents the relationship between this threat and one or more vectors
	vectors?: [...#MultiEntryMapping] @go(Vectors)

	// actors describes the relevant internal or external threat actors
	actors?: [...#Actor]
}

// Capability describes a system capability such as a feature, component or object.
#Capability: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes this capability at a glance
	title: string

	// description provides a detailed overview of this capability
	description: string
}
