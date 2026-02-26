---
layout: page
title: Organizational Risk & Policy Guide
---

This guide walks through creating a **policy document** using the [Gemara](https://gemara.openssf.org/) project. The document conforms to the **Layer 3** schema in `layer-3.cue`.

## What You'll Build

A **policy document** that defines:

- **Who** is responsible, accountable, consulted, and informed (RACI).
- **What** is in and out of scope (technologies, regions, sensitivity, users).
- **Which** external policies, control catalogs, and guidance the policy imports (and any exclusions, constraints, or assessment-requirement modifications).
- **When** the policy becomes active (implementation plan: evaluation and enforcement timelines).
- **Which risks** are mitigated vs accepted (with justification for accepted risks).
- **How** adherence is evaluated and enforced (evaluation methods, assessment plans, enforcement methods, non-compliance handling).

The policy conforms to the Gemara Layer 3 schema in `layer-3.cue`. Top-level structure:

| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes | Display name for the policy. |
| `metadata` | Yes | id, description, author, optional version, date, mapping-references. |
| `contacts` | Yes | RACI: responsible, accountable, optional consulted, informed. Each entry: name, optional affiliation, optional email. |
| `scope` | Yes | `in` and optional `out` as dimensions (technologies, geopolitical, sensitivity, users, groups). |
| `imports` | Yes | Optional policies, catalogs, and guidance. |
| `implementation-plan` | No | Optional notification-process; evaluation-timeline and enforcement-timeline with start, optional end, notes. |
| `risks` | No | Optional mitigated (mappings to risk entries) and/or accepted (risk entries with optional scope and justification). |
| `adherence` | Yes | Optional evaluation-methods, assessment-plans, enforcement-methods, non-compliance. |

**Imports:**

- **`imports.guidance`** — List of guidance imports used when the policy aligns to Layer 1 guidance catalogs. Use reference-id (must match metadata.mapping-references). Optional exclusions and constraints (id, target-id, text).
- **`imports.catalogs`** — List of catalog imports used when the policy references Layer 2 control catalogs. Use assessment-requirement-modifications to tailor how assessment requirements are applied (add, modify, remove, replace, override).

**Adherence:**

- **`adherence.evaluation-methods`** / **`adherence.enforcement-methods`** — Lists of accepted methods. Type: manual, behavioral, automated, autoremediation, or gate. Optional description and executor.
- **`adherence.assessment-plans`** — List of assessment plans: id, requirement-id, frequency, evaluation-methods, optional evidence-requirements and parameters.
- **`adherence.non-compliance`** — String describing how non-compliance is communicated or handled.

Use [layer-3.cue](https://github.com/gemaraproj/gemara/blob/main/layer-3.cue) as the source of truth for required vs optional fields and nested structures.

## Scenario

Your organization needs a policy that applies to cloud and web applications in specific regions. You will:

1. Set metadata and mapping-references so you can reference external control catalogs and guidance by id.
2. Define contacts (responsible, accountable, consulted, informed).
3. Set scope (in-scope: e.g. technologies, geopolitical regions; optionally out-of-scope).
4. Import control catalogs and/or guidance, with optional constraints or assessment-requirement-modifications.
5. Optionally add an implementation-plan (evaluation and enforcement timelines).
6. Optionally list risks (mitigated and accepted with justification).
7. Define adherence (evaluation methods, assessment plans, enforcement methods, non-compliance.

## Walkthrough

### Step 1: Metadata and mapping-references

Set `title` and `metadata` (see [metadata.cue](https://github.com/gemaraproj/gemara/blob/main/metadata.cue) for the standard metadata fields). Include `mapping-references` for every external catalog or guidance document you reference in `imports` (by `reference-id`).

**Example (YAML):**

```yaml
title: Information Security Policy for Cloud and Web Applications
metadata:
  id: org-policy-001
  description: Policy for cloud and web application security; references control catalogs and guidance.
  version: 1.0.0
  author:
    id: security-team
    name: Security Team
    type: Human
  mapping-references:
    - id: SEC.SLAM.CM
      title: Container Management Tool Security Control Catalog
      version: "1.0.0"
      description: Control catalog for container management tool security.
    - id: ORG.SSD.001
      title: Secure Software Development Guidance
      version: "1.0.0"
      description: Internal secure development guidelines.
    - id: ORG-RISK
      title: Organization Risk Catalog
      version: "1.0.0"
      description: Internal risk catalog for policy risk mappings.
```

### Step 2: Contacts

Define `contacts` with at least `responsible` and `accountable`. Each entry has `name`; optionally `affiliation` and `email`. Add `consulted` and `informed` if needed.

**Example (YAML):**

```yaml
contacts:
  responsible:
    - name: Platform Engineering
      affiliation: Engineering
      email: platform@example.com
  accountable:
    - name: CISO
      affiliation: Security
      email: ciso@example.com
  consulted:
    - name: Legal
      affiliation: Legal
  informed:
    - name: All Engineering
      affiliation: Engineering
```

### Step 3: Scope

Set `scope.in` (and optionally `scope.out`) with dimension fields such as `technologies`, `geopolitical`, `sensitivity`, `users`, `groups`. These define where and to whom the policy applies.

**Example (YAML):**

```yaml
scope:
  in:
    technologies:
      - Cloud Computing
      - Web Applications
    geopolitical:
      - United States
      - European Union
    users:
      - developers
      - platform-engineers
  out:
    technologies:
      - Legacy On-Premises
```

### Step 4: Imports

Under `imports`:

- **`guidance`** — List of guidance imports used when the policy aligns to Layer 1 guidance catalogs. Each entry: reference-id (match metadata.mapping-references), optional exclusions and constraints.
- **`catalogs`** — List of catalog imports used when the policy references Layer 2 control catalogs. Use assessment-requirement-modifications to tailor how assessment requirements are applied (add, modify, remove, replace, override).

Ensure each `reference-id` appears in `metadata.mapping-references`.

**Example (YAML):**

```yaml
imports:
  guidance:
    - reference-id: ORG.SSD.001
      constraints:
        - id: GL01-annual-review
          target-id: ORG.SSD.GL01
          text: Review guideline applicability at least annually.
  catalogs:
    - reference-id: SEC.SLAM.CM
      assessment-requirement-modifications:
        - id: CTL01-AR01-strict
          target-id: SEC.SLAM.CM.CTL01.AR01
          modification-type: override
          modification-rationale: Require digest in all environments for this org.
          text: The system MUST resolve image references to a digest before pull or run in all environments.
```

### Step 5: Implementation plan (optional)

Add `implementation-plan` with `evaluation-timeline` and `enforcement-timeline`. Each has `start` (ISO 8601 datetime), optional `end`, and `notes`. Optionally add `notification-process` (string).

**Example (YAML):**

```yaml
implementation-plan:
  notification-process: Policy communicated via internal wiki and team leads.
  evaluation-timeline:
    start: 2025-03-01T00:00:00Z
    end: 2025-06-01T00:00:00Z
    notes: Initial evaluation phase; automated checks rolled out by Q2.
  enforcement-timeline:
    start: 2025-06-01T00:00:00Z
    notes: Enforcement begins after evaluation baseline is established.
```

### Step 6: Risks (optional)

Add `risks` with `mitigated` (mappings to risk catalogs/entries) and/or `accepted` (accepted-risk entries with risk reference, optional scope, optional justification).

**Example (YAML):**

```yaml
risks:
  mitigated:
    - reference-id: ORG-RISK
      entries:
        - reference-id: ORG-RISK.R.01
      remarks: Image tampering risk addressed by control catalog.
  accepted:
    - risk:
        reference-id: ORG-RISK
        entry-id: ORG-RISK.R.LOW-01
      justification: Accepted for non-production workloads with no sensitive data.
```

### Step 7: Adherence

Define `adherence` with at least one of: `evaluation-methods`, `assessment-plans`, `enforcement-methods`, or `non-compliance`. Methods use type (manual, behavioral, automated, autoremediation, gate) and optional description and executor. Assessment-plans use requirement-id, frequency, evaluation-methods, and optionally evidence-requirements and parameters. Set `non-compliance` to a string as needed.

**Example (YAML):**

```yaml
adherence:
  evaluation-methods:
    - type: automated
      description: CI pipeline runs control checks via Privateer.
    - type: manual
      description: Quarterly review of exception requests.
  assessment-plans:
    - id: plan-ctl01
      requirement-id: SEC.SLAM.CM.CTL01.AR01
      frequency: every push
      evaluation-methods:
        - type: automated
  enforcement-methods:
    - type: gate
      description: Block merge if control check fails.
  non-compliance: Non-compliance is reported to responsible contacts and tracked in issue tracker; critical failures block deployment.
```

## Example YAML (combined)

The following combines the snippets above into a single policy document. Omit optional sections (e.g. implementation-plan, risks) if not needed.

```yaml
title: Information Security Policy for Cloud and Web Applications
metadata:
  id: org-policy-001
  description: Policy for cloud and web application security; references control catalogs and guidance.
  version: 1.0.0
  author:
    id: security-team
    name: Security Team
    type: Human
  mapping-references:
    - id: SEC.SLAM.CM
      title: Container Management Tool Security Control Catalog
      version: "1.0.0"
      description: Control catalog for container management tool security.
    - id: ORG.SSD.001
      title: Secure Software Development Guidance
      version: "1.0.0"
      description: Internal secure development guidelines.

contacts:
  responsible:
    - name: Platform Engineering
      affiliation: Engineering
      email: platform@example.com
  accountable:
    - name: CISO
      affiliation: Security
      email: ciso@example.com

scope:
  in:
    technologies:
      - Cloud Computing
      - Web Applications
    geopolitical:
      - United States
      - European Union

imports:
  guidance:
    - reference-id: ORG.SSD.001
  catalogs:
    - reference-id: SEC.SLAM.CM

adherence:
  evaluation-methods:
    - type: automated
      description: CI pipeline runs control checks.
  non-compliance: Non-compliance is reported to responsible contacts and tracked.
```

## Validation

Validate against the Gemara Layer 3 schema:

```bash
cue vet -c -d '#Policy' ./layer-3.cue ./metadata.cue ./mapping.cue ./base.cue your-policy.yaml
```

Or with the module:

```bash
cue vet -c -d '#Policy' github.com/gemaraproj/gemara@latest your-policy.yaml
```

Fix any errors (e.g. missing required fields, invalid reference-ids, or type mismatches) so the policy is schema-valid.

## What's Next

- Use the policy in **Layer 5** evaluations to determine whether implementations conform.
- Use **Layer 7** audit and continuous monitoring to assess policy effectiveness.

See the [Layer 3 schema documentation](https://gemara.openssf.org/schema/layer-3.html) and [layer-3.cue](https://github.com/gemaraproj/gemara/blob/main/layer-3.cue) for the full specification.
