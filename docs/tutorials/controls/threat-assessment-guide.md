---
layout: page
title: Threat Assessment Guide
description: Step-by-step guide to performing Gemara-compatible threat assessments
---

## What This Is

This guide walks through a threat assessment using the [Gemara](https://gemara.openssf.org/) project.

**The basic idea:** Think of a project like a house. First, you identify what the house can do: its **capabilities** (e.g., "allow entry/exit", "store belongings"). Then, you identify **threats**, what could go wrong with those capabilities (e.g., "unauthorized entry through unlocked door", "theft of stored belongings").

In technical terms:
* **Capabilities** define what the technology can do. These form a primary component of the **attack surface** because every intended function represents a potential path for unintended use.
* **Threats** define specific ways those capabilities could be misused or exploited.

This exercise helps you systematically identify what could go wrong so you can build appropriate defenses.

## Walkthrough

### Step 0: Define Scope

Select a component or technology to assess (service, API, infrastructure component, or technology stack).

**Leverage existing resources**: Gemara supports importing threats and capabilities from external catalogs so you don't have to start from scratch. The FINOS Common Cloud Controls (CCC) Core [catalog](https://github.com/finos/common-cloud-controls/releases/download/v2025.10/CCC.Core_v2025.10.yaml) defines well-vetted capabilities and threats that apply broadly across cloud services. These pre-built items can help accelerate your assessment.

We will explore how this is leveraged below as we dive into our container management tool example (i.e., SEC.SLAM.CM).

### Step 1: Setting Up Metadata

Declare your scope and mapping references. Key fields:

| Field                               | What It Is                                                   | Why                                                                                       |
|-------------------------------------|--------------------------------------------------------------|-------------------------------------------------------------------------------------------|
| `title`                             | Display name for the threat catalog (top-level field)        | Human-readable label used in reports and tooling output                                   |
| `mapping-references` with `id: CCC` | A pointer to the CCC Core catalog release                    | Tells parsers where to resolve the imported capability and threat IDs used in later steps |
| `imported-capabilities` (Step 2)    | Specific CCC Core capabilities by ID (e.g., `CCC.Core.CP29`) | Brings in common capabilities without redefining them                                     |
| `imported-threats` (Step 3)         | Specific CCC Core threats by ID (e.g., `CCC.Core.TH14`)      | Brings in common threats without redefining them                                          |

**Example (YAML):**

```yaml
title: Container Management Tool Security Threat Catalog
metadata:
  id: SEC.SLAM.CM
  description: Threat catalog for container management tool security assessment
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls/releases
      description: |
        Foundational repository of reusable security controls, capabilities,
        and threat models maintained by FINOS.
```

### Step 2: Identify Capabilities

Capabilities are the core functions or features within the scope.

**Start with the imported capabilities** you can leverage from FINOS CCC. Ask: "Which common cloud capabilities does this technology have?"

A container management tool actively reaches out to registries to pull images and configuration.
Since CCC Core already defines this as **CP29** (Active Ingestion), we import it rather than redefining it.
Image tags also function as version identifiers — the tool resolves a tag like `latest` or `v1.0` to a specific image.
CCC Core defines this as **CP18** (Resource Versioning), so we import that as well.

**Example (YAML)**

```yaml
imported-capabilities:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CP29
        remarks: Active Ingestion
      - reference-id: CCC.Core.CP18
        remarks: Resource Versioning
```

**Then, define specific capabilities** unique to your target. Required fields:

| Field         | Required | Description                                                        |
|---------------|----------|--------------------------------------------------------------------|
| Capability ID | Yes      | Unique identifier following the pattern `ORG.PROJ.COMPONENT.CAP##` |
| Title         | Yes      | A clear, concise name that describes the capability                |
| Description   | Yes      | A specific explanation of what this capability does                |

**Example (YAML)**

```yaml
capabilities:
  - id: SEC.SLAM.CM.CAP01
    title: Image Retrieval by Tag
    description: |
      Ability to retrieve container images from registries using mutable tag names
      (e.g., 'latest', 'v1.0').
```

### Step 3: Identify Threats

Threats are specific ways capabilities can be misused, exploited, or cause problems. For each capability, identify potential threats.

**Check for imported threats first.** As with capabilities, review the CCC Core catalog for threats linked to the capabilities you imported.
If a threat fits your scope, import it. In this example, CCC Core defines **TH14** ("Older Resource Versions are Used") which is linked to **CP18**.
It applies because mutable image tags let the tool resolve to a stale or compromised version.

**Example (YAML)**

```yaml
imported-threats:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.TH14
```

**Then, define specific threats** unique to your target. Required fields:

| Field             | Required | Description                                                                        |
|-------------------|----------|------------------------------------------------------------------------------------|
| Threat ID         | Yes      | Unique identifier following the pattern `ORG.PROJ.COMPONENT.THR##`                 |
| Title             | Yes      | A clear, concise name describing the threat                                        |
| Description       | Yes      | A specific explanation of what goes wrong and why it matters                       |
| Capabilities      | Yes      | Links this threat to the capability(ies) it exploits                               |

**Example (YAML)**

```yaml
threats:
  - id: SEC.SLAM.CM.THR01
    title: Container Image Tampering or Poisoning
    description: |
      Attackers may replace a legitimately published image tag with a malicious image
      by exploiting tag mutability in image registries, especially when the container
      management tool retrieves images by tag name rather than digest. This enables
      unauthorized access, data exfiltration, and system compromise.
    capabilities:
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.CP29
          - reference-id: CCC.Core.CP18
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.CAP01
```

### Step 4: Validate

The final YAML should look something like this:
```yaml
metadata:
  id: SEC.SLAM.CM
  description: Threat catalog for container management tool security assessment
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls/releases
      description: |
        Foundational repository of reusable security controls, capabilities,
        and threat models maintained by FINOS.
title: Container Management Tool Security Threat Catalog
imported-capabilities:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CP29
        remarks: Active Ingestion
      - reference-id: CCC.Core.CP18
        remarks: Resource Versioning
imported-threats:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.TH14
capabilities:
  - id: SEC.SLAM.CM.CAP01
    title: Image Retrieval by Tag
    description: |
      Ability to retrieve container images from registries using mutable tag names
      (e.g., 'latest', 'v1.0').
threats:
  - id: SEC.SLAM.CM.THR01
    title: Container Image Tampering or Poisoning
    description: |
      Attackers may replace a legitimately published image tag with a malicious image
      by exploiting tag mutability in image registries, especially when the container
      management tool retrieves images by tag name rather than digest. This enables
      unauthorized access, data exfiltration, and system compromise.
    capabilities:
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.CP29
          - reference-id: CCC.Core.CP18
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.CAP01
```

**Validation commands:**

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#ThreatCatalog' github.com/gemaraproj/gemara@latest your-threats.yaml
```

## What's Next

Create a Gemara Control Catalog that maps security controls to the identified threats using the [Control Catalog Guide](control-catalog-guide). See the [Gemara Layer 2 schema documentation](https://gemara.openssf.org/schema/layer-2.html) for the full specification.