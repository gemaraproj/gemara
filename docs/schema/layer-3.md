---
layout: page
title: Layer 3
---

## `Policy`

Policy represents a policy document with metadata, contacts, scope, imports, implementation plan, risks, and adherence requirements.

`adherence` **Adherence** _Required_

Adherence defines evaluation methods, assessment plans, enforcement methods, and non-compliance notifications.

`contacts` **Contacts** _Required_

Contacts defines RACI roles for policy compliance and notification.

`imports` **Imports** _Required_

Imports defines external policies, controls, and guidelines required by this policy.

`metadata` **Metadata** _Required_

Metadata represents common metadata fields shared across all layers

`scope` **Scope** _Required_

Scope defines what is included and excluded from policy applicability.

`title` **string** _Required_


`implementation-plan` **ImplementationPlan**

ImplementationPlan defines when and how the policy becomes active.

`risks` **Risks**

Risks defines mitigated and accepted risks addressed by this policy.

## `Contacts`

Contacts defines RACI roles for policy compliance and notification.

`accountable` **array[Contact]** _Required_

accountable is the person or group accountable for evaluating and enforcing the efficacy of technical controls

`responsible` **array[Contact]** _Required_

responsible is the person or group responsible for implementing controls for technical requirements

`consulted` **array[Contact]**

consulted is an optional person or group who may be consulted for more information about the technical requirements

`informed` **array[Contact]**

informed is an optional person or group who must receive updates about compliance with this policy

## `Scope`

Scope defines what is included and excluded from policy applicability.

`in` **Dimensions** _Required_

Dimensions specify the applicability criteria for a policy

`out` **Dimensions**

Dimensions specify the applicability criteria for a policy

## `Dimensions`

Dimensions specify the applicability criteria for a policy

`geopolitical` **array[string]**

geopolitical is an optional list of geopolitical regions

`groups` **array[string]**


`sensitivity` **array[string]**

sensitivity is an optional list of data classification levels

`technologies` **array[string]**

technologies is an optional list of technology categories or services

`users` **array[string]**

users is an optional list of user roles

## `Imports`

Imports defines external policies, controls, and guidelines required by this policy.

`catalogs` **array[CatalogImport]**


`guidance` **array[GuidanceImport]**


`policies` **array[string]**


## `ImplementationPlan`

ImplementationPlan defines when and how the policy becomes active.

`enforcement-timeline` **ImplementationDetails** _Required_

ImplementationDetails specifies the timeline for policy implementation.

`evaluation-timeline` **ImplementationDetails** _Required_

ImplementationDetails specifies the timeline for policy implementation.

`notification-process` **string**


## `ImplementationDetails`

ImplementationDetails specifies the timeline for policy implementation.

`notes` **string** _Required_


`start` **Datetime** _Required_

Datetime represents an ISO 8601 formatted datetime string

`end` **Datetime**

Datetime represents an ISO 8601 formatted datetime string

## `Risks`

Risks defines mitigated and accepted risks addressed by this policy.

`accepted` **array[AcceptedRisk]**

Accepted risks require rationale (justification) and may include scope. Controls addressing these risks are implicitly identified through threat mappings.

`mitigated` **array[MultiMapping]**

Mitigated risks only need reference-id and risk-id (no justification required)

## `AcceptedRisk`

RiskMapping maps a risk to a reference and optionally includes scope and justification.

`risk` **SingleMapping** _Required_

SingleMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.

`justification` **string**


`scope` **Scope**

Scope and justification are only required for accepted risks (e.g., risk is accepted for TLP:Green and TLP:Clear because they contain non-sensitive data)

## `Adherence`

Adherence defines evaluation methods, assessment plans, enforcement methods, and non-compliance notifications.

`assessment-plans` **array[AssessmentPlan]**


`enforcement-methods` **array[AcceptedMethod]**


`evaluation-methods` **array[AcceptedMethod]**


`non-compliance` **string**


## `AssessmentPlan`

AssessmentPlan defines how a specific assessment requirement is evaluated.

`evaluation-methods` **array[AcceptedMethod]** _Required_


`frequency` **string** _Required_


`id` **string** _Required_


`requirement-id` **string** _Required_


`evidence-requirements` **string**


`parameters` **array[Parameter]**


## `AcceptedMethod`

AcceptedMethod defines a method for evaluation or enforcement.

`type` **string** _Required_


`description` **string**


`executor` **Actor**

Actor represents an entity (human or tool) that can perform actions in evaluations

## `MethodType`

- **Type**: `string`

---

## `Parameter`

Parameter defines a configurable parameter for assessment or enforcement activities.

`description` **string** _Required_


`id` **string** _Required_


`label` **string** _Required_


`accepted-values` **array[string]**


## `GuidanceImport`

GuidanceImport defines how to import guidance documents with optional exclusions and constraints.

`reference-id` **string** _Required_


`constraints` **array[Constraint]**

Constraints allow policy authors to define ad hoc minimum requirements (e.g., "review at least annually").

`exclusions` **array[string]**


## `CatalogImport`

CatalogImport defines how to import control catalogs with optional exclusions, constraints, and assessment requirement modifications.

`reference-id` **string** _Required_


`assessment-requirement-modifications` **array[AssessmentRequirementModifier]**


`constraints` **array[Constraint]**


`exclusions` **array[string]**


## `Constraint`

Constraint defines a prescriptive requirement that applies to a specific guidance or control.

`id` **string** _Required_

Unique ID for this constraint to enable Layer 5/6 tracking

`target-id` **string** _Required_

Links to the specific Guidance or Control being constrained

`text` **string** _Required_

The prescriptive requirement/constraint text

## `AssessmentRequirementModifier`

AssessmentRequirementModifier allows organizations to customize assessment requirements based on how an organization wants to gather evidence for the objective.

`id` **string** _Required_


`modification-rationale` **string** _Required_


`modification-type` **ModType** _Required_

ModType defines the type of modification to the assessment requirement.

`target-id` **string** _Required_


`applicability` **array[string]**

The updated applicability of the assessment requirement

`recommendation` **string**

The updated recommendation for the assessment requirement

`text` **string**

The updated text of the assessment requirement

## `ModType`

ModType defines the type of modification to the assessment requirement.

- **Type**: `string`

---

