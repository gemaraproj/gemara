---
layout: page
title: Layer 2
---

## `ControlCatalog`

ControlCatalog describes a set of related controls and relevant metadata

`metadata` **Metadata** _Required_

metadata provides detailed data about this catalog

`title` **string** _Required_

title describes the contents of this catalog at a glance

`controls` **array[Control]**

controls is a list of unique controls defined by this catalog

`families` **array[Family]**

families contains a list of control families that can be referenced by controls

`imported-controls` **array[MultiMapping]**

imported-controls is a list of controls from another source which are included as part of this document

## `Control`

Control describes a safeguard or countermeasure with a clear objective and assessment requirements

`assessment-requirements` **array[AssessmentRequirement]** _Required_

assessment-requirements is a list of requirements that must be verified to confirm the control objective has been met

`family` **string** _Required_

family references by id a catalog control family that this control belongs to

`id` **string** _Required_

id allows this entry to be referenced by other elements

`objective` **string** _Required_

objective is a unified statement of intent, which may encompass multiple situationally applicable requirements

`title` **string** _Required_

title describes the purpose of this control at a glance

`guideline-mappings` **array[MultiMapping]**

guideline-mappings documents relationships betwen this control and Layer 1 guideline artifacts

`threat-mappings` **array[MultiMapping]**

threat-mappings documents relationships betwen this control and Layer 2 threat artifacts

## `AssessmentRequirement`

AssessmentRequirement describes a tightly scoped, verifiable condition that must be satisfied and confirmed by an evaluator

`applicability` **array[string]** _Required_

applicability is a list of strings describing the situations where this text functions as a requirement for its parent control

`id` **string** _Required_

id allows this entry to be referenced by other elements

`text` **string** _Required_

text is the body of the requirement, typically written as a MUST condition

`recommendation` **string**

recommendation provides readers with non-binding suggestions to aid in evaluation or enforcement of the requirement

## `ThreatCatalog`

ThreatCatalog describes a set of topically-associated threats

`metadata` **Metadata** _Required_

metadata provides detailed data about this catalog

`title` **string** _Required_

title describes the purpose of this catalog at a glance

`capabilities` **array[Capability]**

capabilities is a list of capabilities that make up the system being assessed

`imported-capabilities` **array[MultiMapping]**

imported-capabilities is a list of capabilities from another source which are included as part of this document

`imported-threats` **array[MultiMapping]**

imported-threats is a list of threats from another source which are included as part of this document

`threats` **array[Threat]**

threats is a list of threats defined by this catalog

## `Threat`

Threat describes a specifically-scoped opportunity for a negative impact to the organization

`capabilities` **array[MultiMapping]** _Required_

capabilities documents the relationship between this threat and a system capability

`description` **string** _Required_

description provides a detailed explanation of an opportunity for negative impact

`id` **string** _Required_

id allows this entry to be referenced by other elements

`title` **string** _Required_

title describes this threat at a glance

`actors` **array[Actor]**

actors describes the relevant internal or external threat actors

`external-mappings` **array[MultiMapping]**

external-mappings documents relationships between this threat and any other artifacts

## `Capability`

Capability describes a system capability such as a feature, component or object.

`description` **string** _Required_

description provides a detailed overview of this capability

`id` **string** _Required_

id allows this entry to be referenced by other elements

`title` **string** _Required_

title describes this capability at a glance

