---
layout: page
title: Extend Go SDK for "Layer 4" Based on the Privateer Project
---

- **ADR:** 0004
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted; Modified by [ADR-0009](./0009-sensitive-activities) (numbering for layers 3+)

## Context

_FINOS Common Cloud Controls (CCC)_ community maintains a custom tool, [Privateer](https://privateerproj.com), which uses our SDK to ingest CCC documents for the automatic generation of plugins designed to assess that "Layer 2" catalog's assessment requirements.

The plugin generates an output which is designed to streamline the organization and presentation of evidence following an assessment. The Privateer schema has already served as the foundation for the "Layer 4" schema. We may be able to extract much of the Privateer logic into a shared SDK that can be used by Privateer or other tools seeking to be compatible with our schemas.

## Action

Identify and extract the relevant capabilities from Privateer into a new package within our Go SDK. Support Privateer in migrating to use the new SDK instead of its current internal logic.

## Consequences

Positive: Standardized tooling for "Layer 4" compatible documents
Negative: Significantly increased maintenance requirements for the project

## Alternatives Considered

We could write a net-new independent SDK, or none at all.
