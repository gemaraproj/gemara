// Schema lifecycle: experimental | stable | deprecated
@status("stable")
package gemara

import "list"

@go(gemara)

// ControlCatalog describes a set of related controls and relevant metadata
#ControlCatalog: {
	#Catalog
	metadata: type: "ControlCatalog"

	// controls is a list of unique controls defined by this catalog
	controls?: [#Control, ...#Control] @go(Controls)

	if controls != _|_ {
		_uniqueControlIds: {for i, c in controls {(c.id): i}}
		groups: [#Group, ...#Group]
		metadata: "applicability-groups": [#Group, ...#Group]
		let _validGroupIds = [for g in groups {g.id}]
		let _validApplicabilityIds = [for ag in metadata."applicability-groups" {ag.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, c in controls {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(c.group)
			for j, ar in c."assessment-requirements" {
				for k, a in ar.applicability {
					_applicabilityValidation: "\(i)-\(j)-\(k)": _validApplicabilityIds & list.Contains(a)
				}
			}
		}
	}
}

// Control describes a safeguard or countermeasure with a clear objective and assessment requirements
#Control: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the purpose of this control at a glance
	title: string

	// objective is a unified statement of intent, which may encompass multiple situationally applicable requirements
	objective: string

	// group references by id a catalog group that this control belongs to
	group: string @go(Group)

	// assessment-requirements is a list of requirements that must be verified to confirm the control objective has been met
	"assessment-requirements": [#AssessmentRequirement, ...#AssessmentRequirement] @go(AssessmentRequirements)

	// guidelines documents relationships between this control and Layer 1 guideline artifacts
	guidelines?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Guidelines)

	// threats documents relationships between this control and Layer 2 threat artifacts
	threats?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Threats)

	// state is the lifecycle state of this control
	state: #Lifecycle @go(State) @yaml("state,omitempty")

	// replaced-by references the control that supersedes this one when deprecated or retired
	"replaced-by"?: #EntryMapping @go(ReplacedBy,optional=nillable) @yaml("replaced-by,omitempty")
}

// AssessmentRequirement describes a tightly scoped, verifiable condition that must be satisfied and confirmed by an evaluator
#AssessmentRequirement: {
	// id allows this entry to be referenced by other elements
	id: string

	// text is the body of the requirement, typically written as a MUST condition
	text: string

	// applicability is a list of strings describing the situations where this text functions as a requirement for its parent control
	applicability: [string, ...string]

	// recommendation provides readers with non-binding suggestions to aid in evaluation or enforcement of the requirement
	recommendation?: string

	// state is the lifecycle state of this assessment requirement
	state: #Lifecycle @go(State) @yaml("state,omitempty")

	// replaced-by references the assessment requirement that supersedes this one when deprecated or retired
	"replaced-by"?: #EntryMapping @go(ReplacedBy,optional=nillable) @yaml("replaced-by,omitempty")

	// retired assessment requirements must not have a recommendation
	if state == "Retired" {
		recommendation?: _|_
	}
}
