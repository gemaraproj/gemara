// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
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

	if risks != _|_ {
		_uniqueRiskIds: {for i, r in risks {(r.id): i}}
		groups: [#RiskCategory, ...#RiskCategory]
		let _validGroupIds = [for g in groups {g.id}]

		// Unify the valid ID list with a list.Contains constraint to require each entry's value exists
		for i, r in risks {
			_groupValidation: "\(i)": _validGroupIds & list.Contains(r.group)
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

	// owner defines the RACI roles responsible for managing this risk
	owner?: #RACI @go(Owner)

	// impact describes the business or operational impact
	impact?: string

	// threats link this risk to Layer 2 threats
	"threats"?: [#MultiEntryMapping, ...#MultiEntryMapping] @go(Threats)
}
