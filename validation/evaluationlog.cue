package validation

import "github.com/gemaraproj/gemara"

#EvaluationLog: gemara.#EvaluationLog & {
	metadata: #Metadata
	metadata: type: "EvaluationLog"

	// Each assessment-log requirement must reference the same source as its parent control
	evaluations: [...{
		control: _
		"assessment-logs": [...{
			requirement: "reference-id": control."reference-id"
		}]
	}]
}
