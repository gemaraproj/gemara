---
layout: page
title: Layer 2
---

## `ControlCatalog`

ControlCatalog describes a set of related controls and relevant metadata

Required:

- `metadata`
- `title`

Optional:

- `controls`
- `families`
- `imported-controls`

---

### `controls` (optional)

controls is a list of unique controls defined by this catalog

- **Type**: `array`
- **Items**: [Control]

---

### `families` (optional)

families contains a list of control families that can be referenced by controls

- **Type**: `array`
- **Items**: [Family]

---

### `imported-controls` (optional)

imported-controls is a list of controls from another source which are included as part of this document

- **Type**: `array`
- **Items**: [MultiMapping]

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

## `Control`

Control describes a safeguard or countermeasure with a clear objective and assessment requirements

Required:

- `assessment-requirements`
- `family`
- `id`
- `objective`
- `title`

Optional:

- `guideline-mappings`
- `threat-mappings`

---

### `assessment-requirements`

assessment-requirements is a list of requirements that must be verified to confirm the control objective has been met

- **Type**: `array`
- **Items**: [AssessmentRequirement]

---

### `family`

family references by id a catalog control family that this control belongs to

- **Type**: `string`

---

### `guideline-mappings` (optional)

guideline-mappings documents relationships betwen this control and Layer 1 guideline artifacts

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `objective`

objective is a unified statement of intent, which may encompass multiple situationally applicable requirements

- **Type**: `string`

---

### `threat-mappings` (optional)

threat-mappings documents relationships betwen this control and Layer 2 threat artifacts

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `title`

title describes the purpose of this control at a glance

- **Type**: `string`

---

## `AssessmentRequirement`

AssessmentRequirement describes a tightly scoped, verifiable condition that must be satisfied and confirmed by an evaluator

Required:

- `applicability`
- `id`
- `text`

Optional:

- `recommendation`

---

### `applicability`

applicability is a list of strings describing the situations where this text functions as a requirement for its parent control

- **Type**: `array`
- **Items**: `string`

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `recommendation` (optional)

recommendation provides readers with non-binding suggestions to aid in evaluation or enforcement of the requirement

- **Type**: `string`

---

### `text`

text is the body of the requirement, typically written as a MUST condition

- **Type**: `string`

---

## `ThreatCatalog`

ThreatCatalog describes a set of topically-associated threats

Required:

- `metadata`
- `title`

Optional:

- `capabilities`
- `imported-capabilities`
- `imported-threats`
- `threats`

---

### `capabilities` (optional)

capabilities is a list of capabilities that make up the system being assessed

- **Type**: `array`
- **Items**: [Capability]

---

### `imported-capabilities` (optional)

imported-capabilities is a list of capabilities from another source which are included as part of this document

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `imported-threats` (optional)

imported-threats is a list of threats from another source which are included as part of this document

- **Type**: `array`
- **Items**: [MultiMapping]

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

### `threats` (optional)

threats is a list of threats defined by this catalog

- **Type**: `array`
- **Items**: [Threat]

---

### `title`

title describes the purpose of this catalog at a glance

- **Type**: `string`

---

## `Threat`

Threat describes a specifically-scoped opportunity for a negative impact to the organization

Required:

- `capabilities`
- `description`
- `id`
- `title`

Optional:

- `actors`
- `external-mappings`

---

### `actors` (optional)

actors describes the relevant internal or external threat actors

- **Type**: `array`
- **Items**: [Actor]

---

### `capabilities`

capabilities documents the relationship between this threat and a system capability

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `description`

description provides a detailed explanation of an opportunity for negative impact

- **Type**: `string`

---

### `external-mappings` (optional)

external-mappings documents relationships between this threat and any other artifacts

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `title`

title describes this threat at a glance

- **Type**: `string`

---

## `Capability`

Capability describes a system capability such as a feature, component or object.

Required:

- `description`
- `id`
- `title`

---

### `description`

description provides a detailed overview of this capability

- **Type**: `string`

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `title`

title describes this capability at a glance

- **Type**: `string`

---

