---
layout: page
title: Layer 3
---

## `Policy`

Policy represents a policy document with metadata, contacts, scope, imports, implementation plan, risks, and adherence requirements.

Required:

- `adherence`
- `contacts`
- `imports`
- `metadata`
- `scope`
- `title`

Optional:

- `implementation-plan`
- `risks`

---

### `adherence`

Adherence defines evaluation methods, assessment plans, enforcement methods, and non-compliance notifications.

- **Type**: [Adherence]


Optional:

- `assessment-plans`
- `enforcement-methods`
- `evaluation-methods`
- `non-compliance`

---

#### `adherence.assessment-plans` (optional)

- **Type**: `array`
- **Items**: [AssessmentPlan]

---

#### `adherence.enforcement-methods` (optional)

- **Type**: `array`
- **Items**: [AcceptedMethod]

---

#### `adherence.evaluation-methods` (optional)

- **Type**: `array`
- **Items**: [AcceptedMethod]

---

#### `adherence.non-compliance` (optional)

- **Type**: `string`

---

### `contacts`

Contacts defines RACI roles for policy compliance and notification.

- **Type**: [Contacts]

Required if `contacts` is present:

- `accountable`
- `responsible`

Optional:

- `consulted`
- `informed`

---

#### `contacts.accountable`

accountable is the person or group accountable for evaluating and enforcing the efficacy of technical controls

- **Type**: `array`
- **Items**: [Contact]

---

#### `contacts.consulted` (optional)

consulted is an optional person or group who may be consulted for more information about the technical requirements

- **Type**: `array`
- **Items**: [Contact]

---

#### `contacts.informed` (optional)

informed is an optional person or group who must receive updates about compliance with this policy

- **Type**: `array`
- **Items**: [Contact]

---

#### `contacts.responsible`

responsible is the person or group responsible for implementing controls for technical requirements

- **Type**: `array`
- **Items**: [Contact]

---

### `implementation-plan` (optional)

ImplementationPlan defines when and how the policy becomes active.

- **Type**: [ImplementationPlan]

Required if `implementation-plan` is present:

- `enforcement-timeline`
- `evaluation-timeline`

Optional:

- `notification-process`

---

#### `implementation-plan.enforcement-timeline`

ImplementationDetails specifies the timeline for policy implementation.

- **Type**: [ImplementationDetails]

Required if `implementation-plan.enforcement-timeline` is present:

- `notes`
- `start`

Optional:

- `end`

---

##### `implementation-plan.enforcement-timeline.end` (optional)

- **Type**: [Datetime]

---

##### `implementation-plan.enforcement-timeline.notes`

- **Type**: `string`

---

##### `implementation-plan.enforcement-timeline.start`

- **Type**: [Datetime]

---

#### `implementation-plan.evaluation-timeline`

ImplementationDetails specifies the timeline for policy implementation.

- **Type**: [ImplementationDetails]

Required if `implementation-plan.evaluation-timeline` is present:

- `notes`
- `start`

Optional:

- `end`

---

##### `implementation-plan.evaluation-timeline.end` (optional)

- **Type**: [Datetime]

---

##### `implementation-plan.evaluation-timeline.notes`

- **Type**: `string`

---

##### `implementation-plan.evaluation-timeline.start`

- **Type**: [Datetime]

---

#### `implementation-plan.notification-process` (optional)

- **Type**: `string`

---

### `imports`

Imports defines external policies, controls, and guidelines required by this policy.

- **Type**: [Imports]


Optional:

- `catalogs`
- `guidance`
- `policies`

---

#### `imports.catalogs` (optional)

- **Type**: `array`
- **Items**: [CatalogImport]

---

#### `imports.guidance` (optional)

- **Type**: `array`
- **Items**: [GuidanceImport]

---

#### `imports.policies` (optional)

- **Type**: `array`
- **Items**: `string`

---

### `metadata`

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

### `risks` (optional)

Risks defines mitigated and accepted risks addressed by this policy.

- **Type**: [Risks]


