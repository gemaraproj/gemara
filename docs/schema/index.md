---
layout: page
title: Gemara Schemas
nav-title: Schema
---

Schemas (CUE format) standardize the expression of elements in the model and enable automated interoperability between GRC tools. These schemas provide validation across all layers.

Click on an artifact to view its schema:

<div class="layer-grid">
  <a href="guidancecatalog.html" class="layer-card">
    <span class="layer-badge">Layer 1</span>
    <h3>Guidance Catalog</h3>
    <p>High-level guidance on cybersecurity measures from industry groups and standards bodies.</p>
  </a>

  <a href="vectorcatalog.html" class="layer-card">
    <span class="layer-badge">Layer 1</span>
    <h3>Vector Catalog</h3>
    <p>Attack vectors and techniques used to compromise information systems.</p>
  </a>

  <a href="principlecatalog.html" class="layer-card">
    <span class="layer-badge">Layer 1</span>
    <h3>Principle Catalog</h3>
    <p>Foundational values that guide governance, design, and operational decisions.</p>
  </a>

  <a href="controlcatalog.html" class="layer-card">
    <span class="layer-badge">Layer 2</span>
    <h3>Control Catalog</h3>
    <p>Technology-specific, threat-informed security controls for protecting information systems.</p>
  </a>

  <a href="capabilitycatalog.html" class="layer-card">
    <span class="layer-badge">Layer 2</span>
    <h3>Capability Catalog</h3>
    <p>Technology capabilities that can be leveraged to implement security controls.</p>
  </a>

  <a href="threatcatalog.html" class="layer-card">
    <span class="layer-badge">Layer 2</span>
    <h3>Threat Catalog</h3>
    <p>Threats mapped to technology capabilities and security controls.</p>
  </a>

  <a href="riskcatalog.html" class="layer-card">
    <span class="layer-badge">Layer 3</span>
    <h3>Risk Catalog</h3>
    <p>Organizational risk categories, severity levels, and risk appetite definitions.</p>
  </a>

  <a href="policy.html" class="layer-card">
    <span class="layer-badge">Layer 3</span>
    <h3>Policy</h3>
    <p>Risk-informed guidance tailored to your organization's specific needs and risk appetite.</p>
  </a>

  <a href="evaluationlog.html" class="layer-card">
    <span class="layer-badge">Layer 5</span>
    <h3>Evaluation Log</h3>
    <p>Inspection of code, configurations, and deployments against policies and controls.</p>
  </a>

  <a href="enforcementlog.html" class="layer-card">
    <span class="layer-badge">Layer 6</span>
    <h3>Enforcement Log</h3>
    <p>Prevention or remediation based on assessment findings.</p>
  </a>

  <a href="auditlog.html" class="layer-card">
    <span class="layer-badge">Layer 7</span>
    <h3>Audit Log</h3>
    <p>Review of organizational policy and conformance.</p>
  </a>
</div>

