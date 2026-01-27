---
layout: page
title: Aliases & Base Types
---

## `Contact`

Contact is the contact information for a person or group

Required:

- `name`

Optional:

- `affiliation`
- `email`
- `social`

---

### `affiliation` (optional)

affiliation is the organization with which the contact entity is associated, such as a team, school, or employer

- **Type**: `string`

---

### `email` (optional)

email is the preferred email address to reach the contact

- **Type**: [Email]

---

### `name`

name is the preferred descriptor for the contact entity

- **Type**: `string`

---

### `social` (optional)

social is a social media handle or other profile for the contact, such as GitHub

- **Type**: `string`

---

## `Actor`

Actor represents an entity (human or tool) that can perform actions in evaluations

Required:

- `id`
- `name`
- `type`

Optional:

- `contact`
- `description`
- `uri`
- `version`

---

### `contact` (optional)

contact is contact information for the actor

- **Type**: [Contact]

Required if `contact` is present:

- `name`

Optional:

- `affiliation`
- `email`
- `social`

---

#### `contact.affiliation` (optional)

affiliation is the organization with which the contact entity is associated, such as a team, school, or employer

- **Type**: `string`

---

#### `contact.email` (optional)

email is the preferred email address to reach the contact

- **Type**: [Email]

---

#### `contact.name`

name is the preferred descriptor for the contact entity

- **Type**: `string`

---

#### `contact.social` (optional)

social is a social media handle or other profile for the contact, such as GitHub

- **Type**: `string`

---

### `description` (optional)

description provides additional context about the actor

- **Type**: `string`

---

### `id`

id uniquely identifies the actor and allows this entry to be referenced by other elements

- **Type**: `string`

---

### `name`

name is the name of the actor

- **Type**: `string`

---

### `type`

type specifies the type of entity interacting in the workflow

- **Type**: [ActorType]

---

### `uri` (optional)

uri is a general URI for the actor information

- **Type**: `string`

---

### `version` (optional)

version is the version of the actor (for tools; if applicable)

- **Type**: `string`

---

## `ActorType`

ActorType specifies what entity is interacting in the workflow

- **Type**: `string`

---

## `Email`

Email represents a validated email address pattern

- **Type**: `string`
- **Value**: `^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$`

---

## `Datetime`

Datetime represents an ISO 8601 formatted datetime string

- **Type**: `string`
- **Format**: `date`
- **Value**: `^\d{4}-\d{2}-\d{2}$`

---

## `Date`

Date represents a date string (ISO 8601 date format)

- **Type**: `string`
- **Format**: `date`
- **Value**: `^\d{4}-\d{2}-\d{2}$`

---

## `Category`

Category represents a category used for applicability or classification

Required:

- `description`
- `id`
- `title`

---

### `description`

description explains the significance and traits of entries to this category

- **Type**: `string`

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `title`

title describes the purpose of this category at a glance

- **Type**: `string`

---

## `Family`

Family represents a logical grouping of guidelines or controls which share a common purpose or function

Required:

- `description`
- `id`
- `title`

---

### `description`

description explains the significance and traits of entries to this entity family

- **Type**: `string`

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `title`

title describes the purpose of this family at a glance

- **Type**: `string`

---

