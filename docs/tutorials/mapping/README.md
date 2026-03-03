# Mapping document tutorial

This directory contains an example **Mapping Document** that aligns the [Secure Software Development Guidance](../guidance/guidance-example.yaml) to external standards.

**Step-by-step guide:** [Mapping Document Guide](mapping-document-guide.md) — walkthrough for creating a mapping document (structure, metadata, source/target references, mappings, validation).

## Overview

Per [ADR 0014 (Mapping Boundaries)](https://github.com/gemaraproj/gemara/pull/313), cross-artifact alignment is expressed in a dedicated **#MappingDocument** artifact rather than inline on catalog entries. The mapping document:

- Declares **source** and **target** artifacts via `metadata.mapping-references`
- Uses `source-reference` and `target-reference` to point at those artifacts by id
- Lists **mappings**: each mapping links a source entry (e.g. guideline) to a target entry (e.g. control) with a `relationship` (e.g. `supports`, `implements`, `equivalent`)

## File

| File | Description |
|------|-------------|
| [guidance-to-owasp-mapping.yaml](guidance-to-owasp-mapping.yaml) | Maps guidelines from the guidance example to OWASP Top 10 (2025) controls |

## Schema and validation

The mapping document conforms to **#MappingDocument** in `mapping.cue` as introduced in [PR #313](https://github.com/gemaraproj/gemara/pull/313) (branch `jpower432:mapping-artifact`). To validate once that schema is available:

```bash
cue vet -c -d '#MappingDocument' mapping.cue docs/tutorials/mapping/guidance-to-owasp-mapping.yaml
```

## Relationship types

Allowed values for `relationship` (per schema) include: `implements`, `implemented-by`, `supports`, `supported-by`, `equivalent`, `subsumes`, `no-match`, `relates-to`.

## See also

- [Guidance catalog guide](../guidance/guidance-guide.md) for the source guidance example
- [PR #313 – Mapping boundaries](https://github.com/gemaraproj/gemara/pull/313) for the schema and ADR 0014
