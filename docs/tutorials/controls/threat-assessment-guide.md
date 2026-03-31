---
layout: page
title: Threat Assessment Guide
description: Step-by-step guide to performing Gemara-compatible threat assessments
---

## What This Is

This guide walks through a threat assessment using the [Gemara](https://gemara.openssf.org/) project.

**The basic idea:** Think of a project like a house. First, you identify what the house can do: its **[capabilities](../../model/02-definitions.html#capability)** (e.g., "allow entry/exit", "store belongings"). Then, you identify **[threats](../../model/02-definitions.html#threat)**, what could go wrong with those capabilities (e.g., "unauthorized entry through unlocked door", "theft of stored belongings").

In technical terms:
 * **[Capabilities](../../model/02-definitions.html#capability)** define what the technology can do. These form a primary component of the **attack surface** because every intended function represents a potential path for unintended use.
 * **[Threats](../../model/02-definitions.html#threat)** define specific ways those capabilities could be misused or exploited.

This exercise helps you systematically identify what could go wrong so you can build appropriate defenses.

Gemara splits these into two artifact kinds: `CapabilityCatalog` for capability definitions, and `ThreatCatalog` for threats. A threat catalog references capabilities via `mapping-references` on each threat. 

## Walkthrough

### Step 0: Define Scope

Select a component or technology to assess (service, API, infrastructure component, or technology stack).

**Leverage existing resources**: Gemara supports importing entries from external catalogs so you don't have to start from scratch. The FINOS Common Cloud Controls (CCC) Core [catalog](https://github.com/finos/common-cloud-controls/releases/download/v2025.10/CCC.Core_v2025.10.yaml) defines well-vetted capabilities and threats that apply broadly across cloud services. These pre-built items can help accelerate your assessment.

We will explore how this is leveraged below as we dive into our container management tool example (i.e., **SEC.SLAM.CM**).

### Step 1: Setting Up Metadata (Threat catalog)

Declare your scope and mapping references for the `ThreatCatalog`. Key fields:

| Field | What It Is | Why |
|-------|------------|-----|
| `title` | Display name for the threat catalog (top-level field) | Human-readable label used in reports and tooling output |
| `metadata.type` | Must be `ThreatCatalog` | Identifies the artifact for `#ThreatCatalog` validation |
| `metadata.gemara-version` | String (e.g. `1.0.0-rc.1`) | Declares which Gemara specification version the file conforms to (required) |
| `mapping-references` with `id: CCC` | Pointer to the CCC Core catalog release | Resolve imported CCC capability and threat IDs used in `imports` and in each threat's `capabilities` |
| `mapping-references` for scope capabilities | Pointer to your `CapabilityCatalog` (see Step 2) | Resolve IDs such as `SEC.SLAM.CM.CAP01` referenced from each threat's `capabilities` |
| Top-level `imports` (optional) | List of `#MultiEntryMapping` rows | Pull CCC (or other) capability/threat entries into this catalog without redefining them |

**Example (YAML)** — threat catalog metadata only:

```yaml
title: Container Management Tool Security Threat Catalog
metadata:
  id: SEC.SLAM.CM
  type: ThreatCatalog
  gemara-version: "1.0.0-rc.1"
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
    - id: SEC.SLAM.CM.CAP
      title: Container Management Tool Security Capability Catalog
      version: "1.0.0"
      url: https://example.org/catalogs/SEC.SLAM.CM-capabilities.yaml
      description: |
        Scope-specific capabilities for this assessment (see Step 2).
```

### Step 2: Identify Capabilities (Capability catalog)

Capabilities are the core functions or features within the scope. In Gemara they live in a `CapabilityCatalog`. For SEC.SLAM.CM, scope-specific capabilities are in [capabilities.yaml](capabilities.yaml) (`metadata.id` `SEC.SLAM.CM.CAP`) so threats can reference them with `reference-id: SEC.SLAM.CM.CAP` and `entries` listing IDs such as `SEC.SLAM.CM.CAP01`.

**Start with the imported capabilities** you can leverage from FINOS CCC. Ask: "Which common cloud capabilities does this technology have?" 

A container management tool actively reaches out to registries to pull images and configuration. Since CCC Core already defines that as **CP29** (Active Ingestion), we import it rather than redefining it. Image tags also function as version identifiers - the tool resolves a tag like `latest` or `v1.0` to a specific image. CCC Core defines that behavior as **CP18** (Resource Versioning), so we import that as well. 

> **Note:** Those IDs are referenced from threats under each threat’s `capabilities` and may also appear under top-level `imports` in the `ThreatCatalog`.

**Then, define specific capabilities** unique to your target. Required fields:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier (e.g. `ORG.PROJ.COMPONENT.CAP##`) |
| `title` | Yes | Clear name for the capability |
| `description` | Yes | What this capability does |
| `group` | Yes | `id` of a **group** defined in the same capability catalog |

**Example (YAML)** — see [capabilities.yaml](capabilities.yaml):

```yaml
title: Container Management Tool Security Capability Catalog
metadata:
  id: SEC.SLAM.CM.CAP
  type: CapabilityCatalog
  gemara-version: "1.0.0-rc.1"
  description: |
    Capabilities unique to the container management tool scope.
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human

groups:
  - id: SEC.SLAM.CM.CAPGRP01
    title: Image retrieval and resolution
    description: |
      How the tool retrieves images and resolves references to artifacts.

capabilities:
  - id: SEC.SLAM.CM.CAP01
    title: Image Retrieval by Tag
    description: |
      Ability to retrieve container images from registries using mutable tag names
      (e.g., 'latest', 'v1.0').
    group: SEC.SLAM.CM.CAPGRP01
  - id: SEC.SLAM.CM.CAP02
    title: Image Reference Lookup
    description: |
      The tool resolves image references via network requests; resolution time may
      differ from use time, and references may be mutable.
    group: SEC.SLAM.CM.CAPGRP01
```

### Step 3: Identify Threats (Threat catalog)

 **[Threats](../../model/02-definitions.html#threat)** are specific ways **[capabilities](../../model/02-definitions.html#capability)** can be misused, exploited, or cause problems. For each **capability**, identify potential **threats**.

Check for imported **[threats](../../model/02-definitions.html#threat)** first. As with **[capabilities](../../model/02-definitions.html#capability)**, review the CCC Core catalog for threats linked to the capabilities you imported. If a threat fits your scope, import it. In this example, CCC Core defines **TH14** ("Older Resource Versions are Used") which is linked to **CP18**. It applies because mutable image tags let the tool resolve to a stale or compromised version.

**Importing from CCC.** List CCC rows under top-level `imports` as a list of mappings. You can include both capability and threat IDs from CCC in the same `entries` list when they come from that single mapping reference.

**Example (YAML)** — `imports` on the threat catalog:

```yaml
imports:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CP29
        remarks: Active Ingestion
      - reference-id: CCC.Core.CP18
        remarks: Resource Versioning
      - reference-id: CCC.Core.TH14
        remarks: Older Resource Versions are Used
```

**Then, define specific threats** unique to your target. Required fields:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier following the pattern `ORG.PROJ.COMPONENT.THR##` |
| `title` | Yes | Short name for the threat |
| `description` | Yes | What goes wrong and why it matters |
| `group` | Yes | `id` of a **group** defined in this threat catalog |
| `capabilities` | Yes | Links this threat to the capabilities it exploits |
| `vectors` | No | Optional link to vector catalog entries |
| `actors` | No | Optional threat actors (`#Actor`) |

**Example (YAML)** — a custom threat (*Container Image Tampering or Poisoning*) linked to the capabilities it exploits: **CCC.Core.CP29** (Active Ingestion), **CCC.Core.CP18** (Resource Versioning), and **SEC.SLAM.CM.CAP01** (Image Retrieval by Tag) via your scope capability catalog reference (`SEC.SLAM.CM.CAP`).

```yaml
groups:
  - id: SEC.SLAM.CM.FAM01
    title: Image integrity and supply chain
    description: |
      Threats affecting container image retrieval, integrity, and trust.

threats:
  - id: SEC.SLAM.CM.THR01
    title: Container Image Tampering or Poisoning
    description: |
      Attackers may replace a legitimately published image tag with a malicious image
      by exploiting tag mutability in image registries, especially when the container
      management tool retrieves images by tag name rather than digest. This enables
      unauthorized access, data exfiltration, and system compromise.
    group: SEC.SLAM.CM.FAM01
    capabilities:
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.CP29
          - reference-id: CCC.Core.CP18
      - reference-id: SEC.SLAM.CM.CAP
        entries:
          - reference-id: SEC.SLAM.CM.CAP01
```

### Step 4: Validate

The final YAML should look something like this:

```yaml
title: Container Management Tool Security Threat Catalog

metadata:
  id: SEC.SLAM.CM
  type: ThreatCatalog
  gemara-version: "1.0.0-rc.1"
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
    - id: SEC.SLAM.CM.CAP
      title: Container Management Tool Security Capability Catalog
      version: "1.0.0"
      url: https://example.org/catalogs/SEC.SLAM.CM-capabilities.yaml
      description: |
        Scope-specific capabilities (CAP01, CAP02) for this threat assessment.

groups:
  - id: SEC.SLAM.CM.FAM01
    title: Image integrity and supply chain
    description: |
      Threats affecting container image retrieval, integrity, and trust.

imports:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CP29
        remarks: Active Ingestion
      - reference-id: CCC.Core.CP18
        remarks: Resource Versioning
      - reference-id: CCC.Core.TH14
        remarks: Older Resource Versions are Used

threats:
  - id: SEC.SLAM.CM.THR01
    title: Container Image Tampering or Poisoning
    description: |
      Attackers may replace a legitimately published image tag with a malicious image
      by exploiting tag mutability in image registries, especially when the container
      management tool retrieves images by tag name rather than digest. This enables
      unauthorized access, data exfiltration, and system compromise.
    group: SEC.SLAM.CM.FAM01
    capabilities:
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.CP29
          - reference-id: CCC.Core.CP18
      - reference-id: SEC.SLAM.CM.CAP
        entries:
          - reference-id: SEC.SLAM.CM.CAP01
```

The complete tutorial adds more CCC imports and additional threats; see [threat-catalog.yaml](threat-catalog.yaml). 

**Validation commands:**

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#CapabilityCatalog' github.com/gemaraproj/gemara@v1 docs/tutorials/controls/your-capabilities.yaml
cue vet -c -d '#ThreatCatalog' github.com/gemaraproj/gemara@v1 docs/tutorials/controls/your-threat-catalog.yaml
```

## What's Next

Create a Gemara control catalog that maps security controls to these threats using the [Control Catalog Guide](control-catalog-guide).

**Layer 2 schema documentation:** 
- [Capability Catalog](https://gemara.openssf.org/schema/capabilitycatalog.html)
- [Threat Catalog](https://gemara.openssf.org/schema/threatcatalog.html)
- [Control Catalog](https://gemara.openssf.org/schema/controlcatalog.html)
