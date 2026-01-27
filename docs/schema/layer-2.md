---
layout: page
title: Layer 2
---

## `Catalog`

Required:

- `title`

Optional:

- `capabilities`
- `controls`
- `families`
- `imported-capabilities`
- `imported-controls`
- `imported-threats`
- `metadata`
- `threats`

---

### `capabilities` (optional)

- **Type**: `array`
- **Items**: [Capability]

---

### `controls` (optional)

- **Type**: `array`
- **Items**: [Control]

---

### `families` (optional)

- **Type**: `array`
- **Items**: [Family]

---

### `imported-capabilities` (optional)

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `imported-controls` (optional)

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `imported-threats` (optional)

- **Type**: `array`
- **Items**: [MultiMapping]

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

### `threats` (optional)

- **Type**: `array`
- **Items**: [Threat]

---

### `title`

- **Type**: `string`

---

## `Control`

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

- **Type**: `array`
- **Items**: [AssessmentRequirement]

---

### `family`

Family id that this control belongs to

- **Type**: `string`

---

### `guideline-mappings` (optional)

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `id`

- **Type**: `string`

---

### `objective`

- **Type**: `string`

---

### `threat-mappings` (optional)

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `title`

- **Type**: `string`

---

## `Threat`

Required:

- `capabilities`
- `description`
- `id`
- `title`

Optional:

- `external-mappings`

---

### `capabilities`

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `description`

- **Type**: `string`

---

### `external-mappings` (optional)

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `id`

- **Type**: `string`

---

### `title`

- **Type**: `string`

---

## `Capability`

Required:

- `description`
- `id`
- `title`

---

### `description`

- **Type**: `string`

---

### `id`

- **Type**: `string`

---

### `title`

- **Type**: `string`

---

## `AssessmentRequirement`

Required:

- `applicability`
- `id`
- `text`

Optional:

- `recommendation`

---

### `applicability`

- **Type**: `array`
- **Items**: `string`

---

### `id`

- **Type**: `string`

---

### `recommendation` (optional)

- **Type**: `string`

---

### `text`

- **Type**: `string`

---

