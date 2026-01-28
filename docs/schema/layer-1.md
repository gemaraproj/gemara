---
layout: page
title: Layer 1
---

## `GuidanceCatalog`

GuidanceCatalog represents a concerted documentation effort to help bring about an optimal future without foreknowledge of the implementation details

`_validateExtensions` **object** _Required_

guidelines that extend other guidelines must be in the same family as the extended guideline

`metadata` **[Metadata](metadata#metadata)** _Required_

metadata provides detailed data about this catalog

`title` **string** _Required_

title describes the contents of this catalog at a glance

`type` **[GuidanceType](layer-1#guidancetype)** _Required_

type categorizes this document based on the intent of its contents

`exemptions` **array[[Exemption](layer-1#exemption)]**

exemptions provides information about situations where this guidance is not applicable

`families` **array[[Family](base#family)]**

families contains a list of guidance families that can be referenced by guidance

`front-matter` **string**

front-matter provides introductory text for the document to be used during rendering

`guidelines` **array[[Guideline](layer-1#guideline)]**

guidelines is a list of unique guidelines defined by this catalog

## `GuidanceType`

GuidanceType restricts the possible types that a catalog may be listed as

- **Type**: `string`

---

## `Exemption`

Exemption describes a single scenario where the catalog is not applicable

`description` **string** _Required_

description identifies who or what is exempt from the full guidance

`reason` **string** _Required_

reason explains why the exemption is granted

`redirect` **[MultiMapping](mapping#multimapping)**

redirect points to alternative guidelines or controls that should be followed instead

## `Guideline`

Guideline provides explanatory context and recommendations for designing optimal outcomes

`family` **string** _Required_

family provides an id to the family that this guideline belongs to

`id` **string** _Required_

id allows this entry to be referenced by other elements

`objective` **string** _Required_

objective is a unified statement of intent, which may encompass multiple situationally applicable statements

`title` **string** _Required_

title describes the contents of this guideline

`applicability` **array[string]**

applicability specifies the contexts in which this guideline applies

`extends` **[SingleMapping](mapping#singlemapping)**

extends is an id for a guideline which this guideline adds to, in this document or elsewhere

`guideline-mappings` **array[[MultiMapping](mapping#multimapping)]**

guideline-mappings documents the relationship between this guideline and external guidelines

`principle-mappings` **array[[MultiMapping](mapping#multimapping)]**

principle-mappings documents the relationship between this guideline and one or more principles

`rationale` **[Rationale](layer-1#rationale)**

rationale provides the context for this guideline

`recommendations` **array[string]**

recommendations is a list of non-binding suggestions to aid in evaluation or enforcement of the guideline

`see-also` **array[string]**

see-also lists related guideline IDs within the same GuidanceCatalog

`statements` **array[[Statement](layer-1#statement)]**

statements is a list of structural sub-requirements within a guideline

`vector-mappings` **array[[MultiMapping](mapping#multimapping)]**

vector-mappings documents the relationship between this guideline and one or more vectors

## `Statement`

Statement represents a structural sub-requirement within a guideline;

`id` **string** _Required_

id allows this entry to be referenced by other elements

`text` **string** _Required_

text is the body of this statement

`recommendations` **array[string]**

recommendations is a list of non-binding suggestions to aid in evaluation or enforcement of the statement

`title` **string**

title describes the contents of this statement

## `Rationale`

Rationale provides a structured way to communicate a guideline author's intent

`goals` **array[string]** _Required_

goals is a list of outcomes this guideline seeks to achieve

`importance` **string** _Required_

importance is an explanation of why this guideline matters

