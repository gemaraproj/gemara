---
layout: page
title: Guidance Catalog Guide
description: Step-by-step guide to creating Gemara-compatible guidance catalogs
---

## What This Is

This guide walks through creating a **guidance catalog** using the [Gemara](https://gemara.openssf.org/) project.

**The basic idea:** A guidance catalog is a structured set of **guidelines**—recommendations, requirements, or best practices—that help readers achieve desired outcomes. Guidelines are grouped into **families** and can reference external standards so you can align internal guidance with industry or regulatory sources.

In technical terms:
* **Guidance catalogs** have a **type** (Standard, Regulation, Best Practice, or Framework), **families** that group guidelines by theme, and **guidelines** with an objective, optional recommendations, and optional mappings to other artifacts.
* **Guidelines** state intent and context; they can map to external guidance (e.g., OWASP, NIST, HIPAA, GDPR, CRA, PCI, ISO) via `guideline-mappings` and to controls via downstream Gemara layers.

**Who might write guidance:** Authors can represent **internal** teams (unique organizational circumstances), **industry groups** (e.g., OWASP Top 10, PCI standards), **government agencies** (e.g., NIST Cybersecurity Framework, HIPAA), or **international standards bodies** (e.g., GDPR, CRA, ISO). Compliance professionals can use Gemara as a logical model for categorizing and mapping compliance activities to these sources.

## Walkthrough

### Step 0: Define Scope and Catalog Type

Choose the scope of your guidance (e.g., secure development, supply chain, data protection) and the **catalog type** that best fits intent:

| Type           | When to use                                                                 |
|----------------|-----------------------------------------------------------------------------|
| `Standard`     | Formal, normative specifications (e.g., ISO, PCI)                         |
| `Regulation`   | Legal or regulatory requirements (e.g., HIPAA, GDPR, CRA)                  |
| `Best Practice`| Non-mandatory recommendations (e.g., internal playbooks, OWASP-style)      |
| `Framework`    | High-level structure or taxonomy (e.g., NIST CSF)                          |

You can later add `mapping-references` to external documents (OWASP, NIST, HIPAA, GDPR, CRA, PCI, ISO) and use `guideline-mappings` on each guideline to align with those sources.

### Step 1: Setting Up Metadata

Declare your catalog and, if you will reference external standards, add mapping references. Key fields:

| Field                   | What It Is                                                | Why                                                                 |
|-------------------------|-----------------------------------------------------------|---------------------------------------------------------------------|
| `title`                 | Display name for the guidance catalog (top-level)         | Human-readable label in reports and tooling                         |
| `type`                  | One of Standard, Regulation, Best Practice, Framework      | Required by schema; clarifies intent                                |
| `metadata.id`           | Unique identifier for this catalog                        | Used when other artifacts reference this catalog                    |
| `metadata.mapping-references` | Pointers to external standards (e.g., OWASP, NIST)  | Resolve IDs used in `guideline-mappings` on guidelines             |

**Example (YAML):**

```yaml
title: Secure Software Development Guidance
type: Best Practice
metadata:
  id: ORG.SSD.001
  description: Internal secure development guidelines aligned to industry standards
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: OWASP
      title: OWASP Top 10
      version: "2021"
      url: https://owasp.org/Top10
      description: OWASP Top 10 Web Application Security Risks
```

### Step 2: Define Families

**Families** group guidelines by theme. The Layer 1 schema requires at least one family when the catalog defines `guidelines`. Each guideline’s `family` field must match the `id` of one of these groups (id, title, description).

**Example (YAML):**

```yaml
families:
  - id: ORG.SSD.FAM01
    title: Secure Dependencies and Supply Chain
    description: Guidelines for selecting, updating, and verifying dependencies and images.
```

### Step 3: Define Guidelines

**Guidelines** are the core content. Required fields for each guideline (see `layer-1.cue`):

| Field       | Required | Description                                              |
|-------------|----------|----------------------------------------------------------|
| `id`        | Yes      | Unique identifier (e.g., `ORG.SSD.GL01`)                 |
| `title`     | Yes      | Short name for the guideline                             |
| `objective` | Yes      | Unified statement of intent                              |
| `family`    | Yes      | `id` of a family in this catalog                         |

Optional: `state` (active, draft, deprecated, retired; defaults to active in schema), `recommendations`, `applicability`, `rationale`, `statements`, `guideline-mappings`, `vector-mappings`, and others (see `layer-1.cue`).

**Example (YAML):**

```yaml
guidelines:
  - id: ORG.SSD.GL01
    title: Prefer Immutable Image References
    objective: |
      Use digest-based or otherwise immutable references for container images
      to prevent tampering and ensure repeatable deployments.
    family: ORG.SSD.FAM01
    state: active
    recommendations:
      - Prefer pull-by-digest over tags for production.
      - Pin base image digests in Dockerfiles or equivalent.
    applicability: ["containerized workloads", "CI/CD"]
    guideline-mappings:
      - reference-id: OWASP
        entries:
          - reference-id: A06:2021-Vulnerable-and-Outdated-Components
            remarks: Reduces risk of outdated or substituted components
```

### Step 4: Validate

The catalog must conform to the Gemara Layer 1 schema in `layer-1.cue`. Validate with CUE:

**Validation commands:**

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#GuidanceCatalog' github.com/gemaraproj/gemara@latest your-guidance.yaml
```

Or from a local clone (repo root):

```bash
cue vet -c -d '#GuidanceCatalog' ./layer-1.cue ./metadata.cue ./mapping.cue ./base.cue docs/tutorials/guidance/guidance-example.yaml
```

### Minimal Full Example

A complete, schema-valid copy of this catalog is in [guidance-example.yaml](guidance-example.yaml) in this directory. Combined minimal catalog:

```yaml
title: Secure Software Development Guidance
type: Best Practice
metadata:
  id: ORG.SSD.001
  description: Internal secure development guidelines aligned to industry standards
  version: 1.0.0
  author:
    id: example
    name: Example
    type: Human
  mapping-references:
    - id: OWASP
      title: OWASP Top 10
      version: "2021"
      url: https://owasp.org/Top10
      description: OWASP Top 10 Web Application Security Risks

families:
  - id: ORG.SSD.FAM01
    title: Secure Dependencies and Supply Chain
    description: Guidelines for selecting, updating, and verifying dependencies and images.

guidelines:
  - id: ORG.SSD.GL01
    title: Prefer Immutable Image References
    objective: |
      Use digest-based or otherwise immutable references for container images
      to prevent tampering and ensure repeatable deployments.
    family: ORG.SSD.FAM01
    recommendations:
      - Prefer pull-by-digest over tags for production.
      - Pin base image digests in Dockerfiles or equivalent.
    applicability: ["containerized workloads", "CI/CD"]
    guideline-mappings:
      - reference-id: OWASP
        entries:
          - reference-id: A06:2021-Vulnerable-and-Outdated-Components
            remarks: Reduces risk of outdated or substituted components
```

## What's Next

Map guidelines to Layer 2 controls via control catalogs’ `guideline-mappings`, or reference this guidance from policy and evaluation layers. See the [Gemara Layer 1 schema documentation](https://gemara.openssf.org/schema/layer-1.html) for optional fields such as `exemptions`, `front-matter`, `rationale`, `statements`, and `principle-mappings` or `vector-mappings`.
