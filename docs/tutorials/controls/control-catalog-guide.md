---
layout: page
title: Control Catalog Guide
description: Step-by-step guide to creating Gemara-compatible control catalogs
---

## What This Is

This guide walks through creating a **control catalog** using the [Gemara](https://gemara.openssf.org/) project, building on the threats and scope you identified in the [Threat Assessment Guide](threat-assessment-guide).

In technical terms:
* **Controls** are safeguards with a stated objective and a list of assessment requirements.
* **Families** group related controls by domain (e.g., supply chain, access control).
* **Threat mappings** link each control to the threat(s) it mitigates, connecting your control catalog to your threat catalog.

This exercise produces a structured way to develop control objectives and corresponding testable conditions to determine if the objective is met.

## Prerequisites

Complete the [Threat Assessment Guide](threat-assessment-guide) for your scope (e.g., the container management tool **SEC.SLAM.CM**). You will reference the threat catalog’s ID and threat IDs when defining controls and threat-mappings.


## Walkthrough

### Step 0: Define Scope and References

Reuse the same component or technology as your threat assessment. You will reference that threat catalog so controls can map to the threats you identified.

**Leverage existing resources**: As with threat catalogs, you can import controls from external catalogs (e.g., FINOS Common Cloud Controls) via `mapping-references` and `imported-controls`, so you don’t have to define every control from scratch.

We continue with the container management tool example (SEC.SLAM.CM) and assume a threat catalog for it already exists.

### Step 1: Setting Up Metadata

Declare your control catalog and mapping references. Key fields:

| Field                               | What It Is                                                   | Why                                                                                       |
|-------------------------------------|--------------------------------------------------------------|-------------------------------------------------------------------------------------------|
| `title`                             | Display name for the control catalog (top-level field)       | Human-readable label used in reports and tooling output                                   |
| `mapping-references` with threat catalog id | A pointer to your threat catalog (e.g., SEC.SLAM.CM)        | Lets parsers resolve threat IDs used in `threat-mappings`                                 |
| `mapping-references` with `id: CCC` | Optional pointer to CCC or another control/threat catalog    | Required if you use `imported-controls` or reference CCC threats in threat-mappings       |
| `imported-controls` (Step 3)        | External controls by ID (e.g., from CCC)                     | Reuse common controls without redefining them                                             |
| `threat-mappings` (Step 4)          | Links from each control to threat IDs (your catalog + CCC)   | Documents which threats each control mitigates                                            |

**Example (YAML):**

```yaml
title: Container Management Tool Security Control Catalog
metadata:
  id: SEC.SLAM.CM
  description: Control catalog for container management tool security; mitigates threats from the SEC.SLAM.CM threat catalog.
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: SEC.SLAM.CM
      title: Container Management Tool Security Threat Catalog
      version: "1.0.0"
      url: file://threats.yaml
      description: |
        Threat catalog for the same scope; provides threat IDs for threat-mappings.
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls/releases
      description: |
        Foundational repository of reusable security controls, capabilities,
        and threat models maintained by FINOS.
```

### Step 2: Define Control Families

**Families** group controls by theme. The Layer 2 schema requires at least one family when the catalog defines its own `controls`. Each control’s `family` field must match the `id` of one of these groups.

Required fields for each family (see `base.cue`):

| Field         | Required | Description                                  |
|---------------|----------|----------------------------------------------|
| `id`          | Yes      | Unique identifier referenced by controls     |
| `title`       | Yes      | Short name for the family                    |
| `description` | Yes      | Explanation of what controls in this family address |

**Example (YAML):**

```yaml
families:
  - id: SEC.SLAM.CM.FAM01
    title: Image Integrity and Supply Chain
    description: |
      Controls that ensure container images are authentic, unmodified,
      and from trusted sources throughout retrieval and use.
```

### Step 3: Import or Define Controls

**Option A — Import controls:** If an external catalog (e.g., CCC) defines controls that address your threats, reference it in `mapping-references` and list the control IDs in `imported-controls`.

**Example (YAML):**

```yaml
imported-controls:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CTL42
        remarks: Image signing and verification
```

**Option B — Define your own controls:** For controls specific to your scope, define them under `controls`. Required and common fields (see `layer-2.cue`):

| Field                    | Required | Description                                                                 |
|--------------------------|----------|-----------------------------------------------------------------------------|
| `id`                     | Yes      | Unique identifier (e.g., `ORG.PROJ.COMPONENT.CTL##`)                        |
| `title`                  | Yes      | Short name describing the control                                           |
| `objective`              | Yes      | Unified statement of intent for the control                                 |
| `family`                 | Yes      | `id` of a family in this catalog (or in a referenced catalog for imports)  |
| `assessment-requirements`| Yes      | List of verifiable conditions; each has `id`, `text`, `applicability`      |
| `threat-mappings`        | No       | Links to threat catalog(s) and threat IDs this control mitigates            |
| `state`                  | No       | Lifecycle: `active` (default), `draft`, `deprecated`, `retired`             |

Each **assessment requirement** (see `layer-2.cue`) must have:

| Field           | Required | Description                                                                 |
|-----------------|----------|-----------------------------------------------------------------------------|
| `id`            | Yes      | Unique within the control (e.g., `SEC.SLAM.CM.CTL01.AR01`)                 |
| `text`          | Yes      | Verifiable condition (typically a MUST statement)                           |
| `applicability` | Yes      | List of strings describing when this requirement applies                    |

**Example (YAML):**

```yaml
controls:
  - id: SEC.SLAM.CM.CTL01
    title: Use Immutable Image References by Digest
    objective: |
      Ensure the container management tool resolves and uses container images
      by digest (or digest-plus-tag) so that image identity is immutable and
      tampering or tag swapping cannot substitute a different image.
    family: SEC.SLAM.CM.FAM01
    assessment-requirements:
      - id: SEC.SLAM.CM.CTL01.AR01
        text: |
          The system MUST resolve image references to a digest (e.g., sha256:...)
          before pull or run and MUST use that digest for all subsequent use.
        applicability: ["all deployments"]
      - id: SEC.SLAM.CM.CTL01.AR02
        text: |
          Configuration and policies MUST disallow or override use of tag-only
          references for production or sensitive workloads where supported.
        applicability: ["production", "sensitive workloads"]
    threat-mappings:
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.THR01
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.TH14
```

### Step 4: Validate Against the Layer 2 Schema

The final catalog must conform to the Gemara Layer 2 schema in `layer-2.cue`. Validate with CUE:

**Validation commands:**

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@latest your-control-catalog.yaml
```

Or validate from a local clone:

```bash

```

Fix any reported errors (e.g., missing required fields, invalid `family` reference, or malformed `threat-mappings`) so the catalog is schema-consistent.

### Step 5: Assemble the Full Catalog and Validate

Combine metadata, mapping-references, families, imported-controls (if any), and controls into a single YAML document. A complete, schema-valid catalog for the SEC.SLAM.CM scenario is in [control-catalog.yaml](control-catalog.yaml) in this directory.

**Complete example (SEC.SLAM.CM):**

```yaml
title: Container Management Tool Security Control Catalog

metadata:
  id: SEC.SLAM.CM
  description: |
    Control catalog for container management tool security; mitigates threats
    from the SEC.SLAM.CM threat catalog.
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: SEC.SLAM.CM
      title: Container Management Tool Security Threat Catalog
      version: "1.0.0"
      url: https://example.org/catalogs/SEC.SLAM.CM-threats.yaml
      description: |
        Threat catalog for the same scope; provides threat IDs for threat-mappings.
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls/releases
      description: |
        Foundational repository of reusable security controls, capabilities,
        and threat models maintained by FINOS.

families:
  - id: SEC.SLAM.CM.FAM01
    title: Image Integrity and Supply Chain
    description: |
      Controls that ensure container images are authentic, unmodified,
      and from trusted sources throughout retrieval and use.

controls:
  - id: SEC.SLAM.CM.CTL01
    title: Use Immutable Image References by Digest
    objective: |
      Ensure the container management tool resolves and uses container images
      by digest (or digest-plus-tag) so that image identity is immutable and
      tampering or tag swapping cannot substitute a different image.
    family: SEC.SLAM.CM.FAM01
    assessment-requirements:
      - id: SEC.SLAM.CM.CTL01.AR01
        text: |
          The system MUST resolve image references to a digest (e.g., sha256:...)
          before pull or run and MUST use that digest for all subsequent use.
        applicability: ["all deployments"]
      - id: SEC.SLAM.CM.CTL01.AR02
        text: |
          Configuration and policies MUST disallow or override use of tag-only
          references for production or sensitive workloads where supported.
        applicability: ["production", "sensitive workloads"]
    threat-mappings:
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.THR01
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.TH14
```

**Validate** from the repo root:

```bash
cue vet -c -d '#ControlCatalog' ./layer-2.cue ./metadata.cue ./mapping.cue ./base.cue docs/tutorials/controls/control-catalog.yaml
```

Fix any reported errors (e.g., missing required fields, invalid `family` reference, or malformed `threat-mappings`) so the catalog is schema-consistent.

## Summary: From Threat Assessment to Control Catalog

| From threat assessment        | Use in control catalog                                      |
|-------------------------------|-------------------------------------------------------------|
| Scope / catalog id (e.g. SEC.SLAM.CM) | Same id in control catalog metadata; reference in mapping-references |
| Threat IDs (e.g. SEC.SLAM.CM.THR01, CCC.Core.TH14) | `threat-mappings` on each control that mitigates those threats |
| Understanding of what goes wrong | Control objectives and assessment requirements that prevent or detect it |

## What's Next

Use the control catalog in downstream Gemara layers: for example, map controls to Layer 1 guidelines via `guideline-mappings`, or use the catalog in Layer 5 evaluations to record which assessment requirements have been verified.

See the [Gemara Layer 2 schema documentation](https://gemara.openssf.org/schema/layer-2.html) for the full specification and optional fields (e.g., `state`, `replaced-by`, `guideline-mappings`).
