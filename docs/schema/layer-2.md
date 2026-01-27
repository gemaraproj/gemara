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

- **Type**: `array`
- **Items**: [Category]

---

#### `metadata.author`

Actor represents an entity (human or tool) that can perform actions in evaluations.

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

Contact provides contact information for the actor.

- **Type**: [Contact]

Required if `metadata.author.contact` is present:

- `name`

Optional:

- `affiliation`
- `email`
- `social`

---

###### `metadata.author.contact.affiliation` (optional)

The entity with which the contact is affiliated, such as a school or employer.

- **Type**: `string`

---

###### `metadata.author.contact.email` (optional)

A preferred email address to reach the contact.

- **Type**: [Email]

---

###### `metadata.author.contact.name`

The contact person's name.

- **Type**: `string`

---

###### `metadata.author.contact.social` (optional)

A social media handle or profile for the contact.

- **Type**: `string`

---

##### `metadata.author.description` (optional)

Description provides additional context about the actor.

- **Type**: `string`

---

##### `metadata.author.id`

Id uniquely identifies the actor.

- **Type**: `string`

---

##### `metadata.author.name`

Name provides the name of the actor.

- **Type**: `string`

---

##### `metadata.author.type`

Type specifies the type of entity interacting in the workflow.

- **Type**: [ActorType]

---

##### `metadata.author.uri` (optional)

Uri provides a general URI for the actor information.

- **Type**: `string`

---

##### `metadata.author.version` (optional)

Version specifies the version of the actor (if applicable, e.g., for tools).

- **Type**: `string`

---

#### `metadata.date` (optional)

- **Type**: [Date]

---

#### `metadata.description`

- **Type**: `string`

---

#### `metadata.draft` (optional)

- **Type**: `boolean`

---

#### `metadata.id`

- **Type**: `string`

---

#### `metadata.lexicon` (optional)

- **Type**: `string`

---

#### `metadata.mapping-references` (optional)

- **Type**: `array`
- **Items**: [MappingReference]

---

#### `metadata.version` (optional)

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

