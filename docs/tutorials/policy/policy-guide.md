---
layout: page
title: Organizational Risk & Policy Guide
---

## What This Is

This guide walks through creating a **[Policy Document](../../model/02-definitions.html#policy)** using the [Gemara](https://gemara.openssf.org/) project. The document conforms to the Policy Definition in [`policy.cue`](https://github.com/gemaraproj/gemara/blob/main/policy.cue).

Terms to know:
* **RACI**: Who is responsible, accountable, consulted, and informed.
* **Scope**: What is in and out of scope (technologies, regions, sensitivity, users).
* **Imports**: Which external policies, Control Catalogs, and Guidance the policy references (and any exclusions, constraints, or assessment-requirement modifications).
* **Implementation plan**: When the policy becomes active (evaluation and enforcement timelines).
* **[Risks](../../model/02-definitions.html#risk)**: Which risks are mitigated vs accepted (with justification for accepted risks).
* **Adherence**: How compliance is evaluated and enforced (evaluation methods, assessment plans, enforcement methods, non-compliance handling).

This exercise produces a policy document that captures scope, imported controls and guidance, and how adherence is evaluated and enforced.

## Walkthrough

### Step 0: Metadata and mapping-references

Set `title` and `metadata` (see [metadata.cue](https://github.com/gemaraproj/gemara/blob/main/metadata.cue) for the standard metadata fields). Include `mapping-references` for every external catalog, guidance document, or policy you reference in `imports` (by `reference-id`). Key fields (see [`policy.cue`](https://github.com/gemaraproj/gemara/blob/main/policy.cue) and [metadata.cue](https://github.com/gemaraproj/gemara/blob/main/metadata.cue):

| Field                         | What It Is                                                                 | Why                                                                 |
|-------------------------------|----------------------------------------------------------------------------|---------------------------------------------------------------------|
| `title`                       | Display name for the policy (top-level)                                   | Human-readable label in reports and tooling                         |
| `metadata.id`                 | Unique identifier for this policy                                         | Used when other documents reference this policy                     |
| `metadata.type`               | Must be `Policy`                                                          | Required by schema; identifies the artifact kind                    |
| `metadata.gemara-version`     | Gemara specification version (e.g. `"1.0.0"`)                        | Required by schema; should match the module tag you pass to `cue vet` |
| `metadata.description`        | High-level summary of the policy's purpose and scope                       | Required by schema; clarifies intent                                |
| `metadata.author`             | Actor (id, name, type) primarily responsible for this policy              | Required by schema; identifies the author                            |
| `metadata.version`            | Version identifier (e.g. `"1.0.0"`)                                       | Optional; supports versioning and references                         |
| `metadata.mapping-references` | Pointers to external catalogs, guidance, or policies referenced in imports                   | Required for `imports`; each `reference-id` must match an entry here |

> **Note:** Include a `mapping-references` entry for every external catalog, guidance document, or policy you reference in `imports` (by `reference-id`).

**Example (YAML):**

```yaml
title: "Information Security Policy for Cloud and Web Applications"
metadata:
  id: "org-policy-001"
  type: Policy
  gemara-version: "1.0.0"
  description: "Policy for cloud and web application security; references control catalogs."
  version: "1.0.0"
  author:
    id: security-team
    name: "Security Team"
    type: Human
  mapping-references:
    - id: "SEC.SLAM.CM"
      title: "Container Management Tool Security Control Catalog"
      version: "1.0.0"
      description: "Control catalog for container management tool security."
```


### Step 1: Contacts

Define `contacts` with at least `responsible` and `accountable`. Each entry has `name`; optionally `affiliation` and `email`. Add `consulted` and `informed` if needed.

**Example (YAML):**

```yaml
contacts:
  responsible:
    - name: "Platform Engineering"
      affiliation: "Engineering"
      email: "platform@example.com"
  accountable:
    - name: "CISO"
      affiliation: "Security"
      email: "ciso@example.com"
  consulted:
    - name: "Legal"
      affiliation: "Legal"
  informed:
    - name: "All Engineering"
      affiliation: "Engineering"
```

### Step 2: Scope

Set `scope.in` (and optionally `scope.out`) with dimension fields that define where and to whom the policy applies:

| Field           | Purpose                    |
| --------------- | -------------------------- |
| `technologies`  | Tech stack or systems      |
| `geopolitical`  | Regions or jurisdictions   |
| `sensitivity`   | Data or asset sensitivity  |
| `users`         | User roles or populations  |
| `groups`        | Teams or org units         |

**Example (YAML):**

```yaml
scope:
  in:
    technologies:
      - "Cloud Computing"
      - "Web Applications"
    geopolitical:
      - "United States"
      - "European Union"
    users:
      - "developers"
      - "platform-engineers"
  out:
    technologies:
      - "Legacy On-Premises"
```

### Step 3: Imports

Under `imports`:

- **`policies`** — List of external policy imports. Each entry: reference-id (must match metadata.mapping-references) to reference other policy documents this policy inherits or extends.
- **`guidance`** — List of guidance imports used when the policy aligns to Layer 1 Guidance Catalogs. Each entry: reference-id (match metadata.mapping-references), optional exclusions and constraints.
- **`catalogs`** — List of catalog imports used when the policy references Layer 2 Control Catalogs. Use assessment-requirement-modifications to tailor how assessment requirements are applied (Add, Modify, Remove, Replace, or Override).
 
Ensure each `reference-id` appears in `metadata.mapping-references`.

**Example (YAML):**

```yaml
imports:
  catalogs:
    - reference-id: "SEC.SLAM.CM"
      assessment-requirement-modifications:
        - id: "CTL02-AR01-strict"
          target-id: "SEC.SLAM.CM.CTL02.AR01"
          modification-type: Override
          modification-rationale: "Require TLS and certificate pinning for all registry communication in this org."
          text: "The system MUST use TLS/SSL for all registry communication and MUST pin to the expected server certificate or public key (or certificate chain) for the registry."
        - id: "CTL02-AR02-strict"
          target-id: "SEC.SLAM.CM.CTL02.AR02"
          modification-type: Override
          modification-rationale: "Require VPN or trusted path on untrusted networks for registry traffic in this org."
          text: "On untrusted networks, the system or deployment pipeline MUST use a VPN or other trusted path for registry traffic, or MUST restrict image pulls to environments where the network is trusted."

```

### Step 4: Implementation plan (optional)

Add `implementation-plan` with `evaluation-timeline` and `enforcement-timeline`. Each has `start`, optional `end`, and `notes`. Optionally add `notification-process` (string).

**Example (YAML):**

```yaml
implementation-plan:
  notification-process: "Policy communicated via internal wiki and team leads; rollout via Platform Engineering."
  evaluation-timeline:
    start: "2025-03-01T00:00:00Z"
    end: "2025-06-01T00:00:00Z"
    notes: "Initial evaluation phase; automated checks rolled out by Q2."
  enforcement-timeline:
    start: "2025-06-01T00:00:00Z"
    notes: "Enforcement begins after evaluation baseline is established."
```

### Step 5: Adherence

Define `adherence` with at least one of the following:

| Field                   | Purpose                                                                 |
| ----------------------- | ----------------------------------------------------------------------- |
| `evaluation-methods`    | List of methods: required `id`; `type` (`Behavioral`, `Intent`, `Remediation`, `Gate`); `mode` (`Manual`, `Automated`); optional `required`, `description`, `executor` |
| `assessment-plans`      | Plans: required `id`, `requirement-id`, `frequency`, and `evaluation-methods`; optional `evidence-requirements`, `parameters` |
| `enforcement-methods`   | Same structure as `evaluation-methods` (required `id`, `type`, `mode`) |
| `non-compliance`        | String describing handling of non-compliance                            |

**Example (YAML):**

```yaml
adherence:
  evaluation-methods:
    - id: "EV-AUTO-01"
      type: "Behavioral"
      mode: "Automated"
      required: true
      description: "CI pipeline runs control checks via Privateer."
    - id: "EV-MANUAL-01"
      type: "Behavioral"
      mode: "Manual"
      required: true
      description: "Quarterly review of exception requests."
  assessment-plans:
    - id: "plan-ctl01-ar01"
      requirement-id: "SEC.SLAM.CM.CTL01.AR01"
      frequency: "every push"
      evaluation-methods:
        - id: "EV-AUTO-02"
          type: "Behavioral"
          mode: "Automated"
          required: true
  enforcement-methods:
    - id: "EM-GATE-01"
      type: "Gate"
      mode: "Automated"
      required: true
      description: "Block merge if control check fails."
  non-compliance: "Non-compliance is reported to responsible contacts and tracked in issue tracker; critical failures block deployment."
```

## Step 6: Validation

The policy must conform to the Policy Definition defined in the CUE module. Validate with CUE:

**Validation commands:**

Using the **published** module:

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#Policy' github.com/gemaraproj/gemara@v1 your-policy.yaml
```

Fix any errors (e.g. missing required fields, invalid reference-ids, or type mismatches) so the policy is schema-valid.

## Minimal Full Example

A complete, schema-valid copy of this policy is in [policy-example.yaml](policy-example.yaml) in this directory. The following combines the snippets above into a single policy document. Omit optional sections (e.g. implementation-plan, risks) if not needed. 

```yaml
title: "Information Security Policy for Cloud and Web Applications"
metadata:
  id: "org-policy-001"
  type: Policy
  gemara-version: "1.0.0"
  description: "Policy for cloud and web application security; references control catalogs."
  version: "1.0.0"
  author:
    id: security-team
    name: "Security Team"
    type: Human
  mapping-references:
    - id: "SEC.SLAM.CM"
      title: "Container Management Tool Security Control Catalog"
      version: "1.0.0"
      description: "Control catalog for container management tool security."

contacts:
  responsible:
    - name: "Platform Engineering"
      affiliation: "Engineering"
      email: "platform@example.com"
  accountable:
    - name: "CISO"
      affiliation: "Security"
      email: "ciso@example.com"

scope:
  in:
    technologies:
      - "Cloud Computing"
      - "Web Applications"
    geopolitical:
      - "United States"
      - "European Union"

imports:
  catalogs:
    - reference-id: "SEC.SLAM.CM"
      assessment-requirement-modifications:
        - id: "CTL02-AR01-strict"
          target-id: "SEC.SLAM.CM.CTL02.AR01"
          modification-type: Override
          modification-rationale: "Require TLS and certificate pinning for all registry communication in this org."
          text: "The system MUST use TLS/SSL for all registry communication and MUST pin to the expected server certificate or public key (or certificate chain) for the registry."
        - id: "CTL02-AR02-strict"
          target-id: "SEC.SLAM.CM.CTL02.AR02"
          modification-type: Override
          modification-rationale: "Require VPN or trusted path on untrusted networks for registry traffic in this org."
          text: "On untrusted networks, the system or deployment pipeline MUST use a VPN or other trusted path for registry traffic, or MUST restrict image pulls to environments where the network is trusted."

adherence:
  evaluation-methods:
    - id: "EV-AUTO-01"
      type: "Behavioral"
      mode: "Automated"
      required: true
      description: "CI pipeline runs control checks."
  non-compliance: "Non-compliance is reported to responsible contacts and tracked."
```

## What's Next

- Use the policy in **Layer 5** evaluations to determine whether implementations conform.
- Use **Layer 7** audit and continuous monitoring to assess policy effectiveness.

See the [Policy schema documentation](https://gemara.openssf.org/schema/policy.html) and [policy.cue](https://github.com/gemaraproj/gemara/blob/main/policy.cue) for the full specification.