Optional:

- `accepted`
- `mitigated`

---

#### `risks.accepted` (optional)

Accepted risks require rationale (justification) and may include scope. Controls addressing these risks are implicitly identified through threat mappings.

- **Type**: `array`
- **Items**: [AcceptedRisk]

---

#### `risks.mitigated` (optional)

Mitigated risks only need reference-id and risk-id (no justification required)

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `scope`

Scope defines what is included and excluded from policy applicability.

- **Type**: [Scope]

Required if `scope` is present:

- `in`

Optional:

- `out`

---

#### `scope.in`

Dimensions specify the applicability criteria for a policy

- **Type**: [Dimensions]


Optional:

- `geopolitical`
- `groups`
- `sensitivity`
- `technologies`
- `users`

---

##### `scope.in.geopolitical` (optional)

geopolitical is an optional list of geopolitical regions

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.groups` (optional)

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.sensitivity` (optional)

sensitivity is an optional list of data classification levels

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.technologies` (optional)

technologies is an optional list of technology categories or services

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.users` (optional)

users is an optional list of user roles

- **Type**: `array`
- **Items**: `string`

---

#### `scope.out` (optional)

Dimensions specify the applicability criteria for a policy

- **Type**: [Dimensions]


Optional:

- `geopolitical`
- `groups`
- `sensitivity`
- `technologies`
- `users`

---

##### `scope.out.geopolitical` (optional)

geopolitical is an optional list of geopolitical regions

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.groups` (optional)

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.sensitivity` (optional)

sensitivity is an optional list of data classification levels

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.technologies` (optional)

technologies is an optional list of technology categories or services

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.users` (optional)

users is an optional list of user roles

- **Type**: `array`
- **Items**: `string`

---

### `title`

- **Type**: `string`

---

## `Contacts`

Contacts defines RACI roles for policy compliance and notification.

Required:

- `accountable`
- `responsible`

Optional:

- `consulted`
- `informed`

---

### `accountable`

accountable is the person or group accountable for evaluating and enforcing the efficacy of technical controls

- **Type**: `array`
- **Items**: [Contact]

---

### `consulted` (optional)

consulted is an optional person or group who may be consulted for more information about the technical requirements

- **Type**: `array`
- **Items**: [Contact]

---

### `informed` (optional)

informed is an optional person or group who must receive updates about compliance with this policy

- **Type**: `array`
- **Items**: [Contact]

---

### `responsible`

responsible is the person or group responsible for implementing controls for technical requirements

- **Type**: `array`
- **Items**: [Contact]

---

## `Scope`

Scope defines what is included and excluded from policy applicability.

Required:

- `in`

Optional:

- `out`

---

### `in`

Dimensions specify the applicability criteria for a policy

- **Type**: [Dimensions]


Optional:

- `geopolitical`
- `groups`
- `sensitivity`
- `technologies`
- `users`

---

#### `in.geopolitical` (optional)

geopolitical is an optional list of geopolitical regions

- **Type**: `array`
- **Items**: `string`

---

#### `in.groups` (optional)

- **Type**: `array`
- **Items**: `string`

---

#### `in.sensitivity` (optional)

sensitivity is an optional list of data classification levels

- **Type**: `array`
- **Items**: `string`

---

#### `in.technologies` (optional)

technologies is an optional list of technology categories or services

- **Type**: `array`
- **Items**: `string`

---

#### `in.users` (optional)

users is an optional list of user roles

- **Type**: `array`
- **Items**: `string`

---

### `out` (optional)

Dimensions specify the applicability criteria for a policy

- **Type**: [Dimensions]


Optional:

- `geopolitical`
- `groups`
- `sensitivity`
- `technologies`
- `users`

---

#### `out.geopolitical` (optional)

geopolitical is an optional list of geopolitical regions

- **Type**: `array`
- **Items**: `string`

---

#### `out.groups` (optional)

- **Type**: `array`
- **Items**: `string`

---

#### `out.sensitivity` (optional)

