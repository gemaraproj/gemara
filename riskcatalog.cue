// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// A RiskCatalog is a structured collection of documented risks that may affect an organization,
// system, or service. It provides a centralized reference for risks that can be mapped to threats
// and referenced by policies when documenting how those risks are mitigated or accepted.
#RiskCatalog: {

	// title describes the contents of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// categories is a list of risk categories used to classify risks
	categories?: [...#RiskCategory] @go(Categories)

	// risks is a list of risks defined by this catalog
	risks?: [...#Risk] @go(Risks)
}

// RiskCategory describes a grouping of risks and defines appetite boundaries
#RiskCategory: {
	#Group

	// appetite defines the acceptable level of risk for this category
	appetite: #RiskAppetite @go(Appetite)

	// max-severity defines the highest allowed severity within this category
	"max-severity"?: #Severity @go(MaxSeverity) @yaml("max-severity,omitempty")
}

// Severity defines the allowed impact levels for a risk
#Severity: "Low" | "Medium" | "High" | "Critical" @go(-)

// RiskAppetite defines the acceptable level of exposure for a risk category
#RiskAppetite: "Zero" | "Low" | "Moderate" | "High" @go(-)

// A Risk represents the potential for negative impact resulting from one or more threats.
#Risk: {
	// id allows this risk to be referenced by other elements
	id: string

	// title describes the risk
	title: string

	// description explains the risk scenario
	description: string

	// severity describes the impact level
	severity: #Severity @go(Severity)

	// owner defines the RACI roles responsible for managing this risk
	owner?: #RACI @go(Owner)

	// impact describes the business or operational impact
	impact?: string

	// threats link this risk to Layer 2 threats
	"threats"?: [...#MultiEntryMapping] @go(Threats)
}
