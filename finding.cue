// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

@go(gemara)

// Disposition is how a finding was settled. Open is outstanding, Resolved was
// dealt with, and Waived means a subject that cannot conform was granted an
// exception, so the finding stands and enforcement stands down.
//
// It is independent of #Effect, which says what an enforcement action did to its
// target: a blocked merge leaves its finding Open. The documented
// correspondences live on #Effect, and like #Result's roll-up precedence they are
// stated rather than enforced.
#Disposition: "Open" | "Resolved" | "Waived" @go(-)

// Finding records something an enforcement action or an audit found, together
// with what became of it. Its identity across logs is derived, never stored: the
// target of the log that holds it, and the control it concerns, or the risk when
// it names no control, each resolved to an identifier that means the same thing
// in every document. Findings with the same identity are the same finding, which
// is what lets a later log resolve a finding an earlier one opened.
//
// The shape is the one OCSF's Compliance Finding expects, so a finding projects
// without rearranging.
#Finding: {
	// disposition is how this finding was settled: still outstanding, dealt with,
	// or waived because a subject that cannot conform was granted an exception. It
	// is deliberately not #Effect, which says what an enforcement action did.
	disposition?: #Disposition

	// id identifies this finding where something references it. A finding
	// reported against a named check needs none, so it is optional here and
	// required where it is addressed.
	id?: string

	// control names the control this finding concerns. A finding cites a control
	// rather than one of its assessment requirements: requirements are what gets
	// tested, and a control is what gets remediated.
	control?: #EntryMapping @go(Control,optional=nillable)

	// description explains what was found. It is optional because a routine
	// conclusion needs no prose: a control decided Compliant says enough by
	// saying so. An override, or anything a reader would question, explains itself.
	description?: string

	// severity states how material the finding is, on the same scale the risk
	// catalog uses. It is optional here because not every finding rates one: an
	// audit finding that records conformity has no materiality to state. It is
	// required where a finding is negative by construction: anything an audit
	// determines Not Compliant, and every finding in an enforcement
	// justification.
	severity?: #Severity

	// recommendations record the corrective actions for this finding.
	recommendations?: [#Recommendation, ...#Recommendation]
}

// Recommendation provides a corrective action for a finding.
#Recommendation: {
	// id uniquely identifies this recommendation
	id?: string

	// text describes the recommended corrective action
	text: string

	// required indicates whether this recommendation is a mandatory corrective action
	required: *false | bool @gemara(default=false)
}
