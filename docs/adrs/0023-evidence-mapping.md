---
layout: page
title: Evidence Mapping
---

- **ADR:** 0023
- **Proposal Author(s):** @jpower432
- **Status:** Draft
- **Supersedes:** ADR-0022

## Context

ADR-0022 introduced `#Evidence` on `#AssessmentLog` so that evaluation tools can record what they consulted during an assessment. 

The v1.3.0 implementation added the field and a reshaped `#Evidence` type, but shipped without the `address` and `digest` fields described in the accepted ADR. The type as released can only carry inline data via `payload`, which means tools cannot record *where* evidence was observed without embedding the content itself. During review, the question of whether `address` should be a plain string or a structured reference was left unresolved. 

## Action

A new mapping type specific to evidence references. It extends the `mapping-reference` system with fields for sub-origin, entry-level references, and integrity verification.

```cue
// EvidenceMapping references a source consulted during evidence collection.
// Properties describe the specific observation, not the artifact itself.
#EvidenceMapping: {
    // reference-id ties this evidence to a mapping-reference in the artifact's metadata
    "reference-id": string

    // coordinate identifies a specific origin within the referenced source
    // (line number, JSON path, API endpoint path, file fragment)
    // Used for opaque sources where Gemara cannot address internal structure.
    coordinate?: string

    // entry-id identifies a specific entry within a structured Gemara artifact
    // Used when the referenced source is a Gemara artifact with addressable entries.
    "entry-id"?: string @go(EntryId)

    // digest is a hash of the evidence content at collection time
    // Verifies that mutable-address evidence has not changed since the tool observed it.
    // Not needed for content-addressable systems (OCI, git) or inline payloads.
    digest?: string

    // remarks is prose regarding this evidence reference
    remarks?: string
}
```

`coordinate` and `entry-id` serve different layers but are not mutually exclusive. An audit-layer evidence entry may reference a specific assessment log entry (`entry-id`) within an evaluation log, while `coordinate` could identify a sub-origin within that entry's output. For opaque sources, only `coordinate` applies. For structured sources, `entry-id` may be sufficient.

`digest` is a mapping-level property because it captures the content hash *at collection time*. The same artifact referenced by two different evidence entries may have different digests if observed at different times.

Replace the v1.3.0 `#Evidence` definition with one that uses `#EvidenceMapping` for source references.

```cue
#Evidence: {
    // id uniquely identifies this evidence entry
    id: string

    // type categorizes the kind of evidence
    type: #EvidenceType

    // collected-at is the timestamp when the evidence was gathered
    "collected-at": #Datetime @go(CollectedAt)

    // payload is the raw evidence data collected inline
    payload?: _ @go(Payload,type=any)

    // origin references the source consulted during evidence collection
    origin?: #EvidenceMapping @go(Origin)

    // description explains what this evidence represents
    description?: string
}
```

`payload` and `origin` serve complementary roles:

- `payload` carries *what was observed* — the raw data, a snippet, or a summary.
- `origin` records *where it came from* — the source, the specific position within it, and its integrity at collection time.

Both may be present. A tool may record the specific line it observed (`payload`) alongside where that line lives (`origin`). Or a tool may provide only one: `payload` for inline-only evidence, `origin` for reference-only evidence.

The `evidence` field on `#AssessmentLog` in stable `evaluationlog.cue` remains `[#Evidence, ...#Evidence]`. No change to the stable schema's field type. When the same source is consulted by multiple steps or assessments, it appears once in `mapping-references` and is referenced by multiple evidence entries — each with its own `coordinate`, `digest`, and `payload` capturing that specific observation.

## Examples

### Evaluation layer — opaque source, two observations of the same file

```yaml
metadata:
  mapping-references:
    - id: "gomod"
      title: "Go module definition"
      version: "v0.5.0"
      url: "file://go.mod"

evidence:
  - id: "e-001"
    type: "config-file"
    collected-at: "2026-06-13T14:30:00Z"
    origin:
      reference-id: "gomod"
      coordinate: "go.mod:3"
      digest: "sha256:abc123..."
    payload: "go 1.23"
    description: "Go version declaration"

  - id: "e-002"
    type: "config-file"
    collected-at: "2026-06-13T14:30:02Z"
    origin:
      reference-id: "gomod"
      coordinate: "go.mod:12"
      digest: "sha256:abc123..."
    payload: "require github.com/sigstore/cosign/v2 v2.4.1"
    description: "Cosign dependency version"
```

### Evaluation layer — inline only, simple tool

```yaml
evidence:
  - id: "e-003"
    type: "api-response"
    collected-at: "2026-06-13T14:30:00Z"
    payload:
      mfa_enforced: true
      admin_accounts: 5
    description: "IAM MFA policy summary"
```

### Evaluation layer — external reference with integrity

```yaml
evidence:
  - id: "e-004"
    type: "config-file"
    collected-at: "2026-06-13T14:30:00Z"
    origin:
      reference-id: "k8s-manifest"
      coordinate: "/etc/kubernetes/manifests/kube-apiserver.yaml:42"
      digest: "sha256:d4e5f6..."
    description: "API server audit policy configuration"
```

### Audit layer — structured Gemara artifact reference

```yaml
metadata:
  mapping-references:
    - id: "eval-run-42"
      title: "Quarterly evaluation — Q2 2026"
      version: "1.0.0"
      url: "oci://registry.example.com/evaluations/run-42:v1.0.0"

evidence:
  - id: "e-005"
    type: "EvaluationLog"
    collected-at: "2026-06-13T15:00:00Z"
    origin:
      reference-id: "eval-run-42"
      entry-id: "assessment-mfa-check"
    description: "MFA enforcement assessment result from Q2 evaluation"
```

