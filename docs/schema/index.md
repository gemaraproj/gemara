---
layout: page
title: Gemara Schemas
nav-title: Schema
---

Schemas (CUE format) standardize the expression of elements in the model and enable automated interoperability between [GRC](../model/02-definitions.html#grc) tools. These schemas provide validation across all layers.

Click on a layer to view its schema: 

<div class="layer-grid">
  <a href="layer-1.html" class="layer-card">
    <h3>Layer 1: Guidance</h3>
    <p>High-level guidance on cybersecurity measures from industry groups and standards bodies.</p>
  </a>

  <a href="layer-2.html" class="layer-card">
    <h3>Layer 2: Controls</h3>
    <p>Technology-specific, threat-informed security controls for protecting information systems.</p>
  </a>

  <a href="layer-3.html" class="layer-card">
    <h3>Layer 3: Policy</h3>
    <p>Risk-informed guidance tailored to your organization's specific needs and risk appetite.</p>
  </a>

  <a href="layer-5.html" class="layer-card">
    <h3>Layer 5: Evaluation</h3>
    <p>Inspection of code, configurations, and deployments against policies and controls.</p>
  </a>

  <div class="layer-card">
    <h3>Layer 6: Enforcement</h3>
    <p>Prevention or remediation based on assessment findings. (Coming Soon)</p>
  </div>

  <div class="layer-card">
    <h3>Layer 7: Audit</h3>
    <p>Review of organizational policy and conformance. (Coming Soon)</p>
  </div>
</div>

**[Browse all schemas in the CUE Central Registry →](https://registry.cue.works/docs/github.com/gemaraproj/gemara@latest)**

## Schema Documentation

Schema documentation generated from CUE. One page per schema file:

<!-- SCHEMA_LIST_START -->

- [Aliases & Base Types](base.html)
- [Metadata](metadata.html)
- [Mapping](mapping.html)
- [Layer 1](layer-1.html)
- [Layer 2](layer-2.html)
- [Layer 3](layer-3.html)
- [Layer 5](layer-5.html)

<!-- SCHEMA_LIST_END -->

### Validation with CUE

The base package contains field-level type definitions and validation subpackage extends them with cross-field constraints.

```bash
# Install CUE
go install cuelang.org/go/cmd/cue@latest

# Validate with base types only
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@latest your-controls.yaml

# Validate with base types + cross-field constraints
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@latest:validation your-controls.yaml
```

#### Metadata Type Consistency

Each top-level type pins its expected `metadata.type` value.

| Artifact | Required `metadata.type` |
|:---|:---|
| GuidanceCatalog | `"GuidanceCatalog"` |
| ControlCatalog | `"ControlCatalog"` |
| ThreatCatalog | `"ThreatCatalog"` |
| VectorCatalog | `"VectorCatalog"` |
| MappingDocument | `"MappingDocument"` |
| EvaluationLog | `"EvaluationLog"` |
| Policy | `"Policy"` |

#### ID Uniqueness

Duplicate IDs within a collection are rejected.

| Type | Field |
|:---|:---|
| Metadata | `mapping-references[].id`, `applicability-categories[].id` |
| GuidanceCatalog | `guidelines[].id`, `families[].id`, `guidelines[].statements[].id` |
| VectorCatalog | `vectors[].id` |
| ControlCatalog | `controls[].id`, `families[].id`, `controls[].assessment-requirements[].id` |
| ThreatCatalog | `threats[].id`, `capabilities[].id` |
| MappingDocument | `mappings[].id` |
| Policy | `adherence.assessment-plans[].id` |

#### Mapping-Reference Integrity

All `reference-id` fields must point to a valid `metadata.mapping-references` entry.

| Type | Fields Validated |
|:---|:---|
| GuidanceCatalog | `guidelines[].principles`, `guidelines[].vectors`, `exemptions[].redirect` |
| ControlCatalog | `extends`, `controls[].guidelines`, `controls[].threats`, `imports.controls` |
| ThreatCatalog | `threats[].capabilities`, `threats[].vectors`, `extends`, `imports.threats`, `imports.capabilities` |
| MappingDocument | `source-reference`, `target-reference` |
| Policy | `imports.policies`, `imports.catalogs`, `imports.guidance`, `risks.mitigated` |

#### Lifecycle Constraints

| Type | Constraint |
|:---|:---|
| Guideline | Retired → no recommendations; Deprecated/Retired → `replaced-by` required |
| Control | Deprecated/Retired → `replaced-by` required |
| AssessmentRequirement | Retired → no recommendation; Deprecated/Retired → `replaced-by` required |

#### Structural Constraints

| Type | Constraint |
|:---|:---|
| GuidanceCatalog | Families required when guidelines exist |
| GuidanceCatalog | Guideline family IDs must reference existing families |
| GuidanceCatalog | Guideline extensions must match families |
| GuidanceCatalog | Guideline applicability must reference existing categories |
| GuidanceCatalog | Guideline see-also must reference existing guideline IDs |
| VectorCatalog | Vector applicability must reference existing categories |
| ControlCatalog | Families required when controls exist |
| ControlCatalog | Control family IDs must reference existing families |
| ControlCatalog | Assessment requirement applicability must reference existing categories |
| EvaluationLog | Assessment log requirement reference must match control reference |
| MappingDocument | mapping-references required (non-empty) |
| MappingDocument | Mapping applicability must reference existing categories |
| Mapping | Target required when relationship is not no-match |

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
| 5     | Measurement | Evaluation Result         | Evaluation Result       | Evaluation Log           | Evaluation Log         |
| 6     | Measurement | Enforcement Result        | Enforcement Result      | Enforcement Log          | Enforcement Log        |
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

Possible schema states include: **Experimental** → **Stable** → **Deprecated**. These are denoted on each layer schema with a `@status(experimental|stable|deprecated)` attribute.

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

* Layers promote independently. Each layer only requires its direct dependencies to be Stable (e.g., Layer 2 requires Layer 1, but not Layer 6).
* Layers can be promoted to Stable at different times. Layer 2 can be Stable while Layer 6 remains Experimental.
* Stable schemas may only reference other Stable schemas.
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
