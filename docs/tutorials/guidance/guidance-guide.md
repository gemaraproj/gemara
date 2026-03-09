---
layout: page
title: Guidance Catalog Guide
description: Step-by-step guide to creating Gemara-compatible guidance catalogs
---

## What This Is

This guide walks through creating a **Guidance Catalog** using the [Gemara](https://gemara.openssf.org/) project.

**The basic idea:** A Guidance Catalog is a structured set of **guidelines**—recommendations, requirements, or best practices—that help readers achieve desired outcomes. Guidelines are grouped into **families**.

In technical terms:
* **Guidance catalogs** have a **type** (Standard, Regulation, Best Practice, or Framework), **families** that group guidelines by theme, and **guidelines** with an objective, optional recommendations, and optional references to other guidelines within the *same* guidance catalog.
* **Guidelines:** state the intent and context; they have statements which act as sub-requirements of the guideline (e.g., `ORG.SSD.001` and statements.id `ORG.SSD.001.1`). The guidance includes a see-also for linking other guidelines within the same guidance catalog (e.g., `ORG.SSD.001` see-also `ORG.SSD.002`, `ORG.SSD.003`).
* **Guidelines** have the ability to be mapped to external guidance (e.g., OWASP, NIST, HIPAA, GDPR, CRA, PCI, ISO) and to controls in a *separate* **Mapping Document**. Downstream Gemara Layers can reference `guidelines`, defining support for specific controls.

> **Coming Soon:** Mapping Document Tutorial.

**Who might write guidance:** Authors can represent **internal** teams (unique organizational circumstances), **industry groups** (e.g., OWASP Top 10, PCI standards), **government agencies** (e.g., NIST Cybersecurity Framework, HIPAA), or **international standards bodies** (e.g., GDPR, CRA, ISO). Compliance professionals can use Gemara as a logical model for categorizing and mapping compliance activities to these sources.

## Walkthrough

### Step 0: Define Scope and Catalog Type

Choose the scope of your guidance (e.g., secure development, supply chain, data protection) and the **catalog type** that best fits intent:

| Type           | When to use                                                                 |
|----------------|-----------------------------------------------------------------------------|
| `Standard`     | Formal, normative specifications (e.g., ISO 27001, PCI-DSS, NIST 800-53)                         |
| `Regulation`   | Legal or regulatory requirements (e.g., HIPAA, GDPR, CRA)                  |
| `Best Practice`| Non-mandatory recommendations (e.g., internal playbooks, OWASP-style)      |
| `Framework`    | High-level structure or taxonomy (e.g., NIST CSF)                          |

You can later add `mapping-references` to external documents and use an **external** [Mapping Document](https://gemara.openssf.org/schema/mapping.html) (Tutorial Coming Soon) to align those sources. 

### Step 1: Setting Up Metadata

Declare your catalog and, if you will reference external standards, add mapping references. Key fields:

| Field                         | What It Is                                                                 | Why                                                                 |
|-------------------------------|----------------------------------------------------------------------------|---------------------------------------------------------------------|
| `title`                       | Display name for the guidance catalog (top-level)                         | Human-readable label in reports and tooling                         |
| `type`                        | One of Standard, Regulation, Best Practice, Framework (catalog intent)     | Required by schema; clarifies intent                                |
| `metadata.id`                 | Unique identifier for this catalog                                        | Used when other artifacts reference this catalog                    |
| `metadata.type`               | Artifact kind (e.g. `GuidanceCatalog`)                                    | Required by schema; identifies the Gemara artifact type              |
| `metadata.gemara-version`     | Gemara specification version (e.g. `"0.20.0"`)                            | Required by schema; declares which spec the artifact conforms to     |
| `metadata.mapping-references` | Pointers to external standards (e.g., OWASP, NIST)                        | Resolve IDs used in external Mapping Document on guidelines. |
| `metadata.applicability-categories` | List of categories (id, title, description) for when guidelines apply | Define scope so guidelines reference these ids in `applicability`; keeps applicability consistent and documented |

**Example (YAML):**

```yaml
title: Secure Software Development Guidance
metadata:
  id: ORG.SSD.001
  type: GuidanceCatalog
  gemara-version: "0.20.0"
  description: Internal secure development and supply chain security guidelines (dependencies, images, and development practices) aligned to industry standards
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
  applicability-categories:
    - id: containerized_workloads
      title: Containerized Workloads
      description: Guidelines that apply to container-based deployments and images.
    - id: ci_cd
      title: CI/CD
      description: Guidelines that apply in continuous integration and deployment pipelines.
    - id: github_repositories
      title: GitHub Repositories
      description: Guidelines that apply to projects using GitHub for source and collaboration.
type: Best Practice
```

> **Minimal Mapping Document example:** A Mapping Document that maps this guidance catalog’s guidelines to OWASP Top 10 (source `ORG-SSD`, target `OWASP`) is in [mapping-document.yaml](mapping-document.yaml).

### Step 2: Define Families

**Families** group guidelines by theme. The Guidance Catalog schema requires at least one family when the catalog defines `guidelines`. Each guideline’s `family` field must match the `id` of one of these groups (id, title, description).

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
| `state`     | Yes      | Lifecycle: `Active`, `Draft`, `Deprecated`, or `Retired` |

Optional: `recommendations`, `applicability`, `rationale`, `statements`, `guidelines`, `vectors`, and others (see `layer-1.cue`).

**Applicability:** When you define `metadata.applicability-categories` in Step 1, use those category **ids** in each guideline’s `applicability` list (e.g. `["containerized_workloads", "ci_cd"]`). That keeps applicability consistent and documented.

**Example (YAML):** The following guidelines illustrate supply chain security for dependencies and images (artifact integrity, source/code integrity, and secure transit):

```yaml
guidelines:
  - id: ORG.SSD.GL01
    title: Prefer Immutable Image References
    objective: |
      Use digest-based or immutable references for container images to prevent
      tampering and ensure repeatable deployments.
    family: ORG.SSD.FAM01
    state: active
    recommendations:
      - Prefer pull-by-digest over tags for production.
      - Pin base image digests in Dockerfiles or equivalent.
    applicability: ["containerized_workloads", "ci_cd"]
    see-also:
      - ORG.SSD.GL02
      - ORG.SSD.GL03
  - id: ORG.SSD.GL02
    title: Prefer GitHub Branch Protection Rules 
    objective: |
      Use branch protection so only approved changes reach the main branch and
      malicious code cannot be merged without review.
    family: ORG.SSD.FAM01
    state: Active
    recommendations:
      - Prefer pull requests submitted from fork branch.
      - Required maintainer/non-author review and approval for merge.
      - Prefer GitHub Actions Quality checks in CI on pull request.
    applicability: ["containerized_workloads", "ci_cd", "github_repositories"]
    see-also:
      - ORG.SSD.GL01
  - id: ORG.SSD.GL03
    title: Prefer VPN on Untrusted Networks
    objective: |
      Use a VPN on untrusted networks to protect traffic from interception and
      DNS spoofing.
    family: ORG.SSD.FAM01
    state: Active
    recommendations:
      - Use a VPN for registry and build traffic on untrusted networks.
    see-also:
      - ORG.SSD.GL02 
```

### Step 4: Validate

The catalog must conform to the Guidance Catalog Definition defined in the CUE module. Validate with CUE:

**Validation commands:**

Using the **published** module:

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#GuidanceCatalog' github.com/gemaraproj/gemara@latest your-guidance.yaml
```

### Minimal Full Example

A complete, schema-valid copy of this catalog is in [guidance-example.yaml](guidance-example.yaml) in this directory. Combined minimal catalog:

```yaml
title: Secure Software Development Guidance
metadata:
  id: ORG.SSD.001
  type: GuidanceCatalog
  gemara-version: "0.20.0"
  description: Internal secure development and supply chain security guidelines (dependencies, images, and development practices) aligned to industry standards
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
  applicability-categories:
    - id: containerized_workloads
      title: Containerized Workloads
      description: Guidelines that apply to container-based deployments and images.
    - id: ci_cd
      title: CI/CD
      description: Guidelines that apply in continuous integration and deployment pipelines.
    - id: github_repositories
      title: GitHub Repositories
      description: Guidelines that apply to projects using GitHub for source and collaboration.
type: Best Practice
front-matter: Example best-practices text for tutorials developed by Gemara maintainers.
families:
  - id: ORG.SSD.FAM01
    title: Secure Dependencies and Supply Chain
    description: Guidelines for selecting, updating, and verifying dependencies and images.

guidelines:
  - id: ORG.SSD.GL01
    title: Prefer Immutable Image References
    objective: |
      Use digest-based or immutable references for container images to prevent
      tampering and ensure repeatable deployments.
    family: ORG.SSD.FAM01
    state: Active
    recommendations:
      - Prefer pull-by-digest over tags for production.
      - Pin base image digests in Dockerfiles or equivalent.
    applicability: ["containerized_workloads", "ci_cd"]
    see-also:
      - ORG.SSD.GL02
      - ORG.SSD.GL03
  - id: ORG.SSD.GL02
    title: Prefer GitHub Branch Protection Rules
    objective: |
      Use branch protection so only approved changes reach the main branch and
      malicious code cannot be merged without review.
    family: ORG.SSD.FAM01
    state: Active
    recommendations:
      - Prefer pull requests submitted from fork branch.
      - Required maintainer/non-author review and approval for merge.
      - Prefer GitHub Actions Quality checks in CI on pull request.
    applicability: ["containerized_workloads", "ci_cd", "github_repositories"]
    see-also:
      - ORG.SSD.GL01
  - id: ORG.SSD.GL03
    title: Prefer VPN on Untrusted Networks
    objective: |
      Use a VPN on untrusted networks to protect traffic from interception and
      DNS spoofing.
    family: ORG.SSD.FAM01
    state: Active
    recommendations:
      - Use a VPN for registry and build traffic on untrusted networks.
    applicability: ["containerized_workloads", "ci_cd"]
    see-also:
      - ORG.SSD.GL02
```

## What's Next

Map guidelines to Layer 2 controls via control catalogs’ `guidelines`, or reference this guidance from a Policy. See the [schema documentation](https://gemara.openssf.org/schema/layer-1.html) for optional fields such as `exemptions`, `see-also`, `replaced-by`, `front-matter`, `rationale`, `statements`, and `principles` or `vectors`.
