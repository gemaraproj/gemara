// Schema lifecycle: experimental | stable | deprecated
@status("stable")

package gemara

import "time"

@go(gemara)

// Datetime represents an ISO 8601 formatted datetime string
#Datetime: time.Format("2006-01-02T15:04:05Z07:00") @go(Datetime,format="date-time")


// Group represents a classification or grouping that can be used in different contexts with semantic meaning derived from its usage
#Group: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the purpose of this group at a glance
	title: string

	// description explains the significance and traits of entries to this group
	description: string
}

// Metadata represents common metadata fields shared across all layers
#Metadata: {
	// id allows this entry to be referenced by other elements
	id: string

	// type identifies the kind of Gemara artifact for unambiguous parsing
	type: #ArtifactType

	// gemara-version declares which version of the Gemara specification this artifact conforms to
	"gemara-version": string @go(GemaraVersion) @yaml("gemara-version")

	// version is the version identifier of this artifact
	version?: string

	// date is the publication or effective date of this artifact
	date?: #Datetime @go(Date)

	// description provides a high-level summary of the artifact's purpose and scope
	description: string

	// author is the person or group primarily responsible for this artifact
	author: #Actor

	// mapping-references is a list of external documents referenced within this artifact
	MR="mapping-references"?: [#MappingReference, ...#MappingReference] @go(MappingReferences) @yaml("mapping-references,omitempty")

	// applicability-groups is a list of groups used to classify within this artifact to specify scope
	AG="applicability-groups"?: [#Group, ...#Group] @go(ApplicabilityGroups) @yaml("applicability-groups,omitempty")

	// draft indicates whether this artifact is a pre-release version; open to modification
	draft?: bool

	// lexicon is a URI pointing to a controlled vocabulary or glossary relevant to this artifact
	lexicon?: #ArtifactMapping @go(Lexicon,optional=nillable)

	if MR != _|_ {
		_uniqueRefIds: {for i, r in MR {(r.id): i}}
	}
	if AG != _|_ {
		_uniqueGroupsIds: {for i, c in AG {(c.id): i}}
	}
}

// ArtifactType identifies the kind of Gemara artifact for unambiguous parsing
#ArtifactType: "CapabilityCatalog" | "ControlCatalog" | "GuidanceCatalog" | "ThreatCatalog" | "RiskCatalog" | "Policy" | "MappingDocument" | "Lexicon" | "EvaluationLog" | "EnforcementLog" | "VectorCatalog" | "PrincipleCatalog" | "AuditLog" @go(-)
