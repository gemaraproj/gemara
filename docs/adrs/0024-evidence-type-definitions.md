---
layout: page
title: Evidence Type Definitions
---

- **ADR:** 0024
- **Proposal Author(s):** @jpower432
- **Status:** Proposed (**Companion to:** ADR-0023)

## Context

ADR-0023 adds `#EvidenceMapping` to `#Evidence` so evaluation tools can record where evidence was observed, with sub-origin coordinates and integrity digests. That addresses the Layer 5 question: *what did we observe and where did it come from?*

This ADR addresses the Layer 3 counterpart: *what do evidence types mean?*

`#AssessmentPlan` currently has `frequency` (a bare string) and `evidence-requirements` (a bare string). These fields are loosely typed and disconnected. `frequency` conflates two distinct concepts: how often the control is reviewed, and how long a specific kind of evidence remains valid. An API scan goes stale after a day even though the control review may not be due for months. A red team report stays current across multiple review cycles. There is no way for tooling to compute either dimension.

## Action

Split the bare-string `frequency` into two typed fields at their natural levels and replace `evidence-requirements` with structured `#EvidenceTypeDefinition`.

```cue
#EvidenceTypeDefinition: {
    // id matches the type field on #Evidence instances
    id: string

    // description explains what this evidence type represents
    description?: string

    // valid-for is the evidence validity period in days.
    // Evidence older than this many days is considered stale.
    "valid-for": int @go(ValidFor)
}
```

On `#AssessmentPlan`, type `frequency` as an integer (control review frequency in days) and add the evidence type registry:

```cue
#AssessmentPlan: {
    id:               string
    "requirement-id": string @go(RequirementId)
    // frequency is the control review frequency in days
    frequency:        int
    "evidence-types":  [#EvidenceTypeDefinition, ...#EvidenceTypeDefinition] @go(EvidenceTypes)
    "evaluation-methods": [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)
    parameters?: [#Parameter, ...#Parameter]
}
```

`frequency` and `valid-for` answer different questions:

- `frequency` (plan-level): "How often is this control reviewed?" — drives the audit schedule.
- `valid-for` (evidence-type-level): "How long does this evidence stay current?" — drives staleness alerts.

## End-to-end example

This traces a single requirement (OSPS-AC-01: MFA enforcement) through all layers.

### Layer 3 — Policy defines evidence types

```yaml
# Policy artifact (type: Policy)
metadata:
  id: "acme-security-policy"
  type: Policy
  gemara-version: "1.3.0"
  mapping-references:
    - id: "OSPS-B"
      title: "OpenSSF Baseline for Open Source Projects"
      version: "2025.02.25"
      url: "https://baseline.openssf.org"

imports:
  catalogs:
    - reference-id: "OSPS-B"

adherence:
  evaluation-methods:
    - id: "EV-AUTO-01"
      type: "Behavioral"
      mode: "Automated"
      required: true
      description: "Automated API and configuration scanning"
    - id: "EV-MANUAL-01"
      type: "Behavioral"
      mode: "Manual"
      description: "Quarterly human review of access controls"
  assessment-plans:
    - id: "plan-ac01"
      requirement-id: "OSPS-AC-01.01"
      frequency: 90
      evidence-types:
        - id: "api-response"
          description: "Live API query of platform settings"
          valid-for: 1
        - id: "config-file"
          description: "Security Insights or configuration file check"
          valid-for: 1
        - id: "quarterly-access-review"
          description: "Human review of MFA enforcement status"
          valid-for: 90
      evaluation-methods:
        - id: "EV-AUTO-01"
          type: "Behavioral"
          mode: "Automated"
          required: true
```

### Layer 4 — EvaluationLog produces evidence (ADR-0023 schema)

An evaluator (e.g., PVTR) produces an EvaluationLog. The assessment log references the plan and attaches evidence. Only the relevant assessment-log fragment is shown here (see ADR-0023 for the full `#Evidence` + `#EvidenceMapping` structure):

