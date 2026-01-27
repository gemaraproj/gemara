---
layout: page
title: Layer 5
---

## `EvaluationLog`

EvaluationLog contains the results of evaluating a set of Layer 2 controls.

Required:

- `evaluations`

Optional:

- `metadata`

---

### `evaluations`

- **Type**: `array`
- **Items**: [ControlEvaluation]

---

### `metadata` (optional)

Metadata represents common metadata fields shared across all layers

- **Type**: [Metadata]

Required if `metadata` is present:

- `author`
- `description`
- `id`

Optional:

- `applicability-categories`
- `date`
- `draft`
- `lexicon`
- `mapping-references`
- `version`

---

#### `metadata.applicability-categories` (optional)

applicability-categories is a list of categories used to classify within this artifact to specify scope

- **Type**: `array`
- **Items**: [Category]

---

#### `metadata.author`

author is the person or group primarily responsible for this artifact

- **Type**: [Actor]

Required if `metadata.author` is present:

- `id`
- `name`
- `type`

Optional:

- `contact`
- `description`
- `uri`
- `version`

---

##### `metadata.author.contact` (optional)

contact is contact information for the actor

- **Type**: [Contact]

Required if `metadata.author.contact` is present:

- `name`

Optional:

- `affiliation`
- `email`
- `social`

---

###### `metadata.author.contact.affiliation` (optional)

affiliation is the organization with which the contact entity is associated, such as a team, school, or employer

- **Type**: `string`

---

###### `metadata.author.contact.email` (optional)

email is the preferred email address to reach the contact

- **Type**: [Email]

---

###### `metadata.author.contact.name`

name is the preferred descriptor for the contact entity

- **Type**: `string`

---

###### `metadata.author.contact.social` (optional)

social is a social media handle or other profile for the contact, such as GitHub

- **Type**: `string`

---

##### `metadata.author.description` (optional)

description provides additional context about the actor

- **Type**: `string`

---

##### `metadata.author.id`

id uniquely identifies the actor and allows this entry to be referenced by other elements

- **Type**: `string`

---

##### `metadata.author.name`

name is the name of the actor

- **Type**: `string`

---

##### `metadata.author.type`

type specifies the type of entity interacting in the workflow

- **Type**: [ActorType]

---

##### `metadata.author.uri` (optional)

uri is a general URI for the actor information

- **Type**: `string`

---

##### `metadata.author.version` (optional)

version is the version of the actor (for tools; if applicable)

- **Type**: `string`

---

#### `metadata.date` (optional)

date is the publication or effective date of this artifact

- **Type**: [Date]

---

#### `metadata.description`

description provides a high-level summary of the artifact's purpose and scope

- **Type**: `string`

---

#### `metadata.draft` (optional)

draft indicates whether this artifact is a pre-release version; open to modification

- **Type**: `boolean`

---

#### `metadata.id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

#### `metadata.lexicon` (optional)

lexicon is a URI pointing to a controlled vocabulary or glossary relevant to this artifact

- **Type**: `string`

---

#### `metadata.mapping-references` (optional)

mapping-references is a list of external documents referenced within this artifact

- **Type**: `array`
- **Items**: [MappingReference]

---

#### `metadata.version` (optional)

version is the version identifier of this artifact

- **Type**: `string`

---

## `ControlEvaluation`

ControlEvaluation contains the results of evaluating a single Layer 5 control.

Required:

- `assessment-logs`
- `assessment-logs`
- `control`
- `message`
- `name`
- `result`

---

### `assessment-logs`

Enforce that control reference and the assessments' references match

- **Type**: `array`
- **Items**: `object`

---

### `control`

SingleMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.

- **Type**: [SingleMapping]

Required if `control` is present:

- `entry-id`

Optional:

- `reference-id`
- `remarks`
- `strength`

---

#### `control.entry-id`

entry-id is the identifier being mapped to in the referenced artifact

- **Type**: `string`

---

#### `control.reference-id` (optional)

reference-id is the id for a MappingReference entry in the artifact's metadata

- **Type**: `string`

---

#### `control.remarks` (optional)

remarks is prose describing the mapping relationship

- **Type**: `string`

---

#### `control.strength` (optional)

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

- **Type**: `string`

---

### `message`

- **Type**: `string`

---

### `name`

- **Type**: `string`

---

### `result`

- **Type**: [Result]

---

## `AssessmentLog`

AssessmentLog contains the results of executing a single assessment procedure for a control requirement.

Required:

- `applicability`
- `description`
- `message`
- `requirement`
- `result`
- `start`
- `steps`

Optional:

- `confidence-level`
- `end`
- `plan`
- `recommendation`
- `steps-executed`

---

### `applicability`

Applicability is elevated from the Layer 2 Assessment Requirement to aid in execution and reporting.

- **Type**: `array`
- **Items**: `string`

---

### `confidence-level` (optional)

ConfidenceLevel indicates the evaluator's confidence level in this specific assessment result.

- **Type**: [ConfidenceLevel]

---

### `description`

Description provides a summary of the assessment procedure.

- **Type**: `string`

---

### `end` (optional)

End is the timestamp when the assessment concluded.

- **Type**: [Datetime]

---

### `message`

Message provides additional context about the assessment result.

- **Type**: `string`

---

### `plan` (optional)

Plan maps to the policy assessment plan being executed.

- **Type**: [SingleMapping]

Required if `plan` is present:

- `entry-id`

Optional:

- `reference-id`
- `remarks`
- `strength`

---

#### `plan.entry-id`

entry-id is the identifier being mapped to in the referenced artifact

- **Type**: `string`

---

#### `plan.reference-id` (optional)

reference-id is the id for a MappingReference entry in the artifact's metadata

- **Type**: `string`

---

#### `plan.remarks` (optional)

remarks is prose describing the mapping relationship

- **Type**: `string`

---

#### `plan.strength` (optional)

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

- **Type**: `string`

---

### `recommendation` (optional)

Recommendation provides guidance on how to address a failed assessment.

- **Type**: `string`

---

### `requirement`

Requirement should map to the assessment requirement for this assessment.

- **Type**: [SingleMapping]

Required if `requirement` is present:

- `entry-id`

Optional:

- `reference-id`
- `remarks`
- `strength`

---

#### `requirement.entry-id`

entry-id is the identifier being mapped to in the referenced artifact

- **Type**: `string`

---

#### `requirement.reference-id` (optional)

reference-id is the id for a MappingReference entry in the artifact's metadata

- **Type**: `string`

---

#### `requirement.remarks` (optional)

remarks is prose describing the mapping relationship

- **Type**: `string`

---

#### `requirement.strength` (optional)

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

- **Type**: `string`

---

### `result`

Result is the overall outcome of the assessment procedure, matching the result of the last step that was run.

- **Type**: [Result]

---

### `start`

Start is the timestamp when the assessment began.

- **Type**: [Datetime]

---

### `steps`

Steps are sequential actions taken as part of the assessment, which may halt the assessment if a failure occurs.

- **Type**: `array`
- **Items**: [AssessmentStep]

---

### `steps-executed` (optional)

Steps-executed is the number of steps that were executed as part of the assessment.

- **Type**: `string`

---

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

