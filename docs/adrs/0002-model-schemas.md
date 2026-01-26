---
layout: page
title: Create Schemas for Each Layer in the Logical Model
---

- **ADR:** 0002
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted; Modified by [ADR-0009](./0009-sensitive-activities.md) (numbering for layers 3+)

## Context

A discussion emerged from pockets of the OSCAL user community at the same time as the logical model was being formed, exploring complexities regarding automatic generation and consumption of OSCAL documents at scale.

To partially address this in their own space, _FINOS Common Cloud Controls (CCC)_ community developed an intermediary automation layer which would use a custom schema for writing documents and a custom CI tool to convert to OSCAL on release. Very soon after, the tooling from the CCC community was imitated by the OpenSSF's _Open Source Project Security Baseline (OSPSB)_ project.

## Action

Create an initial set of schemas using CUE to describe a common structure for "Layer 2" documents such as CCC and OSPSB and "Layer 4" results from automated evaluations. The schemas should be 100% OSCAL-compatible while optimizing for automation. This can be extended over time to cover all of the different layers.

## Consequences

Positive: Standardized expression of the structure for similar documents
Negative: Significantly increased maintenance requirements for the project

## Alternatives Considered

JSONSchema is a viable alternative for the expression of the schemas. If needed in the future, both CUE and JSONSchema could theoretically be maintained in tandem.
