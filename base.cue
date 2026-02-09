// Schema lifecycle: experimental | stable | deprecated
@status("stable")
package gemara

import "time"

@go(gemara)

// Contact is the contact information for a person or group
#Contact: {
	// name is the preferred descriptor for the contact entity
	name: string

	// affiliation is the organization with which the contact entity is associated, such as a team, school, or employer
	affiliation?: string @go(Affiliation,type=*string)

	// email is the preferred email address to reach the contact
	email?: #Email @go(Email,type=*Email)

	// social is a social media handle or other profile for the contact, such as GitHub
	social?: string @go(Social,type=*string)
}

// Actor represents an entity (human or tool) that can perform actions in evaluations
#Actor: {
	// id uniquely identifies the actor and allows this entry to be referenced by other elements
	id: string

	// name is the name of the actor
	name: string

	// type specifies the type of entity interacting in the workflow
	type: #ActorType @go(Type)

	// version is the version of the actor (for tools; if applicable)
	version?: string

	// description provides additional context about the actor
	description?: string

	// uri is a general URI for the actor information
	uri?: =~"^https?://[^\\s]+$"

	// contact is contact information for the actor
	contact?: #Contact @go(Contact)
}

// ActorType specifies what entity is interacting in the workflow
#ActorType: "Human" | "Software" | "Software-Assisted" @go(-)

// Email represents a validated email address pattern
#Email: =~"^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$"

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
