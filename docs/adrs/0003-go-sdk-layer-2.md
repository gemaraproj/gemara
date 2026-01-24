---
layout: page
title: Create Go SDK to Support "Layer 2" Authors and Consumers
---

- **ADR:** 0003
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted

## Context

As noted in ADR-0002, much automation work has been done by the _FINOS Common Cloud Controls (CCC)_ and _Open Source Project Security Baseline (OSPSB)_ communities to automate document handling and OSCAL generation. These two projects, and potentially others in the future, will greatly benefit from a central SDK that supports common activities.

## Action

Take the best elements from each project's CI tooling and bring them together into a single SDK. Support those two projects as they migrate their existing tooling to the new SDK. This will be a Go module with a package dedicated to Layer 2 documents. It should be extensible in a way that allows additional packages for other documents in the future.

## Consequences

Positive: Standardized tooling for "Layer 2" compatible documents
Negative: Significantly increased maintenance requirements for the project

## Alternatives Considered

None
