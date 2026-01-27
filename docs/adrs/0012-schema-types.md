---
layout: page
title: Refine Terms for Top-level Schema Types
---

- **ADR:** 0012
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted

## Context

We have been organizing schemas based on the layer they fit into — one document per layer. The terms we use for the schema types have been evolving over time and do not successfully achieve consistency and clarity.

Additionally, the model changes from [ADR-0010](./0010-dual-ladder-layers) have not all been applied to the latest schemas.

Below are the current top-level types at the time of this proposal:

- Layer 1: `#GuidanceDocument`
- Layer 2: `#Catalog`
- Layer 3: `#Policy`
- Layer 4: Not applicable
- Layer 5: `#EvaluationLog`
- Layer 6: Not yet implemented
- Layer 7: Not yet implemented

## Decision

1. Any result which is released as a single self-contained output will be referred to as a "Document"
2. Multiple definitions compiled into a shared artifact will be referred to as a "Catalog"
3. Multiple measurements compiled into a shared artifact will be referred to as a "Log"

The specific artifacts may evolve over time, but the top-level schema terms should be roughly:

- Layer 1: Vector Catalog, Guidance Catalog
- Layer 2: Threat Catalog, Control Catalog
- Layer 3: Risk Catalog, Policy Document
- Layer 4: n/a
- Layer 5: Evaluation Log
- Layer 6: Enforcement Log
- Layer 7: Monitoring Log, Audit Log

## Consequences

- All schemas must be updated
- All documentation and web content must be updated
- Adopters of previous versions will be disrupted by these breaking schema changes; this must be locked-in with release of v1

## Alternatives Considered

We could avoid naming conventions, but it leaves the relevant parts open to potential confusion.
