# Control Catalog

## What This Is

This document is the **control catalog** produced by following the [Control Catalog Guide](control-catalog-guide.md). It uses the same scope as the [Threat Assessment Guide](threat-assessment.md) (container management tool **SEC.SLAM.CM**), maps controls to the threats defined there, and conforms to the Gemara Layer 2 schema (`#ControlCatalog` in `layer-2.cue`).

## Scenario

**Pang**, the Product Security Engineer, has completed the threat assessment for the container management tool and identified threats such as **Container Image Tampering or Poisoning** (SEC.SLAM.CM.THR01) and the imported **Older Resource Versions are Used** (CCC.Core.TH14). This control catalog defines safeguards that mitigate those threats and can be used for evaluations and guideline mappings.

## Example YAML

The following YAML is a complete, schema-valid control catalog. A copy of this catalog is in `control-catalog.yaml` in this directory. Validate it with (from the repo root):

```bash
cue vet -c -d '#ControlCatalog' ./layer-2.cue ./metadata.cue ./mapping.cue ./base.cue docs/tutorials/controls/control-catalog.yaml
```

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

## Validation

From the repo root, validate the catalog against the Gemara Layer 2 schema:

```bash
cue vet -c -d '#ControlCatalog' ./layer-2.cue ./metadata.cue ./mapping.cue ./base.cue docs/tutorials/controls/control-catalog.yaml
```
