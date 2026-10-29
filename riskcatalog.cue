// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// A RiskCatalog is a structured collection of documented risks that may affect an organization,
// system, or service. It provides a centralized reference for risks that can be mapped to threats
// and referenced by policies when documenting how those risks are mitigated or accepted.
#RiskCatalog: {
	#Catalog
	metadata: type: "RiskCatalog"

	// groups narrows the base groups to risk categories with appetite and severity boundaries
	groups?: [#RiskCategory, ...#RiskCategory]

	// risks is a list of risks defined by this catalog
	risks?: [#Risk, ...#Risk] @go(Risks)

	// stakeholders defines the parties, roles, and standing role assignments for this catalog
	stakeholders?: #Stakeholders @go(Stakeholders)

	if risks != _|_ {
		_uniqueRiskIds: {for i, r in risks {(r.id): i}}
		// Each distinct rank value may appear at most once among risks that set rank (partial ranking allowed).
		// Keys are stringified ranks because CUE struct labels cannot be raw integers.
		_uniqueRiskRanks: {for i, r in risks if r.rank != _|_ {"\(r.rank)": i}}
		groups: [#RiskCategory, ...#RiskCategory]
		let _validGroupIds = [for g in groups {g.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, r in risks {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(r.group)
		}

		if stakeholders != _|_ {
			let _validRoleIds = [for r in stakeholders.roles {r.id}]

			for i, r in risks {
				if r."role-ids" != _|_ {
					for j, roleId in r."role-ids" {
						_riskRoleValidation: "\(i)_\(j)": _validRoleIds & list.Contains(roleId)
					}
				}
			}
		}
	}
}

// RiskCategory describes a grouping of risks and defines appetite boundaries
#RiskCategory: {
	#Group

	// appetite defines the acceptable level of risk for this category
	appetite: #RiskAppetite @go(Appetite)

	// max-severity defines the risk tolerance boundary: the highest severity
	// the organization will accept within this category
	"max-severity"?: #Severity @go(MaxSeverity) @yaml("max-severity,omitempty")
}

// Severity defines the assessed level of a risk based on its potential impact and likelihood
#Severity:
	// minor consequence if realized; manageable within normal operations
	"Low" |
	// moderate consequence if realized; may impair specific functions or objectives
	"Medium" |
	// severe consequence if realized; likely to disrupt core operations or objectives
	"High" |
	// extreme consequence if realized; threatens organizational viability or mission
	"Critical" @go(-)

// RiskAppetite defines the acceptable level of exposure for a risk category
#RiskAppetite:
	// organization is willing to accept higher cost to minimize risk
	"Minimal" |
	// organization favors caution but permits limited risk
	"Low" |
	// organization tolerates residual risk when justified by value
	"Moderate" |
	// organization is willing to operate with less restrictive controls
	"High" @go(-)

// A Risk represents the potential for negative impact resulting from one or more threats.
#Risk: {
	// id allows this risk to be referenced by other elements
	id: string

	// title describes the risk
	title: string

	// description explains the risk scenario
	description: string

	// group references by id a catalog group that this risk belongs to
	group: string @go(Group)

	// severity describes the assessed level of this risk
	severity: #Severity @go(Severity)

	// rank optionally orders risks for the same catalog (e.g. when several share the same severity).
	// Lower values mean higher relative importance. Omitted when the four severity levels are enough.
	// When set, each value must be unique among all risks in the catalog that specify rank.
	rank?: int @gemara(projectable=false) @go(Rank) @yaml("rank,omitempty")

	// owner defines the RACI roles responsible for managing this risk
	owner?: #RACI @go(Owner)

	// role-ids references standing stakeholder roles from the catalog that apply to this risk
	"role-ids"?: [string, ...string] @go(RoleIDs) @yaml("role-ids,omitempty")

	// impact describes the business or operational impact
	impact?: string

	// threats link this risk to Layer 2 threats
	"threats"?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Threats)
}