sensitivity is an optional list of data classification levels

- **Type**: `array`
- **Items**: `string`

---

#### `out.technologies` (optional)

technologies is an optional list of technology categories or services

- **Type**: `array`
- **Items**: `string`

---

#### `out.users` (optional)

users is an optional list of user roles

- **Type**: `array`
- **Items**: `string`

---

## `Dimensions`

Dimensions specify the applicability criteria for a policy


Optional:

- `geopolitical`
- `groups`
- `sensitivity`
- `technologies`
- `users`

---

### `geopolitical` (optional)

geopolitical is an optional list of geopolitical regions

- **Type**: `array`
- **Items**: `string`

---

### `groups` (optional)

- **Type**: `array`
- **Items**: `string`

---

### `sensitivity` (optional)

sensitivity is an optional list of data classification levels

- **Type**: `array`
- **Items**: `string`

---

### `technologies` (optional)

technologies is an optional list of technology categories or services

- **Type**: `array`
- **Items**: `string`

---

### `users` (optional)

users is an optional list of user roles

- **Type**: `array`
- **Items**: `string`

---

## `Imports`

Imports defines external policies, controls, and guidelines required by this policy.


Optional:

- `catalogs`
- `guidance`
- `policies`

---

### `catalogs` (optional)

- **Type**: `array`
- **Items**: [CatalogImport]

---

### `guidance` (optional)

- **Type**: `array`
- **Items**: [GuidanceImport]

---

### `policies` (optional)

- **Type**: `array`
- **Items**: `string`

---

## `ImplementationPlan`

ImplementationPlan defines when and how the policy becomes active.

Required:

- `enforcement-timeline`
- `evaluation-timeline`

Optional:

- `notification-process`

---

### `enforcement-timeline`

ImplementationDetails specifies the timeline for policy implementation.

- **Type**: [ImplementationDetails]

Required if `enforcement-timeline` is present:

- `notes`
- `start`

Optional:

- `end`

---

#### `enforcement-timeline.end` (optional)

- **Type**: [Datetime]

---

#### `enforcement-timeline.notes`

- **Type**: `string`

---

#### `enforcement-timeline.start`

- **Type**: [Datetime]

---

### `evaluation-timeline`

ImplementationDetails specifies the timeline for policy implementation.

- **Type**: [ImplementationDetails]

Required if `evaluation-timeline` is present:

- `notes`
- `start`

Optional:

- `end`

---

#### `evaluation-timeline.end` (optional)

- **Type**: [Datetime]

---

#### `evaluation-timeline.notes`

- **Type**: `string`

---

#### `evaluation-timeline.start`

- **Type**: [Datetime]

---

### `notification-process` (optional)

- **Type**: `string`

---

## `ImplementationDetails`

ImplementationDetails specifies the timeline for policy implementation.

Required:

- `notes`
- `start`

Optional:

- `end`

---

### `end` (optional)

- **Type**: [Datetime]

---

### `notes`

- **Type**: `string`

---

### `start`

- **Type**: [Datetime]

---

## `Risks`

Risks defines mitigated and accepted risks addressed by this policy.


Optional:

- `accepted`
- `mitigated`

---

### `accepted` (optional)

Accepted risks require rationale (justification) and may include scope. Controls addressing these risks are implicitly identified through threat mappings.

- **Type**: `array`
- **Items**: [AcceptedRisk]

---

### `mitigated` (optional)

Mitigated risks only need reference-id and risk-id (no justification required)

- **Type**: `array`
- **Items**: [MultiMapping]

---

## `AcceptedRisk`

RiskMapping maps a risk to a reference and optionally includes scope and justification.

Required:

- `risk`

Optional:

- `justification`
- `scope`

---

### `justification` (optional)

- **Type**: `string`

---

### `risk`

SingleMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.

- **Type**: [SingleMapping]

Required if `risk` is present:

- `entry-id`

Optional:

- `reference-id`
- `remarks`
- `strength`

---

#### `risk.entry-id`

