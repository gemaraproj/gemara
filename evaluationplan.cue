// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// EvaluationPlan is the Layer 4 artifact that states which executors are expected to
// run which assessment procedures, in what environment, and how their results are
// deconflicted. Layer 4 also carries the EvaluationLog, which records what actually
// ran. The plan is the asked-for half and the log is the observed half; the two are
// reconcilable only when both name the same executors and environments.
#EvaluationPlan: {
	// title describes the purpose of this evaluation plan at a glance
	title: string

	// metadata provides detailed data about this document
	metadata: #Metadata @go(Metadata)
	metadata: type: "EvaluationPlan"
	metadata: "mapping-references": [#MappingReference, ...#MappingReference]

	// policy identifies the Layer 3 Policy this plan executes; must match a mapping-reference id
	policy: #ArtifactMapping @go(Policy)

	// procedures is one or more planned assessments, each naming the executors expected to run it
	procedures: [#PlannedAssessment, ...#PlannedAssessment]

	// conflict-resolution states how disagreeing executor results are resolved for this plan.
	// A plan naming more than one executor for any procedure MUST state one, because a
	// multi-source evaluation with no stated resolution rule produces a result whose value
	// depends on which log a consumer happened to read first.
	"conflict-resolution"?: #ConflictPolicy @go(ConflictResolution)

	matchN(>=1, [
		{"conflict-resolution"!: #ConflictPolicy},
		{procedures: [...{executors: [_]}]},
	])
}

// PlannedAssessment binds one assessment requirement to the executors expected to evaluate it.
#PlannedAssessment: {
	// id uniquely identifies this planned assessment within the plan
	id: string

	// requirement-id maps to the assessment requirement being planned
	"requirement-id": string @go(RequirementId)

	// plan-id references the #AssessmentPlan in the Layer 3 Policy that this procedure executes
	"plan-id": #EntryMapping @go(PlanId)

	// executors is one or more executors expected to run this procedure
	executors: [#PlannedExecutor, ...#PlannedExecutor]

	// description provides a summary of what this procedure evaluates
	description?: string

	// Executor ranks within a procedure must be unique, so that a rank-based
	// conflict resolution has a total order to work with.
	_uniqueRanks: {for i, e in executors if e.rank != _|_ {"\(e.rank)": i}}
}

// PlannedExecutor names an executor expected to run a procedure, its precedence,
// and the environment it is expected to run in.
#PlannedExecutor: {
	// executor identifies the actor expected to perform the assessment
	executor: #Actor

	// rank orders this executor against the others on the same procedure; lower is higher precedence
	rank?: int & >=1

	// environment-requirements states the environment this executor is expected to run in.
	// Its observed counterpart belongs on the EvaluationLog: a plan that states a required
	// environment and a log that records none are not reconcilable, and the pair is the
	// point of stating either.
	"environment-requirements"?: #EnvironmentSpec @go(EnvironmentRequirements)
}

// EnvironmentSpec states the properties of an execution environment that a plan requires
// and that a log can be checked against. Every field is optional because a plan may
// constrain as little or as much as it needs, and a field that is absent from the plan
// is unconstrained rather than unimportant.
#EnvironmentSpec: {
	// label is a human-readable name for the environment, such as production or staging
	label?: string

	// image-id identifies the container or machine image the executor is expected to run from
	"image-id"?: string @go(ImageId)

	// digests pins the exact content the executor is expected to run, by algorithm:encoded digest.
	// Naming a digest here is what makes an environment claim in a log checkable rather than
	// comparable only as a string.
	digests?: [...#Digest]

	// config-digest pins the runtime configuration the executor is expected to run under
	"config-digest"?: #Digest @go(ConfigDigest)

	// observation-source names how the environment is to be observed. A plan that requires
	// an observation source the executor itself controls has required a self-report.
	"observation-source"?: string @go(ObservationSource)
}

// Digest is a cryptographic hash in algorithm:encoded form (e.g. sha256:abc123...).
// The grammar is exactly the one #EvidenceMapping.digest already uses, so a digest
// written in a plan and a digest written in an evidence mapping are the same shape.
#Digest: =~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$"

// ConflictPolicy states how disagreeing executor results are resolved.
#ConflictPolicy: "highest-rank" | "unanimous" | "most-recent" | "escalate" @go(-)
