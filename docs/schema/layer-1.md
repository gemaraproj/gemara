---
layout: page
title: Layer 1
---

## `GuidanceDocument`

Required:

- `_validateExtensions`
- `document-type`
- `metadata`
- `title`

Optional:

- `exemptions`
- `families`
- `front-matter`
- `guidelines`

---

### `_validateExtensions`

Guidelines that extend other guidelines must be in the same family as the

- **Type**: `object`

---

### `document-type`

- **Type**: [DocumentType]

---

### `exemptions` (optional)

- **Type**: `array`
- **Items**: [Exemption]

---

### `families` (optional)

- **Type**: `array`
- **Items**: [Family]

---

### `front-matter` (optional)

Introductory text for the document to be used during rendering

- **Type**: `string`

---

### `guidelines` (optional)

- **Type**: `array`
- **Items**: [Guideline]

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

### `title`

- **Type**: `string`

---

## `DocumentType`

- **Type**: `string`

---

## `Exemption`

Exemption represents those who are exempt from the full guidance document.

Required:

- `description`
- `reason`

Optional:

- `redirect`

---

### `description`

Description identifies who or what is exempt from the full guidance

- **Type**: `string`

---

### `reason`

Reason explains why the exemption is granted

- **Type**: `string`

---

### `redirect` (optional)

Redirect points to alternative guidelines or controls that should be followed instead

- **Type**: [MultiMapping]

Required if `redirect` is present:

- `entries`
- `reference-id`

Optional:

- `remarks`

---

#### `redirect.entries`

- **Type**: `array`
- **Items**: [MappingEntry]

---

#### `redirect.reference-id`

ReferenceId should reference the corresponding MappingReference id from metadata

- **Type**: `string`

---

#### `redirect.remarks` (optional)

- **Type**: `string`

---

## `Guideline`

Guideline represents a single guideline within a guidance document

Required:

- `family`
- `id`
- `title`

Optional:

- `applicability`
- `extends`
- `guideline-mappings`
- `objective`
- `principle-mappings`
- `rationale`
- `recommendations`
- `see-also`
- `statements`

---

### `applicability` (optional)

Applicability specifies the contexts in which this guideline applies.

- **Type**: `array`
- **Items**: `string`

---

### `extends` (optional)

Extends allows you to add supplemental guidance within a local guidance document

- **Type**: [SingleMapping]

Required if `extends` is present:

- `entry-id`

Optional:

- `reference-id`
- `remarks`

---

#### `extends.entry-id`

- **Type**: `string`

---

#### `extends.reference-id` (optional)

ReferenceId should reference the corresponding MappingReference id from metadata

- **Type**: `string`

---

#### `extends.remarks` (optional)

- **Type**: `string`

---

### `family`

Family id that this guideline belongs to

- **Type**: `string`

---

### `guideline-mappings` (optional)

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `id`

- **Type**: `string`

---

### `objective` (optional)

- **Type**: `string`

---

### `principle-mappings` (optional)

A list for associated key principle ids

- **Type**: `array`
- **Items**: [MultiMapping]

---

### `rationale` (optional)

Rationale provides contextual information to help with development and understanding of

- **Type**: [Rationale]

Required if `rationale` is present:

- `goals`
- `importance`

---

#### `rationale.goals`

- **Type**: `array`
- **Items**: `string`

---

#### `rationale.importance`

- **Type**: `string`

---

### `recommendations` (optional)

Maps to fields commonly seen in controls with implementation guidance

- **Type**: `array`
- **Items**: `string`

---

### `see-also` (optional)

SeeAlso lists related guideline IDs within the same Guidance document.

- **Type**: `array`
- **Items**: `string`

---

### `statements` (optional)

- **Type**: `array`
- **Items**: [Statement]

---

### `title`

- **Type**: `string`

---

## `Statement`

Statement represents a structural sub-requirement within a guideline

Required:

- `id`
- `text`

Optional:

- `recommendations`
- `title`

---

### `id`

- **Type**: `string`

---

### `recommendations` (optional)

- **Type**: `array`
- **Items**: `string`

---

### `text`

- **Type**: `string`

---

### `title` (optional)

- **Type**: `string`

---

## `Rationale`

Rationale provides contextual information to help with development and understanding of

Required:

- `goals`
- `importance`

---

### `goals`

- **Type**: `array`
- **Items**: `string`

---

### `importance`

- **Type**: `string`

---

