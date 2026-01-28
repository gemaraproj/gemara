---
layout: page
title: Aliases & Base Types
---

## `Contact`

Contact is the contact information for a person or group

`name` **string** _Required_

name is the preferred descriptor for the contact entity

`affiliation` **string**

affiliation is the organization with which the contact entity is associated, such as a team, school, or employer

`email` **[Email](base#email)**

email is the preferred email address to reach the contact

`social` **string**

social is a social media handle or other profile for the contact, such as GitHub

## `Actor`

Actor represents an entity (human or tool) that can perform actions in evaluations

`id` **string** _Required_

id uniquely identifies the actor and allows this entry to be referenced by other elements

`name` **string** _Required_

name is the name of the actor

`type` **[ActorType](base#actortype)** _Required_

type specifies the type of entity interacting in the workflow

`contact` **[Contact](base#contact)**

contact is contact information for the actor

`description` **string**

description provides additional context about the actor

`uri` **string**

uri is a general URI for the actor information

`version` **string**

version is the version of the actor (for tools; if applicable)

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

`description` **string** _Required_

description explains the significance and traits of entries to this category

`id` **string** _Required_

id allows this entry to be referenced by other elements

`title` **string** _Required_

title describes the purpose of this category at a glance

## `Family`

Family represents a logical grouping of guidelines or controls which share a common purpose or function

`description` **string** _Required_

description explains the significance and traits of entries to this entity family

`id` **string** _Required_

id allows this entry to be referenced by other elements

`title` **string** _Required_

title describes the purpose of this family at a glance

