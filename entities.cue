// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
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

	// uri is a general URI for the entity information.
	// Any URI scheme is accepted (e.g. https, file, oci, s3, arn) so entities
	// hosted outside http(s) can be referenced.
	uri?: =~"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]+$"
}

// Actor represents an entity (human or tool) that performs actions in evaluations
#Actor: {
	#Entity

	// contact is contact information for the actor
	contact?: #Contact @go(Contact)

	// execution-environment describes where this actor runs. When the actor is the
	// executor a Policy names on an AcceptedMethod, it states the environment the
	// policy accepts that method's results from; when the actor is a log's author, it
	// records the environment the log was produced in.
	"execution-environment"?: #ExecutionEnvironment @go(ExecutionEnvironment,optional=nillable)
}

// ExecutionEnvironment identifies the environment an actor runs in, so that a log's
// author can be compared with the executor its policy accepts. It describes the
// actor, never the resources the actor evaluates.
#ExecutionEnvironment: {
	// digests pins the content the actor runs, one entry per component such as an
	// image manifest or a plugin binary; the actor's uri and version name it
	digests?: [#Digest, ...#Digest] @go(Digests,type=[]string)

	// config-digest pins the configuration the actor runs with
	"config-digest"?: #Digest @go(ConfigDigest,type=string)

	// observation-vantage states where the actor observes its targets from
	"observation-vantage"?: #ObservationVantage @go(ObservationVantage)
}

// ObservationVantage states whether every input a result depends on was obtained at a
// vantage the evaluated resource could neither forge nor suppress (substrate: a network
// boundary, syscall supervision, a hypervisor's read of guest state) or rests on output
// the resource produced or could influence (artifact). A result drawing on both is artifact.
// Values follow observation_vantage in https://github.com/probityai/agent-evidence-vocabulary
#ObservationVantage: "substrate" | "artifact" @go(-)

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
