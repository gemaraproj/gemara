---
layout: page
title: Mapping Document Guide
description: Step-by-step guide to creating Gemara-compatible mapping documents
---

## What This Is

This guide walks through creating a **Mapping Document** using the [Gemara](https://gemara.openssf.org/) project, building on the guidance catalog you created in the [Guidance Catalog Guide](../guidance/guidance-guide).

A mapping document captures the user's intent for how entries in a **source** artifact relate to entries in a **target** artifact. It is the place to express alignment between independently authored catalogs (e.g., your guidance to OWASP, or controls to regulations) in a single, directed way.

In technical terms:
* **Source reference** is the artifact you map *from* (e.g., your guidance catalog). Its `reference-id` must match an id in `metadata.mapping-references`.
* **Target reference** is the artifact you map *to* (e.g., OWASP Top 10). Its `reference-id` must also match an id in `metadata.mapping-references`.
* **Mappings** are one or more atomic relationships: each links a source entry to an optional target entry and a **relationship** type. For `no-match`, the source has no counterpart in the target and `target` is omitted.
* **Relationship types** (see [mapping.cue](https://github.com/gemaraproj/gemara/blob/main/mapping.cue)): `implements`, `implemented-by`, `supports`, `supported-by`, `equivalent`, `subsumes`, `no-match`, `relates-to`.
* **Entry reference**: `entry-id` (id within the referenced artifact) and `entry-type` (`Guideline`, `Statement`, `Control`, `AssessmentRequirement`, or `Vector`).

This exercise produces a mapping document that downstream tools and policies can use to understand how two artifacts align.

## Optional Workflow

Complete the [Guidance Catalog Guide](../guidance/guidance-guide) and have a guidance catalog (e.g., [guidance-example.yaml](../guidance/guidance-example.yaml)) with `mapping-references` that include the target framework you want to map to (e.g., OWASP). You will reference that guidance catalog as the source and the external framework as the target.

## Walkthrough

We use the **Secure Software Development Guidance** (ORG.SSD) from the guidance tutorial and map its guidelines to **OWASP Top 10**. The same pattern applies when mapping controls to regulations or other frameworks.

### Step 0: Define Scope and References

Decide which two artifacts you are mapping: **source** (the one you map *from*) and **target** (the one you map *to*). Reuse the same catalog ids as in your guidance catalog’s `mapping-references` so the mapping document can refer to them.

**Leverage existing resources:** Your guidance catalog already declares `mapping-references` (e.g., OWASP). The mapping document must declare **both** the source and target in its own `metadata.mapping-references`; the source is your guidance catalog (e.g., ORG-SSD) and the target is the external framework (e.g., OWASP).

We continue with the Secure Software Development Guidance as source and OWASP Top 10 as target.

### Step 1: Setting Up Metadata

Declare your mapping document and mapping references. Key fields:

| Field                               | What It Is                                                   | Why                                                                                       |
|-------------------------------------|--------------------------------------------------------------|-------------------------------------------------------------------------------------------|
| `title`                             | Display name for the mapping document (top-level)            | Human-readable label in reports and tooling                                               |
| `metadata.id`                       | Unique identifier for this mapping document                  | Used when other documents or tools reference this mapping                                 |
| `metadata.type`                     | `MappingDocument`                                           | Required by schema; identifies the Gemara artifact type                                   |
| `metadata.description`              | High-level summary of what is being mapped and why           | Required by schema; clarifies intent                                                      |
| `metadata.mapping-references`       | List of source and target artifacts (by id)                 | **Required** for MappingDocument; `source-reference` and `target-reference` must match ids here |

> **Note:** Both `source-reference.reference-id` and `target-reference.reference-id` must appear in `metadata.mapping-references`.

**Example (YAML):**

```yaml
title: Secure Software Development Guidance to OWASP Top 10
metadata:
  id: SSD-OWASP-MAP-001
  version: "1.0.0"
  type: MappingDocument
  gemara-version: "0.20.0"
  description: >
    Maps Secure Software Development Guidance guidelines to OWASP Top 10
    categories. Minimal example for tutorials; relationship types are relates-to.
  author:
    id: gemara-example
    name: Gemara Example Author
    type: Human
  mapping-references:
    - id: ORG-SSD
      title: Secure Software Development Guidance
      version: "1.0.0"
      url: "file://guidance-example.yaml"
    - id: OWASP
      title: OWASP Top 10
      version: "2021"
      url: "https://owasp.org/Top10"
```

### Step 2: Source and Target References

Set **source-reference** and **target-reference**. Each uses a `reference-id` that must match an id in `metadata.mapping-references`. Optionally add top-level **remarks** to describe the mapping.

| Field               | What It Is                                                                 |
|---------------------|----------------------------------------------------------------------------|
| `source-reference`  | Artifact you are mapping *from*; use `reference-id` (and optional `remarks`) |
| `target-reference`  | Artifact you are mapping *to*; use `reference-id` (and optional `remarks`) |
| `remarks`           | Optional prose about this mapping document as a whole                      |

**Example (YAML):**

```yaml
source-reference:
  reference-id: ORG-SSD
target-reference:
  reference-id: OWASP
remarks: Guidance guidelines ORG.SSD.GL01–GL03 mapped to OWASP for tutorial use.
```

### Step 3: Define Mappings

Define one or more **mappings**. Each mapping links a source entry to an optional target entry and a relationship type. Required and common fields (see [mapping.cue](https://github.com/gemaraproj/gemara/blob/main/mapping.cue)):

| Field                 | Required | Description                                                                 |
|-----------------------|----------|-----------------------------------------------------------------------------|
| `id`                  | Yes      | Unique identifier for this mapping                                         |
| `source`              | Yes      | Entry in the source artifact: `entry-id` and `entry-type`                  |
| `target`              | Yes*     | Entry in the target artifact: `entry-id` and `entry-type`. Omit for `no-match` |
| `relationship`        | Yes      | One of `implements`, `implemented-by`, `supports`, `supported-by`, `equivalent`, `subsumes`, `no-match`, `relates-to` |
| `strength`            | No       | Author's estimate of how completely source satisfies target (1–10); omit for `no-match` |
| `confidence-level`    | No       | `Undetermined`, `Low`, `Medium`, or `High`                                |
| `applicability`       | No       | List of group ids for when this mapping holds (define `applicability-groups` in metadata if used) |
| `rationale`           | No       | Why this relationship exists                                              |
| `remarks`             | No       | General prose regarding this mapping                                      |

**Example (YAML):** Map the three guidelines from the Secure Software Development Guidance (ORG.SSD.GL01, GL02, GL03) to OWASP Top 10 categories (A06, A01, A02):

```yaml
mappings:
# Guidance GL01 -> OWASP Top 10 A06
  - id: GL01-A06
    source:
      entry-id: ORG.SSD.GL01
      entry-type: Guideline
    target:
      entry-id: "A06"
      entry-type: Guideline
    relationship: relates-to
    strength: 7
    rationale: Immutable image references support supply chain integrity; OWASP A06 covers vulnerable and outdated components.

  - id: GL02-A01
    source:
      entry-id: ORG.SSD.GL02
      entry-type: Guideline
    target:
      entry-id: "A01"
      entry-type: Guideline
    relationship: relates-to
    strength: 6
    rationale: Branch protection reduces unauthorized code changes; OWASP A01 covers broken access control.

  - id: GL03-A02
    source:
      entry-id: ORG.SSD.GL03
      entry-type: Guideline
    target:
      entry-id: "A02"
      entry-type: Guideline
    relationship: relates-to
    strength: 6
    rationale: VPN on untrusted networks protects data in transit; OWASP A02 covers cryptographic failures.
```

### Step 4: Validate

The document must conform to the MappingDocument definition in the CUE module. Validate with CUE:

**Validation commands:**

Using the **published** module:

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#MappingDocument' github.com/gemaraproj/gemara@latest your-mapping-document.yaml
```

Fix any errors (e.g. missing `mapping-references`, invalid relationship type, missing `target` when relationship is not `no-match`, or entry-type not in the allowed set) so the document is schema-valid.

### Step 5: Assemble the Full Document and Validate

Combine metadata, source-reference, target-reference, remarks, and mappings into a single YAML document. A complete, schema-valid mapping for the Secure Software Development Guidance → OWASP example is in [mapping-document.yaml](mapping-document.yaml) in this directory.

**Validate** from the repo root:

```bash
cue vet -c -d '#MappingDocument' . docs/tutorials/mapping/mapping-document.yaml
```

## Summary: From Guidance Catalog to Mapping Document

| From guidance catalog        | Use in mapping document                                        |
|------------------------------|----------------------------------------------------------------|
| Catalog id (e.g. ORG-SSD in mapping-references) | Same id as `source-reference.reference-id` and in `metadata.mapping-references` |
| Guideline ids (e.g. ORG.SSD.GL01) | `source.entry-id` in each mapping                              |
| External framework in mapping-references (e.g. OWASP) | Same id as `target-reference.reference-id` and in `metadata.mapping-references` |
| Target framework entry ids (e.g. A06, A01) | `target.entry-id` in each mapping                              |

## What's Next

- Use this mapping in **Layer 2** or **Layer 3** workflows to show how your guidance or controls align to external frameworks.
- Add **applicability-groups** in metadata and use `applicability` on mappings when the relationship holds only in certain contexts (e.g., manufacturer vs open-source steward). You can combine multiple relationship types and `no-match` entries in one document per [mapping.cue](https://github.com/gemaraproj/gemara/blob/main/mapping.cue).

See the [Mapping schema documentation](https://gemara.openssf.org/schema/mapping.html) and [mapping.cue](https://github.com/gemaraproj/gemara/blob/main/mapping.cue) for the full specification.
