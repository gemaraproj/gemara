// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@status("stable")
package gemara

@go(gemara)

// Entity represents a human or tool
#Entity: {
	// id uniquely identifies the entity and allows this entry to be referenced by other elements
	id: string

	// name is the name of the entity
	name: string

	// type specifies the type of entity interacting in the workflow
	type: #EntityType

	// version is the version of the entity (for tools; if applicable)
	version?: string

	// description provides additional context about the entity
	description?: string

	// uri is a general URI for the entity information
	uri?: =~"^https?://[^\\s]+$"
}

// Actor represents an entity (human or tool) that performs actions in evaluations
#Actor: {
	#Entity

	// contact is contact information for the actor
	contact?: #Contact @go(Contact)
}

// Resource represents an entity that exists in the system and can be evaluated
#Resource: {
	#Entity

	// environment describes where the resource exists (e.g., production, staging, development, specific region)
	environment?: string @go(Environment)

	// owner is the contact information for the person or group responsible for managing or owning this resource
	owner?: #Contact @go(Owner)
}

// EntityType specifies what entity is interacting in the workflow
#EntityType: "Human" | "Software" | "Software Assisted" @go(-)

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

// Email represents a validated email address pattern
#Email: =~"^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$"

// RACI defines the roles responsible for managing an artifact
#RACI: {
	// responsible identifies the entities responsible for executing work to manage or mitigate the artifact
	responsible: [#Contact, ...#Contact]

	// accountable identifies the entity ultimately accountable for the outcome
	accountable: [#Contact, ...#Contact]

	// consulted identifies entities whose input is required when assessing or responding to the artifact
	consulted?: [#Contact, ...#Contact]

	// informed identifies entities that should be notified about changes to the artifact status
	informed?: [#Contact, ...#Contact]
}