### Evaluation layer — multiple sources with different trust levels

```yaml
metadata:
  mapping-references:
    - id: "github-api"
      title: "GitHub REST API"
      version: "2026"
      url: "https://api.github.com"
    - id: "security-insights"
      title: "Security Insights YAML"
      version: "1.0.0"
      url: "file://SECURITY-INSIGHTS.yml"

evidence:
  - id: "e-007"
    type: "api-response"
    collected-at: "2026-06-13T14:30:00Z"
    origin:
      reference-id: "github-api"
      coordinate: "/repos/org/repo/vulnerability-alerts"
    payload:
      vulnerability_reporting_enabled: true
    description: "GitHub API reports vulnerability reporting is enabled"

  - id: "e-008"
    type: "config-file"
    collected-at: "2026-06-13T14:30:01Z"
    origin:
      reference-id: "security-insights"
      coordinate: "SECURITY-INSIGHTS.yml:12"
      digest: "sha256:e9f0a1..."
    payload:
      vulnerability-reporting:
        accepts-vulnerability-reports: false
    description: "Security Insights data contradicts GitHub API — conflicting evidence for vulnerability reporting status"
```

### Evaluation layer — coordinate with cached snippet

```yaml
evidence:
  - id: "e-006"
    type: "config-file"
    collected-at: "2026-06-13T14:30:00Z"
    origin:
      reference-id: "tls-config"
      coordinate: "/etc/nginx/nginx.conf:28-35"
      digest: "sha256:f7a8b9..."
    payload:
      ssl_protocols: "TLSv1.2 TLSv1.3"
      ssl_ciphers: "HIGH:!aNULL:!MD5"
    description: "TLS configuration snippet — full file at coordinate"
```

## Consequences

### Positive

- **No breaking changes.** All additions are optional fields on the experimental `#Evidence` type and a new experimental `#EvidenceMapping` type. The `evidence` field type on `#AssessmentLog` in stable `evaluationlog.cue` is unchanged.
- **Consistent with Gemara's mapping system.** Evidence references tie into `mapping-references` via `reference-id`, following the same pattern as every other Gemara cross-reference.
- **Mapping vs artifact separation.** Observation-level properties (`coordinate`, `digest`, `entry-id`) live on the mapping where they belong. Artifact-level properties (name, version, URL) stay in `mapping-references`. Two observations of the same source produce two evidence entries referencing the same mapping-reference.
- **Leverages existing infrastructure.** `mapping-references` provides discoverability for evidence sources. No new artifact type is needed.
- **Supports evidence trust differentiation.** When multiple sources provide conflicting evidence, `origin` enables consumers to trace each piece of evidence to its source and weigh findings based on source trust and collection method. Combined with the existing `confidence-level` on `#AssessmentLog`, this provides the full chain from raw observation to assessed confidence.

### Negative

`#EvidenceMapping` adds to the schema vocabulary. This is justified by evidence-specific concerns (integrity, sub-origin, temporal coupling) that do not exist on other mapping types.

## Alternatives Considered

### Plain `address` and `digest` fields on `#Evidence` (no mapping type)

Add `address` and `digest` directly to `#Evidence` without introducing `#EvidenceMapping`.

**Rejected because:** Evidence references have concerns that justify a dedicated mapping type: sub-origin within opaque sources, entry-level references for structured sources, integrity verification, and the principle that mapping-level properties should not be mixed with artifact-level properties. A flat structure also cannot distinguish between opaque (`address`) and structured (`entry-id`) reference patterns.

### Reuse `#ArtifactMapping` or `#EntryMapping`

Use an existing mapping type on `#Evidence` instead of creating a new one.

**Rejected because:** Existing Gemara mapping types (`#ArtifactMapping`, `#EntryMapping`, `#MultiEntryMapping`) do not fit evidence references for several reasons:

1. **Evidence references have unique concerns.** Integrity verification (`digest`), sub-origin within opaque sources (`coordinate`), and temporal coupling (`collected-at`) are specific to evidence collection. These concerns do not exist on other Gemara mappings.

2. **Opaque and structured sources require different fields.** Evaluation-layer evidence points to opaque sources (files, API responses, configs) where `coordinate` provides a sub-origin. Audit-layer evidence points to structured Gemara artifacts (EvaluationLogs, EnforcementLogs) where `entry-id` identifies a specific entry. No existing mapping type supports both patterns — `#ArtifactMapping` has no `entry-id`, and `#EntryMapping` requires it.

3. **Multiple observations of the same source.** Two assessment steps may consult the same artifact but observe different parts of it. The artifact is one entry in `mapping-references`; the observations are many. Properties like `coordinate`, `digest`, and the evidence payload describe the specific observation, not the artifact. These are mapping-level concerns, not artifact-level concerns.

4. **Evidence collection patterns vary.** Evidence may be collected "up from" steps (a step observes a specific slice of the payload and reports what it saw) or "just-in-time" at the assessment level (the assessment gathers evidence independently of step execution). In both cases, the evidence's relationship to its source — which part was observed, when, and with what result — is contextual to the collection event.

5. **Evidence from multiple sources with varying trust.** A single control may be assessed against evidence from different sources — API responses, configuration files, tool output, user-provided data — each with different levels of trust, quality, and specificity. Evidence from a synthetic push test (high confidence) is qualitatively different from evidence derived from API configuration data (medium confidence). The evidence type and its origin are essential for downstream consumers to weigh findings appropriately.