```yaml
# Within an EvaluationLog's assessment-logs:
- requirement:
    entry-id: "OSPS-AC-01.01"
  plan:
    reference-id: "acme-policy"
    entry-id: "plan-ac01"
  result: Passed
  message: "Two-factor authentication is enforced by the organization"
  confidence-level: High
  start: "2026-06-25T14:00:00Z"
  end: "2026-06-25T14:00:01Z"
  evidence:
    - id: "EV-AC-01-01"
      type: "api-response"
      collected-at: "2026-06-25T14:00:00Z"
      origin:
        reference-id: "github-api"
        coordinate: "/orgs/acme/settings"
      payload:
        login: acme
        two_factor_requirement_enabled: true
      description: "GitHub API confirms org-wide MFA enforcement"
    - id: "EV-AC-01-02"
      type: "config-file"
      collected-at: "2026-06-25T14:00:01Z"
      origin:
        reference-id: "security-insights"
        coordinate: "SECURITY-INSIGHTS.yml:8"
        digest: "sha256:e9f0a1b2c3..."
      payload:
        mfa:
          required: true
          enforcement: "organization"
      description: "Security Insights confirms MFA policy"
```

### Consumer — walking the evidence chain

A consumer resolving posture for a requirement uses both ADRs together. `#EvidenceTypeDefinition` (this ADR) tells the consumer what evidence is expected and whether it is fresh. `#EvidenceMapping` (ADR-0023) tells the consumer where each piece of evidence came from.

```
Given an assessment log for requirement OSPS-AC-01.01:

1. Control review due?
   plan.frequency → 90 days
   Last review: 15 days ago → NOT DUE

2. For each evidence entry:
   a. Is it fresh?
      evidence.type → "api-response"
      Look up evidence-type definition by id → valid-for: 1 day
      evidence.collected-at → 2h ago → CURRENT

   b. Where did it come from?
      evidence.origin → #EvidenceMapping
      origin.reference-id → "github-api" → resolve in mapping-references
        → url: "https://api.github.com", version: "2026"
      origin.coordinate → "/orgs/acme/settings"
      origin.digest → (none — live API, not content-addressable)

   Result: current api-response evidence from GitHub API at /orgs/acme/settings
```

| Evidence | Type | Valid for | Collected | Source | Coordinate | Status |
|:---|:---|:---|:---|:---|:---|:---|
| EV-AC-01-01 | api-response | 1 day | 2h ago | GitHub API | /orgs/acme/settings | Current |
| EV-AC-01-02 | config-file | 1 day | 2h ago | Security Insights | SECURITY-INSIGHTS.yml:8 | Current |

### Posture reporting

Consumers can generate reports that combine both dimensions — freshness from the type definition, provenance from the mapping:

```
Requirement: OSPS-AC-01.01 (MFA Enforcement)
  Control review: not due (last reviewed 15d ago, frequency 90d)
  Evidence:
    - api-response — STALE (collected 36h ago, valid-for 1d)
        source: GitHub API → /orgs/acme/settings
    - config-file — STALE (collected 36h ago, valid-for 1d)
        source: Security Insights → SECURITY-INSIGHTS.yml:8
        integrity: sha256:e9f0a1b2c3...
    - quarterly-access-review — current (collected 15d ago, valid-for 90d)
  Alert: api-response and config-file overdue — re-scan recommended
```

## Consequences

### Positive

- **No changes to `#Evidence`.** ADR-0023's schema is unchanged. The `type` field already exists.
- **Two dimensions, typed separately.** `frequency` (plan-level) drives the audit schedule; `valid-for` (evidence-type-level) drives staleness alerts. Both are integers in days — computable, not prose.

### Negative

- Adds `#EvidenceTypeDefinition` to the schema vocabulary.
- Retypes `frequency` from `string` to `int` and replaces `evidence-requirements` on `#AssessmentPlan`, which is a breaking change to the experimental schema.

## Alternatives Considered

### Keep `frequency` and `evidence-requirements` as bare strings

Leave `#AssessmentPlan` as-is and let consumers interpret the fields.

**Rejected because:** Every consumer would independently invent freshness semantics. The bare strings carry no structure for tooling to act on.