**[Browse all schemas in the CUE Central Registry →](https://registry.cue.works/docs/github.com/gemaraproj/gemara@latest)**

## Schema Documentation

Schema documentation generated from CUE. One page per schema file:

<!-- SCHEMA_LIST_START -->

- [Aliases & Base Types](base.html)
- [Metadata](metadata.html)
- [Mapping Primitives](mapping-inline.html)
- [Guidance Catalog](guidancecatalog.html)
- [Vector Catalog](vectorcatalog.html)
- [Principle Catalog](principlecatalog.html)
- [Control Catalog](controlcatalog.html)
- [Capability Catalog](capabilitycatalog.html)
- [Threat Catalog](threatcatalog.html)
- [Risk Catalog](riskcatalog.html)
- [Policy](policy.html)
- [Mapping Document](mappingdocument.html)
- [Lexicon](lexicon.html)
- [Evaluation Log](evaluationlog.html)
- [Enforcement Log](enforcementlog.html)
- [Audit Log](auditlog.html)

<!-- SCHEMA_LIST_END -->

### Validation with CUE

```bash
# Install CUE
go install cuelang.org/go/cmd/cue@latest

# Validate a control catalog using the ControlCatalog definition
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@latest your-controls.yaml
```

## Architecture Decisions

Significant implementation changes are documented in [Architecture Decision Records (ADRs)](../adrs/index.html).

## Schema Standards

Schemas follow these naming standards:

| Layer | Category    | Atomic Unit (Interactive) | Atomic Unit (Defensive) | Collection (Interactive) | Collection (Defensive) |
|:------|:------------|:--------------------------|:------------------------|:-------------------------|:-----------------------|
| 1     | Definition  | Vector                    | Guidance                | Vector Catalog           | Guidance Catalog       |
| 2     | Definition  | Threat                    | Control                 | Threat Catalog           | Control Catalog        |
| 3     | Definition  | Risk                      | Policy                  | Risk Catalog             | —                      |
| 4     | —           | —                         | —                       | —                        | —                      |
| 5     | Measurement | Control Evaluation        | Control Evaluation      | Evaluation Log           | Evaluation Log         |
| 6     | Measurement | Action Result             | Action Result           | Enforcement Log          | Enforcement Log        |
| 7     | Measurement | —                         | Audit Result            | —                        | Audit Log              |

**Note:** "—" indicates no planned schema implementation.

**Naming conventions**:
- **Definition layers (1-3)**: Collections of atomic units are called **Catalogs**
- **Measurement layers (5-7)**: Collections of atomic units are called **Logs**

## Versioning and Maintenance

Project release deliverables are divided into the **Core Specification** and language-specific **SDKs**.

* **Core Specification Release:** The CUE schemas are versioned and released as the core specification. These schemas implement the Model, which is published separately.
* **SDK Releases:** Language-specific implementations that provide tooling, types, and helpers to work with Gemara documents. SDK types are generated from the CUE schemas.

Each **maintains its own** independent [SemVer](https://semver.org/) lifecycle.

### Specification Release Versioning

The core specification release versions the CUE schemas as a single unit.

| Change Type | Version Bump | Examples                                                     |
|:------------|:-------------|:-------------------------------------------------------------|
| Major       | v2.0.0       | Breaking changes to Stable schemas.                         |
| Minor       | v1.(x+1).0   | Additive changes, schema promotions, or new optional fields. |
| Patch       | v1.x.(y+1)   | Bug fixes in schema logic or documentation.                  |

### Schema Lifecycle and Major Version Example

Possible schema states include: **Experimental** → **Stable** → **Deprecated**. These are denoted on each schema file with a `@status(experimental|stable|deprecated)` attribute.

The following table illustrates how schemas progress through their lifecycle and how major version changes are handled:

| Version | Status              | Change Type | Example Scenario                        |
|:--------|:--------------------|:------------|:----------------------------------------|
| v1.0.0  | Experimental        | Initial     | Schema first published                  |
| v1.1.0  | Experimental        | Minor       | Optional fields added                   |
| v1.2.0  | Stable              | Minor       | Promoted to Stable                      |
| v1.3.0  | Stable              | Minor       | Additive changes                        |
| v1.3.1  | Stable              | Patch       | Bug fix                                 |
| v1.4.0  | Stable              | Minor       | Field deprecated, replacement added     |
| v1.5.0  | Stable → Deprecated | Minor       | Original deprecated, replacement Stable |
| v2.0.0  | Stable              | Major       | Deprecated schema removed               |
| v2.1.0  | Stable              | Minor       | Additive changes                        |

### Experimental Status
* All new schemas start as Experimental.
* Adding Experimental schemas or making breaking changes triggers minor version increments.
* Breaking changes and performance issues may occur during this phase.
* These schemas may not be feature-complete.

### Stable Status

* Schemas promote independently and at different times (e.g., ControlCatalog can be Stable while Policy remains Experimental).
* When a schema is promoted to Stable, all types it references must also be Stable (e.g., promoting `#ControlCatalog` requires `#Catalog`, `#Metadata`, `#Group`, and any other referenced types to already be Stable).
* Stable schemas may only reference other Stable types.
* Stable schemas maintain backward and forward compatibility within major versions, allowing **additive optional changes** only.
* Breaking changes require major version increments and should be avoided in all normal circumstances.

### Deprecated Status

* Schemas or fields within schemas may be deprecated when replaced.
* Replacement schema or field must be added in Experimental status and promoted to Stable before deprecation.
* Deprecated schemas and fields maintain the same support guarantees as Stable schemas and remain functional.
* Deprecated schemas are removed in the next major version release.

## Contributing

The Schemas evolve based on community needs. If you find an error or have suggestions for a schema improvement, open a [GitHub Issue](https://github.com/gemaraproj/gemara/issues) or raise your suggestion on a [community](/community) meeting.

See the [Contributing Guide](https://github.com/gemaraproj/gemara/blob/main/CONTRIBUTING.md) for details.

## Relationship to Other Components

### [The Model](../model)
The Model is published separately and provides the conceptual foundation. These CUE schemas are an implementation of the model, with each schema corresponding to a layer in the model.

### [The SDKs](../sdk/)
SDKs support programmatic access to Gemara documents. SDK types are generated from these schemas, ensuring consistency between validation and programmatic access.

## Questions or Feedback

For questions about versioning strategy or to propose changes:

* Open an issue on [GitHub](https://github.com/gemaraproj/gemara/issues)
* Discuss in the [OpenSSF Slack #gemara channel](https://openssf.slack.com/archives/C09A9PP765Q)
* Attend the [biweekly Gemara meeting](https://calendar.google.com/calendar/u/0?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ)