entry-id is the identifier being mapped to in the referenced artifact

- **Type**: `string`

---

#### `risk.reference-id` (optional)

reference-id is the id for a MappingReference entry in the artifact's metadata

- **Type**: `string`

---

#### `risk.remarks` (optional)

remarks is prose describing the mapping relationship

- **Type**: `string`

---

#### `risk.strength` (optional)

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

- **Type**: `string`

---

### `scope` (optional)

Scope and justification are only required for accepted risks (e.g., risk is accepted for TLP:Green and TLP:Clear because they contain non-sensitive data)

- **Type**: [Scope]

Required if `scope` is present:

- `in`

Optional:

- `out`

---

#### `scope.in`

Dimensions specify the applicability criteria for a policy

- **Type**: [Dimensions]


Optional:

- `geopolitical`
- `groups`
- `sensitivity`
- `technologies`
- `users`

---

##### `scope.in.geopolitical` (optional)

geopolitical is an optional list of geopolitical regions

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.groups` (optional)

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.sensitivity` (optional)

sensitivity is an optional list of data classification levels

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.technologies` (optional)

technologies is an optional list of technology categories or services

- **Type**: `array`
- **Items**: `string`

---

##### `scope.in.users` (optional)

users is an optional list of user roles

- **Type**: `array`
- **Items**: `string`

---

#### `scope.out` (optional)

Dimensions specify the applicability criteria for a policy

- **Type**: [Dimensions]


Optional:

- `geopolitical`
- `groups`
- `sensitivity`
- `technologies`
- `users`

---

##### `scope.out.geopolitical` (optional)

geopolitical is an optional list of geopolitical regions

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.groups` (optional)

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.sensitivity` (optional)

sensitivity is an optional list of data classification levels

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.technologies` (optional)

technologies is an optional list of technology categories or services

- **Type**: `array`
- **Items**: `string`

---

##### `scope.out.users` (optional)

users is an optional list of user roles

- **Type**: `array`
- **Items**: `string`

---

## `Adherence`

Adherence defines evaluation methods, assessment plans, enforcement methods, and non-compliance notifications.


Optional:

- `assessment-plans`
- `enforcement-methods`
- `evaluation-methods`
- `non-compliance`

---

### `assessment-plans` (optional)

- **Type**: `array`
- **Items**: [AssessmentPlan]

---

### `enforcement-methods` (optional)

- **Type**: `array`
- **Items**: [AcceptedMethod]

---

### `evaluation-methods` (optional)

- **Type**: `array`
- **Items**: [AcceptedMethod]

---

### `non-compliance` (optional)

- **Type**: `string`

---

## `AssessmentPlan`

AssessmentPlan defines how a specific assessment requirement is evaluated.

Required:

- `evaluation-methods`
- `frequency`
- `id`
- `requirement-id`

Optional:

- `evidence-requirements`
- `parameters`

---

### `evaluation-methods`

- **Type**: `array`
- **Items**: [AcceptedMethod]

---

### `evidence-requirements` (optional)

- **Type**: `string`

---

### `frequency`

- **Type**: `string`

---

### `id`

- **Type**: `string`

---

### `parameters` (optional)

- **Type**: `array`
- **Items**: [Parameter]

---

### `requirement-id`

- **Type**: `string`

---

## `AcceptedMethod`

AcceptedMethod defines a method for evaluation or enforcement.

Required:

- `type`

Optional:

- `description`
- `executor`

---

### `description` (optional)

- **Type**: `string`

---

### `executor` (optional)

Actor represents an entity (human or tool) that can perform actions in evaluations

- **Type**: [Actor]

Required if `executor` is present:

- `id`
- `name`
- `type`

Optional:

- `contact`
- `description`
- `uri`
- `version`

---

#### `executor.contact` (optional)

contact is contact information for the actor

- **Type**: [Contact]

Required if `executor.contact` is present:

- `name`

Optional:

- `affiliation`
- `email`
- `social`

---

##### `executor.contact.affiliation` (optional)

