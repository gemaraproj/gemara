---
layout: page
title: Mapping Document Guide
description: Step-by-step guide to creating Gemara-compatible mapping documents for cross-artifact alignment
---

## What This Is

This guide walks through creating a **mapping document** using the [Gemara](https://gemara.openssf.org/) project. A mapping document expresses how entries in one artifact (e.g., your guidance or control catalog) align to entries in another artifact (e.g., OWASP Top 10, NIST, or another framework).

In technical terms:
* A **Mapping Document** is a dedicated artifact that declares a **source** and **target** and lists **mappings**: each mapping links a source entry to a target entry with a **relationship** (e.g., `supports`, `implements`, `equivalent`).
* **Mapping references** in the document’s metadata list the artifacts involved; `source-reference` and `target-reference` point to which of those artifacts are “from” and “to” for this alignment.
* Per [ADR 0014 (Mapping Boundaries)](https://github.com/gemaraproj/gemara/pull/313), cross-artifact alignment is expressed in this separate document rather than inline on catalog entries, so catalogs stay focused on definitions and design rationale while mapping quality and direction live in the mapping document.

This exercise produces a single, auditable artifact that describes how your guidance or controls map to external standards.

## Why a Separate Mapping Document?

Catalogs (guidance, controls, threats) can declare **inline** relationships to other artifacts—e.g., a control’s `threats` or a guideline’s `vectors`—to capture **design rationale** (what this entry is intended to address). Cross-framework alignment that grows large (e.g., many guidelines mapped to many external controls) is better expressed in a **#MappingDocument**, which:

* Keeps catalog files from growing unbounded.
* Lets you state **relationship** and optional **strength** per mapping in one place.
* Makes it clear that the mapping is an **assertion** by the mapping author, in a specific direction (source → target).

Schema details are in [PR #313](https://github.com/gemaraproj/gemara/pull/313); the mapping document conforms to **#MappingDocument** in `mapping.cue`.

## Optional Workflow

Have at least one **source** artifact (e.g., a [Guidance Catalog](../guidance/guidance-example.yaml) or [Control Catalog](../controls/control-catalog.yaml)) and a **target** artifact (e.g., OWASP Top 10, NIST, or another framework). You will reference both in the mapping document’s `metadata.mapping-references` and then point `source-reference` and `target-reference` at them by id.

## Walkthrough

### Step 0: Choose Source and Target Artifacts

Decide which artifact you are mapping **from** (source) and which you are mapping **to** (target). Each must be identifiable by a stable **id** (e.g., catalog `metadata.id` or a known framework id like `OWASP`).

**Example:** Map the [Secure Software Development Guidance](../guidance/guidance-example.yaml) (id `ORG.SSD.001`) to OWASP Top 10 2025 (id `OWASP`).

### Step 1: Setting Up Metadata and Mapping References

Declare the mapping document’s metadata and list every artifact that participates in this mapping (both source and target) under `metadata.mapping-references`. Each reference needs an `id`, `title`, `version`, and optionally `description` and `url`.

| Field | What It Is | Why |
|-------|------------|-----|
| `title` | Display name for the mapping document (top-level) | Human-readable label for reports and tooling |
| `metadata.id` | Unique id for this mapping document | Identifies the mapping artifact itself |
| `metadata.description` | Summary of what this mapping expresses | Explains scope and intent |
| `metadata.mapping-references` | List of source and target artifacts | Each must have `id`, `title`, `version`; `source-reference` and `target-reference` refer to these ids |

**Example (YAML):**

```yaml
title: Secure Software Development Guidance to OWASP Top 10
description: |
  Alignment from our internal secure development guidelines (ORG.SSD) to OWASP Top 10 (2025)
  for traceability and audit.

metadata:
  id: ORG.SSD.OWASP-MAP.001
  version: "1.0.0"
  description: Mapping from ORG.SSD guidance catalog to OWASP Top 10 Web Application Security Risks (2025).
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: ORG.SSD.001
      title: Secure Software Development Guidance
      version: "1.0.0"
      description: Internal secure development and supply chain security guidelines.
      url: https://example.org/guidance/guidance-example.yaml
    - id: OWASP
      title: OWASP Top 10
      version: "2025"
      description: OWASP Top 10 Web Application Security Risks (2025)
      url: https://owasp.org/Top10/2025
```

### Step 2: Set Source and Target References

Set `source-reference` and `target-reference` to the **ids** of the artifacts you are mapping from and to. Those ids must match an `id` in `metadata.mapping-references`.

| Field | What It Is | Why |
|-------|------------|-----|
| `source-reference.reference-id` | Id of the artifact you are mapping **from** | Must equal one of the `mapping-references[].id` values |
| `target-reference.reference-id` | Id of the artifact you are mapping **to** | Same constraint |

**Example (YAML):**

```yaml
source-reference:
  reference-id: ORG.SSD.001

target-reference:
  reference-id: OWASP
```

### Step 3: Define Mappings

Each item in `mappings` is one **atomic** relationship: one source entry → one target entry (or a no-match). Required fields:

| Field | Required | Description |
|-------|----------|-------------|
| `source` | Yes | **EntryReference**: `entry-id` (e.g., guideline or control id) and `entry-type` (e.g., `Guideline`, `Control`) |
| `target` | No (omit for `no-match`) | **EntryReference**: `entry-id` and `entry-type` in the target artifact |
| `relationship` | Yes | One of the relationship types below |
| `remarks` | No | Prose describing the mapping |

**Entry types** (per schema): `Guideline`, `Statement`, `Control`, `Assessment Requirement`, or a custom string.

**Relationship types** (per schema): `implements`, `implemented-by`, `supports`, `supported-by`, `equivalent`, `subsumes`, `no-match`, `relates-to`.

**Example (YAML):**

```yaml
mappings:
  - source:
      entry-id: ORG.SSD.GL01
      entry-type: Guideline
    target:
      entry-id: A03:2025
      entry-type: Control
    relationship: supports
    remarks: Immutable image references reduce software supply chain and tampering risks.
  - source:
      entry-id: ORG.SSD.GL01
      entry-type: Guideline
    target:
      entry-id: A08:2025
      entry-type: Control
    relationship: supports
    remarks: Digest-based references support software or data integrity.
  - source:
      entry-id: ORG.SSD.GL02
      entry-type: Guideline
    target:
      entry-id: A01:2025
      entry-type: Control
    relationship: supports
    remarks: Branch protection and review support access control.
```

For **no-match** (source entry has no counterpart in the target), omit `target` and use `relationship: no-match`:

```yaml
  - source:
      entry-id: ORG.SSD.GL99
      entry-type: Guideline
    relationship: no-match
    remarks: No direct OWASP equivalent; organization-specific.
```

### Step 4: Validate Against the Schema

Validate the mapping document with CUE. The **#MappingDocument** definition is in `mapping.cue` as introduced in [PR #313](https://github.com/gemaraproj/gemara/pull/313) (branch `jpower432:mapping-artifact`).

**Validation command (from repo root, when schema is available):**

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#MappingDocument' mapping.cue docs/tutorials/mapping/your-mapping-document.yaml
```

If using the published module:

```bash
cue vet -c -d '#MappingDocument' github.com/gemaraproj/gemara@latest your-mapping-document.yaml
```

Fix any reported errors (e.g., missing required fields, `source-reference` or `target-reference` not in `mapping-references`, invalid `relationship` or `entry-type`).

### Step 5: Assemble the Full Document and Validate

Combine `title`, `description`, `metadata` (with `mapping-references`), `source-reference`, `target-reference`, and `mappings` into a single YAML file. A complete, schema-oriented example that maps the [guidance example](../guidance/guidance-example.yaml) to OWASP Top 10 (2025) is in [guidance-to-owasp-mapping.yaml](guidance-to-owasp-mapping.yaml) in this directory.

**Validate** from the repo root (once the PR #313 schema is in your tree):

```bash
cue vet -c -d '#MappingDocument' mapping.cue docs/tutorials/mapping/guidance-to-owasp-mapping.yaml
```

## Summary: From Catalogs to Mapping Document

| From your catalogs | Use in the mapping document |
|--------------------|-----------------------------|
| Source catalog id (e.g., `ORG.SSD.001`) | `metadata.mapping-references[].id` and `source-reference.reference-id` |
| Target framework id (e.g., `OWASP`) | `metadata.mapping-references[].id` and `target-reference.reference-id` |
| Guideline/control/entry ids | `mappings[].source.entry-id` and `mappings[].target.entry-id` |
| Intent of the alignment | `mappings[].relationship` and optional `mappings[].remarks` |

## What's Next

Use the mapping document alongside your guidance and control catalogs: tools can consume it to report coverage, traceability, or gaps (e.g., entries with `no-match`). Layer 1 and Layer 2 catalogs keep **inline** fields (e.g., `vectors`, `guidelines`, `threats`) for design rationale; the mapping document remains the place for cross-artifact alignment and strength when the schema supports it.

See [PR #313](https://github.com/gemaraproj/gemara/pull/313) and ADR 0014 for the mapping boundary decision, and the [Gemara schema documentation](https://gemara.openssf.org/schema/) for the full specification once the mapping schema is released.
