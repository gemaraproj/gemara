---
layout: page
title: Metadata
---

## `Metadata`

Metadata represents common metadata fields shared across all layers

Required:

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

### `applicability-categories` (optional)

applicability-categories is a list of categories used to classify within this artifact to specify scope

- **Type**: `array`
- **Items**: [Category]

---

### `author`

author is the person or group primarily responsible for this artifact

- **Type**: [Actor]

Required if `author` is present:

- `id`
- `name`
- `type`

Optional:

- `contact`
- `description`
- `uri`
- `version`

---

#### `author.contact` (optional)

contact is contact information for the actor

- **Type**: [Contact]

Required if `author.contact` is present:

- `name`

Optional:

- `affiliation`
- `email`
- `social`

---

##### `author.contact.affiliation` (optional)

affiliation is the organization with which the contact entity is associated, such as a team, school, or employer

- **Type**: `string`

---

##### `author.contact.email` (optional)

email is the preferred email address to reach the contact

- **Type**: [Email]

---

##### `author.contact.name`

name is the preferred descriptor for the contact entity

- **Type**: `string`

---

##### `author.contact.social` (optional)

social is a social media handle or other profile for the contact, such as GitHub

- **Type**: `string`

---

#### `author.description` (optional)

description provides additional context about the actor

- **Type**: `string`

---

#### `author.id`

id uniquely identifies the actor and allows this entry to be referenced by other elements

- **Type**: `string`

---

#### `author.name`

name is the name of the actor

- **Type**: `string`

---

#### `author.type`

type specifies the type of entity interacting in the workflow

- **Type**: [ActorType]

---

#### `author.uri` (optional)

uri is a general URI for the actor information

- **Type**: `string`

---

#### `author.version` (optional)

version is the version of the actor (for tools; if applicable)

- **Type**: `string`

---

### `date` (optional)

date is the publication or effective date of this artifact

- **Type**: [Date]

---

### `description`

description provides a high-level summary of the artifact's purpose and scope

- **Type**: `string`

---

### `draft` (optional)

draft indicates whether this artifact is a pre-release version; open to modification

- **Type**: `boolean`

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `lexicon` (optional)

lexicon is a URI pointing to a controlled vocabulary or glossary relevant to this artifact

- **Type**: `string`

---

### `mapping-references` (optional)

mapping-references is a list of external documents referenced within this artifact

- **Type**: `array`
- **Items**: [MappingReference]

---

### `version` (optional)

version is the version identifier of this artifact

- **Type**: `string`

---

