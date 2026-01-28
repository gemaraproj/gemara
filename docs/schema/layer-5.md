---
layout: page
title: Layer 5
---

## `EvaluationLog`

EvaluationLog contains the results of evaluating a set of Layer 2 controls.

`evaluations` **array[ControlEvaluation]** _Required_


`metadata` **Metadata**

Metadata represents common metadata fields shared across all layers

## `ControlEvaluation`

ControlEvaluation contains the results of evaluating a single Layer 5 control.

`assessment-logs` **array[object]** _Required_

Enforce that control reference and the assessments' references match

`control` **SingleMapping** _Required_

SingleMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.

`message` **string** _Required_


`name` **string** _Required_


`result` **Result** _Required_


## `AssessmentLog`

AssessmentLog contains the results of executing a single assessment procedure for a control requirement.

`applicability` **array[string]** _Required_

Applicability is elevated from the Layer 2 Assessment Requirement to aid in execution and reporting.

`description` **string** _Required_

Description provides a summary of the assessment procedure.

`message` **string** _Required_

Message provides additional context about the assessment result.

`requirement` **SingleMapping** _Required_

Requirement should map to the assessment requirement for this assessment.

`result` **Result** _Required_

Result is the overall outcome of the assessment procedure, matching the result of the last step that was run.

`start` **Datetime** _Required_

Start is the timestamp when the assessment began.

`steps` **array[AssessmentStep]** _Required_

Steps are sequential actions taken as part of the assessment, which may halt the assessment if a failure occurs.

`confidence-level` **ConfidenceLevel**

ConfidenceLevel indicates the evaluator's confidence level in this specific assessment result.

`end` **Datetime**

End is the timestamp when the assessment concluded.

`plan` **SingleMapping**

Plan maps to the policy assessment plan being executed.

`recommendation` **string**

Recommendation provides guidance on how to address a failed assessment.

`steps-executed` **string**

Steps-executed is the number of steps that were executed as part of the assessment.

## `AssessmentStep`

- **Type**: `string`

---

## `Result`

- **Type**: `string`

---

## `ConfidenceLevel`

ConfidenceLevel indicates the evaluator's confidence level in an assessment result.

- **Type**: `string`

---

