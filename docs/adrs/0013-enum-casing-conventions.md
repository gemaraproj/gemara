---
layout: page
title: Enum Casing Conventions
---

- **ADR:** 0013
- **Proposal Author(s):** @jpower432
- **Status:** Accepted

## Context

Enum values across the schema used inconsistent casing: some Title Case, some lowercase, some kebab-lower.
No documented convention existed to guide contributors.

## Decision

Title Case is the convention for all enum values.
Exceptions need to be justified, but as a rule casing convention should optimize for the author, not the potential export format.

Title Case provides visual distinction from kebab-lower field names, making artifacts easier to scan. Multi-word values use spaces, not hyphens (e.g. `"Not Applicable"`, not `"Not-Applicable"`).

## Consequences

- Breaking change for consumers using lowercase `#Lifecycle`, `#MethodType`, and `#ModType` values
- Breaking change for consumers referencing `"Software-Assisted"` in `#ActorType`
- Breaking change for consumers referencing `"Not Set"` in `#ConfidenceLevel`

## Alternatives Considered

- **kebab-lower:** Aligns with OSCAL, STIX, and CycloneDX. However, field names are already kebab-lower, so values become visually indistinct from keys in authored YAML.
- **camelCase:** Highest cognitive load for authors. Uncommon in YAML-heavy ecosystems.
- **PascalCase:** Removes space/hyphen ambiguity, but multi-word values lose readability (`"NotApplicable"`). Optimizes for developers, not YAML authors.
