---
layout: page
title: Mapping Document Guide
description: Step-by-step guide to creating Gemara-compatible mapping documents
---

## What This Is

This guide walks through creating a **Mapping Document** using the [Gemara](https://gemara.openssf.org/) project, building on the guidance catalog you created in the [Guidance Catalog Guide](../guidance/guidance-guide). Examples use `gemara-version: "1.0.0-rc.0"` to match the [v1.0.0-rc.0](https://github.com/gemaraproj/gemara/releases/tag/v1.0.0-rc.0) specification release candidate; adjust if you target a different Gemara version.

A mapping document captures the user's intent for how entries in a **source** artifact relate to entries in a **target** artifact. It is the place to express alignment between independently authored catalogs (e.g., your guidance to OWASP, or controls to regulations) in a single, directed way.

### Using a Mapping Document vs Inline Mappings

* **Mapping Documents** are for complex descriptions of relationships between two or more artifacts.
* **Inline Mappings** are relationships that capture the author's intent at the time of creating that entry.

> **Note:** See the FAQ for additional context on the Mapping Documents.


In technical terms:
* **Source reference** is the artifact you map *from* (e.g., your guidance catalog). Its `reference-id` must match an id in `metadata.mapping-references`. **`entry-type`** on `source-reference` applies to every **source** entry id in `mappings` (see `#TypedMapping` in [mappingdocument.cue](https://github.com/gemaraproj/gemara/blob/main/mappingdocument.cue)).
* **Target reference** is the artifact you map *to* (e.g., OWASP Top 10). Its `reference-id` must also match an id in `metadata.mapping-references`. **`entry-type`** on `target-reference` applies to every **target** row in `targets`.
* **Mappings** are one or more atomic relationships: each links a **source** entry id (string) to one or more **targets** (`#MappingTarget` objects with `entry-id` and optional per-target fields). For `no-match`, the source has no counterpart in the target and **`targets` must be omitted**.
* **Relationship types** (see `#RelationshipType` in [mappingdocument.cue](https://github.com/gemaraproj/gemara/blob/main/mappingdocument.cue)): `implements`, `implemented-by`, `supports`, `supported-by`, `equivalent`, `subsumes`, `no-match`, `relates-to`.
* **Entry type** for both sides is declared once on **`source-reference`** and **`target-reference`** via `#TypedMapping`. Allowed values are `#EntryType` in [mappingdocument.cue](https://github.com/gemaraproj/gemara/blob/main/mappingdocument.cue): `Guideline`, `Statement`, `Control`, `AssessmentRequirement`, `Capability`, `Threat`, `Risk`, or `Vector`.
* **Mapping references** in metadata use `#MappingReference` from [mapping_inline.cue](https://github.com/gemaraproj/gemara/blob/main/mapping_inline.cue) (shared with other layers); see also [metadata.cue](https://github.com/gemaraproj/gemara/blob/main/metadata.cue) for `#Metadata`.

This exercise produces a mapping document that downstream tools and policies can use to understand how two artifacts align.

## Optional Workflow

Complete the [Guidance Catalog Guide](../guidance/guidance-guide) and have a guidance catalog (e.g., [guidance-example.yaml](../guidance/guidance-example.yaml)) with `mapping-references` that include the target framework you want to map to (e.g., OWASP). You will reference that guidance catalog as the source and the external framework as the target.

## Walkthrough

We use the **Secure Software Development Guidance** (`ORG.SSD.001`) from the guidance tutorial and map its guidelines to **OWASP Top 10**. The same pattern applies when mapping controls to regulations or other frameworks.

### Step 0: Define Scope and References

Decide which two artifacts you are mapping: **source** (the one you map *from*) and **target** (the one you map *to*). Reuse the same catalog ids as in your guidance catalog’s `mapping-references` so the mapping document can refer to them.

**Leverage existing resources:** Your guidance catalog already declares `mapping-references` (e.g., OWASP). The mapping document must declare **both** the source and target in its own `metadata.mapping-references`; the source is your guidance catalog (same id as `metadata.id` in that catalog, e.g. `ORG.SSD.001`) and the target is the external framework (e.g., OWASP).

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
  gemara-version: "1.0.0-rc.0"
  description: >
    Maps Secure Software Development Guidance guidelines to OWASP Top 10
    categories. Minimal example for tutorials; relationship types are relates-to.
  author:
    id: gemara-example
    name: Gemara Example Author
    type: Human
  mapping-references:
    - id: ORG.SSD.001
      title: Secure Software Development Guidance
      version: "1.0.0"
      url: "file://../guidance/guidance-example.yaml"
    - id: OWASP
      title: OWASP Top 10
      version: "2021"
      url: "https://owasp.org/Top10"
```

### Step 2: Source and Target References (`#TypedMapping`)

Set **`source-reference`** and **`target-reference`**. Each is a **typed mapping**: `reference-id` (must match `metadata.mapping-references`) plus **`entry-type`** for all entries on that side of the document. Optionally add top-level **remarks** to describe the mapping.

| Field               | What It Is                                                                 |
|---------------------|----------------------------------------------------------------------------|
| `source-reference`  | Artifact you map *from*; `reference-id` and required `entry-type`          |
| `target-reference`  | Artifact you map *to*; `reference-id` and required `entry-type`            |
| `remarks`           | Optional prose about this mapping document as a whole                      |

**Example (YAML):**

```yaml
source-reference:
  reference-id: ORG.SSD.001
  entry-type: Guideline
target-reference:
  reference-id: OWASP
  entry-type: Guideline
remarks: Guidance guidelines ORG.SSD.GL01–GL03 mapped to OWASP for tutorial use.
```

### Step 3: Define Mappings

Define one or more **mappings**. Each mapping has a **source** string (the source entry’s id) and, unless `relationship` is `no-match`, a non-empty **`targets`** list. Each list element is a **`#MappingTarget`**: at minimum `entry-id`, plus optional `strength`, `confidence-level`, `applicability`, `rationale`, and `remarks` for that target. You may also set `strength`, `confidence-level`, `applicability`, `rationale`, and `remarks` on the mapping itself when they apply to the whole row (see [mappingdocument.cue](https://github.com/gemaraproj/gemara/blob/main/mappingdocument.cue)).

| Field                 | Required | Description                                                                 |
|-----------------------|----------|-----------------------------------------------------------------------------|
| `id`                  | Yes      | Unique identifier for this mapping                                         |
| `source`              | Yes      | Source entry id (string), matching an entry in the source artifact        |
| `targets`             | Yes*     | Non-empty list of `{ entry-id: ... }` (and optional per-target fields). Omit only when `relationship` is `no-match` |
| `relationship`        | Yes      | One of `implements`, `implemented-by`, `supports`, `supported-by`, `equivalent`, `subsumes`, `no-match`, `relates-to` |
| `strength`            | No       | Mapping-level estimate (1–10); per-target `strength` on a `#MappingTarget` overrides where used |
| `confidence-level`    | No       | `Undetermined`, `Low`, `Medium`, or `High`                                |
| `applicability`       | No       | List of group ids (define `applicability-groups` in metadata if used)     |
| `rationale`           | No       | Why this relationship exists                                              |
| `remarks`             | No       | General prose regarding this mapping                                      |

**Example (YAML):** Map the three guidelines from the Secure Software Development Guidance (`ORG.SSD.GL01`, `GL02`, `GL03`) to OWASP Top 10 categories (`A06`, `A01`, `A02`):

```yaml
mappings:
  - id: GL01-A06
    source: ORG.SSD.GL01
    relationship: relates-to
    strength: 7
    rationale: Immutable image references support supply chain integrity; OWASP A06 covers vulnerable and outdated components.
    targets:
      - entry-id: "A06"

  - id: GL02-A01
    source: ORG.SSD.GL02
    relationship: relates-to
    strength: 6
    rationale: Branch protection reduces unauthorized code changes; OWASP A01 covers broken access control.
    targets:
      - entry-id: "A01"

  - id: GL03-A02
    source: ORG.SSD.GL03
    relationship: relates-to
    strength: 6
    rationale: VPN on untrusted networks protects data in transit; OWASP A02 covers cryptographic failures.
    targets:
      - entry-id: "A02"
```

For **`no-match`**, omit **`targets`** entirely (the source entry has no counterpart in the target artifact).

### Step 4: Validate against the Mapping Document Schema

Validate with CUE:

**Validation commands:**

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#MappingDocument' github.com/gemaraproj/gemara@latest your-mapping-document-example.yaml
```

From a **clone of this repository** (run from the repo root; replace the placeholder with your file path):

```bash
cue vet -c -d '#MappingDocument' . your-mapping-document-example.yaml
```

A reference copy is [mapping-document.yaml](mapping-document.yaml) in this directory.

### Step 5: Assemble the Full Document and Validate

Combine metadata, source-reference, target-reference, remarks, and mappings into a single YAML document. The following is the full tutorial example (kept in sync with [mapping-document.yaml](mapping-document.yaml) in this directory):

```yaml
# Secure Software Development Guidance to OWASP Top 10 (tutorial example)
# Conforms to Gemara #MappingDocument (mappingdocument.cue).
# gemara-version: v1.0.0-rc.0 — https://github.com/gemaraproj/gemara/releases/tag/v1.0.0-rc.0
# Source guidance catalog: ../guidance/guidance-example.yaml (metadata.id ORG.SSD.001)
# entry-type on source-reference / target-reference applies to all entries on that side (#TypedMapping).
title: Secure Software Development Guidance to OWASP Top 10
metadata:
  id: SSD-OWASP-MAP-001
  version: "1.0.0"
  type: MappingDocument
  gemara-version: "1.0.0-rc.0"
  description: >
    Maps Secure Software Development Guidance guidelines to OWASP Top 10
    categories. Minimal example for tutorials; relationship types are relates-to.
  author:
    id: gemara-example
    name: Gemara Example Author
    type: Human
  mapping-references:
    - id: ORG.SSD.001
      title: Secure Software Development Guidance
      version: "1.0.0"
      url: "file://../guidance/guidance-example.yaml"
    - id: OWASP
      title: OWASP Top 10
      version: "2021"
      url: "https://owasp.org/Top10"

source-reference:
  reference-id: ORG.SSD.001
  entry-type: Guideline
target-reference:
  reference-id: OWASP
  entry-type: Guideline
remarks: Guidance guidelines ORG.SSD.GL01–GL03 mapped to OWASP for tutorial use.

mappings:
  - id: GL01-A06
    source: ORG.SSD.GL01
    relationship: relates-to
    strength: 7
    rationale: Immutable image references support supply chain integrity; OWASP A06 covers vulnerable and outdated components.
    targets:
      - entry-id: "A06"

  - id: GL02-A01
    source: ORG.SSD.GL02
    relationship: relates-to
    strength: 6
    rationale: Branch protection reduces unauthorized code changes; OWASP A01 covers broken access control.
    targets:
      - entry-id: "A01"

  - id: GL03-A02
    source: ORG.SSD.GL03
    relationship: relates-to
    strength: 6
    rationale: VPN on untrusted networks protects data in transit; OWASP A02 covers cryptographic failures.
    targets:
      - entry-id: "A02"
```

**Validate** from the repo root against the checked-in example:

```bash
cue vet -c -d '#MappingDocument' . docs/tutorials/mapping/mapping-document.yaml
```

Fix any errors (e.g. missing `mapping-references`, invalid relationship type, missing **`targets`** when relationship is not `no-match`, missing **`entry-type`** on `source-reference` / `target-reference`, or `entry-type` not in the allowed set) so the document is schema-valid.

## Summary: From Guidance Catalog to Mapping Document

| From guidance catalog        | Use in mapping document                                        |
|------------------------------|----------------------------------------------------------------|
| Catalog id (e.g. `ORG.SSD.001` in mapping-references) | Same id as `source-reference.reference-id` and in `metadata.mapping-references` |
| Guideline ids (e.g. ORG.SSD.GL01) | `source` string in each mapping                              |
| External framework in mapping-references (e.g. OWASP) | Same id as `target-reference.reference-id` and in `metadata.mapping-references` |
| Target framework entry ids (e.g. A06, A01) | `entry-id` under `targets` in each mapping                    |
| Kind of entries (e.g. guidelines on both sides) | `entry-type` on `source-reference` and `target-reference` (`#TypedMapping`) |

## What's Next

- Use this mapping in **Layer 2** or **Layer 3** workflows to show how your guidance or controls align to external frameworks.
- Use **multiple targets** in one mapping when one source row aligns to several target entries (each target is a `#MappingTarget` with its own optional metadata).
- Add **applicability-groups** in metadata and use `applicability` on mappings or on individual `#MappingTarget` rows when the relationship holds only in certain contexts (e.g., manufacturer vs open-source steward).

See the [Mapping Document schema](https://gemara.openssf.org/schema/mappingdocument.html) and the CUE module: [mappingdocument.cue](https://github.com/gemaraproj/gemara/blob/main/mappingdocument.cue) (`#MappingDocument`, `#TypedMapping`, `#Mapping`, `#MappingTarget`, relationships, entry types) and [mapping_inline.cue](https://github.com/gemaraproj/gemara/blob/main/mapping_inline.cue) (`#MappingReference`, `#ArtifactMapping`, and related shared types).
