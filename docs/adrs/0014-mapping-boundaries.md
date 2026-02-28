---
layout: page
title: Mapping Boundaries
---

- **ADR:** 0014
- **Proposal Author(s):** @jpower
- **Status:** Accepted

## Context

The schema supports inline cross-artifact mappings (e.g., `external-mappings`) on entry types alongside self-referential fields (e.g., `see-also`).
Inline mapping fields intended for documentation causes unbounded growth of the artifact resulting in a less than ideal artifact editing experience.
Some inline fields serve as design justifications rather than cross-framework mappings, and should remain on the entry.

## Decision

We will establish two distinct mechanisms for expressing relationships, separated by ownership and scope:

1. Catalog owners declare design rationale directly on entries. These fields are renamed from `*-mappings` to their target nouns (`vectors`, `guidelines`, `threats`); the parent type disambiguates them from definition lists on catalogs.
Technology-specific relationships (e.g., control-to-threat, threat-to-capability) are always inline.
2. `#MappingDocument` is a dedicated artifact for users to express alignment between independently-authored artifacts in a specific direction.
Each mapping captures the author's assertion of relationship and strength.

## Consequences

- Existing inline `*-mappings` fields are removed or renamed to target nouns; this is a breaking change
- Schema documentation must distinguish owner-authored (inline) from non-owner (MappingDocument)
- MappingArtifacts describing mapping quality so strength is removed from inline mappings

## Alternatives Considered

### Keep all mappings inline

Retains current schema shape but does not scale for catalogs with many cross-framework mappings.

### Move all mappings to MappingDocument

Catalogs become pure definitions, but a control without its threats loses its design rationale.
