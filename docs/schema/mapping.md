---
layout: page
title: Mapping
---

## `MappingReference`

MappingReference represents a reference to an external document with full metadata.

`id` **string** _Required_

id allows this entry to be referenced by other elements

`title` **string** _Required_

title describes the purpose of this mapping reference at a glance

`version` **string** _Required_

version is the version identifier of the artifact being mapped to

`description` **string**

description is prose regarding the artifact's purpose or content

`url` **string**

url is the path where the artifact may be retrieved; preferrably responds with Gemara-compatible YAML/JSON

## `MultiMapping`

MultiMapping represents a mapping to an external reference with one or more entries.

`entries` **array[[MappingEntry](mapping#mappingentry)]** _Required_

entries is a list of mapping entries

`reference-id` **string** _Required_

ReferenceId should reference the corresponding MappingReference id from metadata

`remarks` **string**

remarks is prose regarding the mapped artifact or the mapping relationship

## `SingleMapping`

SingleMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.

`entry-id` **string** _Required_

entry-id is the identifier being mapped to in the referenced artifact

`reference-id` **string**

reference-id is the id for a MappingReference entry in the artifact's metadata

`remarks` **string**

remarks is prose describing the mapping relationship

`strength` **string**

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

## `MappingEntry`

MappingEntry represents a single entry within a mapping

`reference-id` **string** _Required_

reference-id is the id for a MappingReference entry in the artifact's metadata

`remarks` **string**

remarks is prose describing the mapping relationship

`strength` **string**

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