affiliation is the organization with which the contact entity is associated, such as a team, school, or employer

- **Type**: `string`

---

##### `executor.contact.email` (optional)

email is the preferred email address to reach the contact

- **Type**: [Email]

---

##### `executor.contact.name`

name is the preferred descriptor for the contact entity

- **Type**: `string`

---

##### `executor.contact.social` (optional)

social is a social media handle or other profile for the contact, such as GitHub

- **Type**: `string`

---

#### `executor.description` (optional)

description provides additional context about the actor

- **Type**: `string`

---

#### `executor.id`

id uniquely identifies the actor and allows this entry to be referenced by other elements

- **Type**: `string`

---

#### `executor.name`

name is the name of the actor

- **Type**: `string`

---

#### `executor.type`

type specifies the type of entity interacting in the workflow

- **Type**: [ActorType]

---

#### `executor.uri` (optional)

uri is a general URI for the actor information

- **Type**: `string`

---

#### `executor.version` (optional)

version is the version of the actor (for tools; if applicable)

- **Type**: `string`

---

### `type`

- **Type**: `string`

---

## `MethodType`

- **Type**: `string`

---

## `Parameter`

Parameter defines a configurable parameter for assessment or enforcement activities.

Required:

- `description`
- `id`
- `label`

Optional:

- `accepted-values`

---

### `accepted-values` (optional)

- **Type**: `array`
- **Items**: `string`

---

### `description`

- **Type**: `string`

---

### `id`

- **Type**: `string`

---

### `label`

- **Type**: `string`

---

## `GuidanceImport`

GuidanceImport defines how to import guidance documents with optional exclusions and constraints.

Required:

- `reference-id`

Optional:

- `constraints`
- `exclusions`

---

### `constraints` (optional)

Constraints allow policy authors to define ad hoc minimum requirements (e.g., "review at least annually").

- **Type**: `array`
- **Items**: [Constraint]

---

### `exclusions` (optional)

- **Type**: `array`
- **Items**: `string`

---

### `reference-id`

- **Type**: `string`

---

## `CatalogImport`

CatalogImport defines how to import control catalogs with optional exclusions, constraints, and assessment requirement modifications.

Required:

- `reference-id`

Optional:

- `assessment-requirement-modifications`
- `constraints`
- `exclusions`

---

### `assessment-requirement-modifications` (optional)

- **Type**: `array`
- **Items**: [AssessmentRequirementModifier]

---

### `constraints` (optional)

- **Type**: `array`
- **Items**: [Constraint]

---

### `exclusions` (optional)

- **Type**: `array`
- **Items**: `string`

---

### `reference-id`

- **Type**: `string`

---

## `Constraint`

Constraint defines a prescriptive requirement that applies to a specific guidance or control.

Required:

- `id`
- `target-id`
- `text`

---

### `id`

Unique ID for this constraint to enable Layer 5/6 tracking

- **Type**: `string`

---

### `target-id`

Links to the specific Guidance or Control being constrained

- **Type**: `string`

---

### `text`

The prescriptive requirement/constraint text

- **Type**: `string`

---

## `AssessmentRequirementModifier`

AssessmentRequirementModifier allows organizations to customize assessment requirements based on how an organization wants to gather evidence for the objective.

Required:

- `id`
- `modification-rationale`
- `modification-type`
- `target-id`

Optional:

- `applicability`
- `recommendation`
- `text`

---

### `applicability` (optional)

The updated applicability of the assessment requirement

- **Type**: `array`
- **Items**: `string`

---

### `id`

- **Type**: `string`

---

### `modification-rationale`

- **Type**: `string`

---

### `modification-type`

- **Type**: [ModType]

---

### `recommendation` (optional)

The updated recommendation for the assessment requirement

- **Type**: `string`

---

### `target-id`

- **Type**: `string`

---

### `text` (optional)

The updated text of the assessment requirement

- **Type**: `string`

---

## `ModType`

ModType defines the type of modification to the assessment requirement.

- **Type**: `string`

---

