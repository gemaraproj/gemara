---
layout: page
title: Layer 1
---

## `GuidanceCatalog`

GuidanceCatalog represents a concerted documentation effort to help bring about an optimal future without foreknowledge of the implementation details

Required:

- `_validateExtensions`
- `metadata`
- `title`
- `type`

Optional:

- `exemptions`
- `families`
- `front-matter`
- `guidelines`

---

### `_validateExtensions`

guidelines that extend other guidelines must be in the same family as the extended guideline

- **Type**: `object`

---

### `exemptions` (optional)

exemptions provides information about situations where this guidance is not applicable

- **Type**: `array`
- **Items**: [Exemption]

---

### `families` (optional)

families contains a list of guidance families that can be referenced by guidance

- **Type**: `array`
- **Items**: [Family]

---

### `front-matter` (optional)

front-matter provides introductory text for the document to be used during rendering

- **Type**: `string`

---

### `guidelines` (optional)

guidelines is a list of unique guidelines defined by this catalog

- **Type**: `array`
- **Items**: [Guideline]

---

### `metadata`

metadata provides detailed data about this catalog

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

### `title`

title describes the contents of this catalog at a glance

- **Type**: `string`

---

### `type`

type categorizes this document based on the intent of its contents

- **Type**: [GuidanceType]

---

## `GuidanceType`

GuidanceType restricts the possible types that a catalog may be listed as

- **Type**: `string`

---

## `Exemption`

Exemption describes a single scenario where the catalog is not applicable

Required:

- `description`
- `reason`

Optional:

- `redirect`

---

### `description`

description identifies who or what is exempt from the full guidance

- **Type**: `string`

---

### `reason`

reason explains why the exemption is granted

- **Type**: `string`

---

### `redirect` (optional)

redirect points to alternative guidelines or controls that should be followed instead

- **Type**: [MultiMapping]

Required if `redirect` is present:

- `entries`
- `reference-id`

Optional:

- `remarks`

---

#### `redirect.entries`

entries is a list of mapping entries

- **Type**: `array`
- **Items**: [MappingEntry]

---

#### `redirect.reference-id`

ReferenceId should reference the corresponding MappingReference id from metadata

- **Type**: `string`

---

#### `redirect.remarks` (optional)

remarks is prose regarding the mapped artifact or the mapping relationship

- **Type**: `string`

---

## `Guideline`

Guideline provides explanatory context and recommendations for designing optimal outcomes

Required:

- `family`
- `id`
- `objective`
- `title`

Optional:

- `applicability`
- `extends`
- `guideline-mappings`
- `principle-mappings`
- `rationale`
- `recommendations`
- `see-also`
- `statements`
- `vector-mappings`

---

### `applicability` (optional)

applicability specifies the contexts in which this guideline applies

- **Type**: `array`
- **Items**: `string`

---

### `extends` (optional)

extends is an id for a guideline which this guideline adds to, in this document or elsewhere

- **Type**: [SingleMapping]

Required if `extends` is present:

- `entry-id`

Optional:

- `reference-id`
- `remarks`
- `strength`

---

#### `extends.entry-id`

entry-id is the identifier being mapped to in the referenced artifact

- **Type**: `string`

---

#### `extends.reference-id` (optional)

reference-id is the id for a MappingReference entry in the artifact's metadata

- **Type**: `string`

---

#### `extends.remarks` (optional)

remarks is prose describing the mapping relationship

- **Type**: `string`

---

#### `extends.strength` (optional)

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

- **Type**: `string`

---

### `family`

family provides an id to the family that this guideline belongs to

- **Type**: `string`

---

### `guideline-mappings` (optional)

guideline-mappings documents the relationship between this guideline and external guidelines

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `objective`

objective is a unified statement of intent, which may encompass multiple situationally applicable statements

- **Type**: `string`

---

### `principle-mappings` (optional)

principle-mappings documents the relationship between this guideline and one or more principles

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `rationale` (optional)

rationale provides the context for this guideline

- **Type**: [Rationale]

Required if `rationale` is present:

- `goals`
- `importance`

---

#### `rationale.goals`

goals is a list of outcomes this guideline seeks to achieve

- **Type**: `array`
- **Items**: `string`

---

#### `rationale.importance`

importance is an explanation of why this guideline matters

- **Type**: `string`

---

### `recommendations` (optional)

recommendations is a list of non-binding suggestions to aid in evaluation or enforcement of the guideline

- **Type**: `array`
- **Items**: `string`

---

### `see-also` (optional)

see-also lists related guideline IDs within the same GuidanceCatalog

- **Type**: `array`
- **Items**: `string`

---

### `statements` (optional)

statements is a list of structural sub-requirements within a guideline

- **Type**: `array`
- **Items**: [Statement]

---

### `title`

title describes the contents of this guideline

- **Type**: `string`

---

### `vector-mappings` (optional)

vector-mappings documents the relationship between this guideline and one or more vectors

- **Type**: `array`
- **Items**: [MultiMapping]

---

## `Statement`

Statement represents a structural sub-requirement within a guideline;

Required:

- `id`
- `text`

Optional:

- `recommendations`
- `title`

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `recommendations` (optional)

recommendations is a list of non-binding suggestions to aid in evaluation or enforcement of the statement

- **Type**: `array`
- **Items**: `string`

---

### `text`

text is the body of this statement

- **Type**: `string`

---

### `title` (optional)

title describes the contents of this statement

- **Type**: `string`

---

## `Rationale`

Rationale provides a structured way to communicate a guideline author's intent

Required:

- `goals`
- `importance`

---

### `goals`

goals is a list of outcomes this guideline seeks to achieve

- **Type**: `array`
- **Items**: `string`

---

### `importance`

importance is an explanation of why this guideline matters

- **Type**: `string`

---

