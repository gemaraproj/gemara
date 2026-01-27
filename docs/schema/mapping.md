---
layout: page
title: Mapping
---

## `MappingReference`

MappingReference represents a reference to an external document with full metadata.

Required:

- `id`
- `title`
- `version`

Optional:

- `description`
- `url`

---

### `description` (optional)

description is prose regarding the artifact's purpose or content

- **Type**: `string`

---

### `id`

id allows this entry to be referenced by other elements

- **Type**: `string`

---

### `title`

title describes the purpose of this mapping reference at a glance

- **Type**: `string`

---

### `url` (optional)

url is the path where the artifact may be retrieved; preferrably responds with Gemara-compatible YAML/JSON

- **Type**: `string`

---

### `version`

version is the version identifier of the artifact being mapped to

- **Type**: `string`

---

## `MultiMapping`

MultiMapping represents a mapping to an external reference with one or more entries.

Required:

- `entries`
- `reference-id`

Optional:

- `remarks`

---

### `entries`

entries is a list of mapping entries

- **Type**: `array`
- **Items**: [MappingEntry]

---

### `reference-id`

ReferenceId should reference the corresponding MappingReference id from metadata

- **Type**: `string`

---

### `remarks` (optional)

remarks is prose regarding the mapped artifact or the mapping relationship

- **Type**: `string`

---

## `SingleMapping`

SingleMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.

Required:

- `entry-id`

Optional:

- `reference-id`
- `remarks`
- `strength`

---

### `entry-id`

entry-id is the identifier being mapped to in the referenced artifact

- **Type**: `string`

---

### `reference-id` (optional)

reference-id is the id for a MappingReference entry in the artifact's metadata

- **Type**: `string`

---

### `remarks` (optional)

remarks is prose describing the mapping relationship

- **Type**: `string`

---

### `strength` (optional)

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

- **Type**: `string`

---

## `MappingEntry`

MappingEntry represents a single entry within a mapping

Required:

- `reference-id`

Optional:

- `remarks`
- `strength`

---

### `reference-id`

reference-id is the id for a MappingReference entry in the artifact's metadata

- **Type**: `string`

---

### `remarks` (optional)

remarks is prose describing the mapping relationship

- **Type**: `string`

---

### `strength` (optional)

strength is the author's estimate of how completely the current/source material satisfies the target/reference material;

- **Type**: `string`

---

