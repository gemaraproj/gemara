---
layout: page
title: Schema Organization by Artifact Type
---

- **ADR:** 0015
- **Proposal Author(s):** @jpower432, @eddie-knight
- **Status:** Accepted
- **Modifies:** [ADR 0006](./0006-unified-package-structure.md)

## Context

Until this point, we have organized the schemas based on layer. For example, `layer-3.cue` contains the schemas for Capabilities, Threats, and Controls.

This layer-based organization is not working as well anymore because (1) the layers have multiple main types and (2) the artifacts are becoming the primary unit of work for users.

## Decision

Reorganize CUE schema files by artifact type instead of by layer.

Each top-level artifact type will be organized into its own file which contains the artifact type definition and its main component types.

These artifacts are represented in `#ArtifactType`

### Website Organization

The website is set to render schemas based on the file they live in, so there will now be one page per artifact type.

## Consequences

- Existing references to layer-based files need updating (documentation, scripts, etc.)
- Schema documentation generation and navigation need updates

### Neutral

1. **Website Presentation**: Layer-based cards and navigation remain unchanged for users
2. **CUE Validation**: No impact on CUE validation or schema correctness
3. **Go SDK**: No impact on Go SDK generation (types are generated from schemas regardless of file organization

## Alternatives Considered

Continue organizing schemas by layer as originally planned, but at this point the benefits of artifact-type organization outweigh the migration effort

## References

- [ADR 0006: Unified Go SDK Package Structure](./0006-unified-package-structure.md)
- [ADR 0012: Refine Terms for Top-level Schema Types](./0012-schema-types.md)
