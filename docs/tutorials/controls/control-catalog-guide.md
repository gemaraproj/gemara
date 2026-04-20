---
layout: page
title: Control Catalog Guide
description: Step-by-step guide to creating Gemara-compatible control catalogs
---

## What This Is

This guide walks through creating a **control catalog** using the [Gemara](https://gemara.openssf.org/) project, building on the threats and scope you identified in the [Threat Assessment Guide](threat-assessment-guide).

In technical terms:
* **[Controls](../../model/02-definitions.html#control)** are safeguards with a stated **objective** and a list of **[assessment requirements](../../model/02-definitions.html#assessment-requirement)**.
* **Groups** group related controls by domain (e.g., supply chain, access control).
* **[Threats](../../model/02-definitions.html#threat)** link each control to the threat(s) it mitigates, connecting your control catalog to your layer 2 threat catalog.

This exercise produces a structured way to develop control objectives and corresponding testable conditions to determine if the objective is met.

## Optional Workflow

Complete the [Threat Assessment Guide](threat-assessment-guide) for your scope (e.g., the container management tool **SEC.SLAM.CM**). You can reference the linked `threats` and `guidelines` to support the intent of Controls in the catalog. 


## Walkthrough

### Step 0: Define Scope and References

Reuse the same component or technology as your threat assessment. You will reference that threat catalog so controls can map to the threats you identified.

**Leverage existing resources**: As with threat catalogs, you can import controls from external catalogs (e.g., FINOS Common Cloud Controls, OSPS Baseline) via `mapping-references` and list imported controls in your catalog. 

We continue with the container management tool example (SEC.SLAM.CM) and assume a threat catalog for it already exists.

### Step 1: Setting Up Metadata

Declare your control catalog and mapping references. Key fields:

| Field                               | What It Is                                                   | Why                                                                                       |
|-------------------------------------|--------------------------------------------------------------|-------------------------------------------------------------------------------------------|
| `title`                             | Display name for the control catalog (top-level field)       | Human-readable label used in reports and tooling output                                   |
| `metadata.type`                     | Must be `ControlCatalog`                                     | Identifies the artifact kind for validation             |
| `metadata.gemara-version`           | String (e.g. `1.0.0`)                                   | Declares which Gemara specification version the file conforms to (required in current schema) |
| `mapping-references` with `id: CCC` | Optional pointer to CCC or another control/threat catalog   | Resolve imported capability and control IDs from the referenced catalog                  |
| `applicability-groups`          | List of groups (id, title, description) for when controls apply | Scope assessment requirements by context (e.g., production, CI/CD) so evaluators know when each requirement applies |

> **Note:** Applicability groups **must** be defined to assign applicability to controls.

**Example (YAML):**

```yaml
title: Container Management Tool Security Control Catalog
metadata:
  id: SEC.SLAM.CM
  type: ControlCatalog
  gemara-version: "1.0.0"
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
        Threat catalog for the same scope; provides threat IDs referenced from each
        control's threats.
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls/releases
      description: |
        Foundational repository of reusable security controls, capabilities,
        and threat models maintained by FINOS.
  applicability-groups:
    - id: production
      title: Production
      description: |
        Production container workloads and clusters; controls apply to
        live environments where security posture is enforced.
    - id: all_deployments
      title: All Deployments
      description: |
        Requirements that apply whenever container images are built, pulled, or run—
        regardless of environment (dev, staging, production) or pipeline stage.
    - id: untrusted_networks
      title: Untrusted Networks
      description: |
        Applies when registry or image traffic traverses untrusted networks.
    - id: ci_cd
      title: CI/CD
      description: |
        Applies in continuous integration and deployment pipelines.
```

### Step 2: Define Control Groups

**Groups** group controls by theme. The Layer 2 schema requires at least one group when the catalog defines its own `controls`. Each control's `group` field must match the `id` of one of these groups.

Required fields for each group:

| Field         | Required | Description                                  |
|---------------|----------|----------------------------------------------|
| `id`          | Yes      | Unique identifier referenced by controls     |
| `title`       | Yes      | Short name for the group                    |
| `description` | Yes      | Explanation of what controls in this group address |

**Example (YAML):**

```yaml
groups:
  - id: SEC.SLAM.CM.FAM01
    title: Image Integrity and Supply Chain
    description: |
      Controls that ensure container images are authentic, unmodified,
      and from trusted sources throughout retrieval and use.
```

### Step 3: Import or Define Controls

> **Note:** For how controls are automatically pulled into policy to mitigate high-severity risks, see the FAQ.


**Option A — Import Controls:** If an external catalog (e.g., CCC) defines controls that address your threats, reference it in `mapping-references` and list imported controls under top-level `imports`.

**Example (YAML):**

```yaml
imports:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CTL42
        remarks: Image signing and verification
```

**Option B — Define your own controls:** For controls specific to your scope, define them under `controls`. Required and common fields (see `controlcatalog.cue`):

| Field                    | Required | Description                                                                 |
|--------------------------|----------|-----------------------------------------------------------------------------|
| `id`                     | Yes      | Unique identifier (e.g., `ORG.PROJ.COMPONENT.CTL##`)                        |
| `title`                  | Yes      | Short name describing the control                                           |
| `objective`              | Yes      | Unified statement of intent for the control                                 |
| `group`                  | Yes      | `id` of a group in this catalog (or in a referenced catalog for imports)  |
| `assessment-requirements`| Yes      | List of verifiable conditions; each has `id`, `text`, `applicability`      |
| `threats`        | No       | Links to threat catalog(s) and threat IDs this control mitigates |
| `state`                  | No       | Lifecycle: `Active`, `Draft`, `Deprecated`, `Retired`; omitted is treated as `Active`. |

Each **assessment requirement** must have:

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
      Require signature validation so that only legitimate, trusted images are
      accepted; then pin each image to an immutable digest (e.g., sha256)
      after the check so that what is used matches what was verified and
      TOCTOU (time-of-check to time-of-use) attacks are prevented.
    group: SEC.SLAM.CM.FAM01
    assessment-requirements:
      - id: SEC.SLAM.CM.CTL01.AR01
        text: |
          The system MUST verify image signature before pull or run, then pin
          the image to a digest (e.g., sha256:...) after the check and use that
          digest for all subsequent use.
        applicability: ["all_deployments"]
      - id: SEC.SLAM.CM.CTL01.AR02
        text: |
          Configuration and policies MUST disallow or override use of tag-only
          references for production or sensitive workloads where supported.
        applicability: ["production"]
    threats:
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.THR01
          - reference-id: SEC.SLAM.CM.THR03
          - reference-id: SEC.SLAM.CM.THR04
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.TH14
  - id: SEC.SLAM.CM.CTL02
    title: Require TLS/SSL with Certificate Pinning
    objective: |
      Mitigate MITM Container Image Interception by protecting registry traffic and verifying artifact integrity: use
      TLS/SSL with certificate pinning for all registry communication, use VPNs on untrusted networks to reduce interception risk, and verify artifact signatures or hashes so that tampered or redirected content is detected even if the channel is compromised.
    group: SEC.SLAM.CM.FAM01
    state: Active
    assessment-requirements:
      - id: SEC.SLAM.CM.CTL02.AR01
        text: |
          The system MUST use TLS/SSL for all registry communication and MUST pin to the expected server certificate or public key (or certificate chain) for the registry.
        applicability: ["all_deployments"]
        state: Active
      - id: SEC.SLAM.CM.CTL02.AR02
        text: |
          On untrusted networks, the system or deployment pipeline MUST use a VPN or other trusted path for registry traffic, or MUST restrict
          image pulls to environments where the network is trusted.
        applicability: ["untrusted_networks", "ci_cd"]
        state: Active
      - id: SEC.SLAM.CM.CTL02.AR03
        text: |
          The system MUST verify artifact signatures or hashes (e.g. signature verification, digest check) before use so that tampered or redirected artifacts are rejected.
        applicability: ["all_deployments"]
        state: Active
    threats:
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.THR02
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.TH02
```

### Step 4: Validate Against the Gemara Schema

Validate the final catalog with CUE:

**Validation commands:**

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@v1 your-control-catalog.yaml
```

Fix any reported errors (e.g., missing required fields, invalid `group` reference, or malformed `threats`) so the catalog is schema-consistent.

### Step 5: Assemble the Full Catalog and Validate

Combine metadata, mapping-references, groups, `imports` (if any), and controls into a single YAML document. A complete, schema-valid catalog for the SEC.SLAM.CM scenario is in [control-catalog.yaml](control-catalog.yaml).

**Complete example (SEC.SLAM.CM):**

```yaml
title: Container Management Tool Security Control Catalog

metadata:
  id: SEC.SLAM.CM
  type: ControlCatalog
  gemara-version: "1.0.0"
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
        Threat catalog for the same scope; provides threat IDs referenced from each
        control's threats.
    - id: CCC
      title: Common Cloud Controls Core
      version: v2025.10
      url: https://github.com/finos/common-cloud-controls/releases
      description: |
        Foundational repository of reusable security controls, capabilities,
        and threat models maintained by FINOS.
  applicability-groups:
    - id: production
      title: Production
      description: |
        Production container workloads and clusters; controls apply to
        live environments where security posture is enforced.
    - id: all_deployments
      title: All Deployments
      description: |
        Requirements that apply whenever container images are built, pulled, or run—
        regardless of environment (dev, staging, production) or pipeline stage.
    - id: untrusted_networks
      title: Untrusted Networks
      description: |
        Applies when registry or image traffic traverses untrusted networks.
    - id: ci_cd
      title: CI/CD
      description: |
        Applies in continuous integration and deployment pipelines.

groups:
  - id: SEC.SLAM.CM.FAM01
    title: Image Integrity and Supply Chain
    description: |
      Controls that ensure container images are authentic, unmodified,
      and from trusted sources throughout retrieval and use.

imports:
  - reference-id: CCC
    entries:
      - reference-id: CCC.Core.CTL42
        remarks: Image signing and verification

controls:
  - id: SEC.SLAM.CM.CTL01
    title: Use Immutable Image References by Digest
    objective: |
      Require signature validation so that only legitimate, trusted images are
      accepted; then pin each image to an immutable digest (e.g., sha256)
      after the check so that what is used matches what was verified and
      TOCTOU (time-of-check to time-of-use) attacks are prevented.
    group: SEC.SLAM.CM.FAM01
    assessment-requirements:
      - id: SEC.SLAM.CM.CTL01.AR01
        text: |
          The system MUST verify image signature before pull or run, then pin
          the image to a digest (e.g., sha256:...) after the check and use that
          digest for all subsequent use.
        applicability: ["all_deployments"]
      - id: SEC.SLAM.CM.CTL01.AR02
        text: |
          Configuration and policies MUST disallow or override use of tag-only
          references for production or sensitive workloads where supported.
        applicability: ["production"]
    threats:
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.THR01
          - reference-id: SEC.SLAM.CM.THR03
          - reference-id: SEC.SLAM.CM.THR04
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.TH14
  - id: SEC.SLAM.CM.CTL02
    title: Require TLS/SSL with Certificate Pinning
    objective: |
      Mitigate MITM Container Image Interception by protecting registry
      traffic and verifying artifact integrity: use TLS/SSL with certificate
      pinning for all registry communication, use VPNs on untrusted networks
      to reduce interception risk, and verify artifact signatures or hashes so
      that tampered or redirected content is detected even if the channel is
      compromised.
    group: SEC.SLAM.CM.FAM01
    state: Active
    assessment-requirements:
      - id: SEC.SLAM.CM.CTL02.AR01
        text: |
          The system MUST use TLS/SSL for all registry communication and MUST
          pin to the expected server certificate or public key (or certificate
          chain) for the registry.
        applicability: ["all_deployments"]
        state: Active
      - id: SEC.SLAM.CM.CTL02.AR02
        text: |
          On untrusted networks, the system or deployment pipeline MUST use a
          VPN or other trusted path for registry traffic, or MUST restrict
          image pulls to environments where the network is trusted.
        applicability: ["untrusted_networks", "ci_cd"]
        state: Active
      - id: SEC.SLAM.CM.CTL02.AR03
        text: |
          The system MUST verify artifact signatures or hashes (e.g. signature
          verification, digest check) before use so that tampered or redirected
          artifacts are rejected.
        applicability: ["all_deployments"]
        state: Active
    threats:
      - reference-id: SEC.SLAM.CM
        entries:
          - reference-id: SEC.SLAM.CM.THR02
      - reference-id: CCC
        entries:
          - reference-id: CCC.Core.TH02
```
**Validation commands:**

Using the **published** module:

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@v1 your-control-catalog.yaml
```

Fix any reported errors (e.g., missing required fields, invalid `group` reference, or malformed `threats`) so the catalog is schema-consistent.

## Summary: From Threat Assessment to Control Catalog

| From threat assessment        | Use in control catalog                                      |
|-------------------------------|-------------------------------------------------------------|
| Scope / catalog id (e.g. SEC.SLAM.CM) | Same id in control catalog metadata |
| Threat IDs (e.g. SEC.SLAM.CM.THR01, CCC.Core.TH14) | `threats` on each control that mitigates those threats |
| Understanding of what goes wrong | Control objectives and assessment requirements that prevent or detect it |

## What's Next

Use the control catalog in downstream Gemara layers: for example, map controls to Layer 1 guidelines via `guidelines`, or use the catalog in Layer 5 evaluations to record which assessment requirements have been verified.

**Layer 2 schema documentation:** 
- [Threat Catalog](https://gemara.openssf.org/schema/threatcatalog.html)
- [Capability Catalog](https://gemara.openssf.org/schema/capabilitycatalog.html)
- [Control Catalog](https://gemara.openssf.org/schema/controlcatalog.html)
