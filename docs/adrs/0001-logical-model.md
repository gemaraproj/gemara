---
layout: page
title: Create A Logical Model for GRC Engineering Activities
---

- **ADR:** 0001
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted; Modified by [ADR-0009](./0009-sensitive-activities.md) (layer numbers)

## Context

During the writing of the CNCF's [Automated Governance Maturity Model (AGMM)](https://www.cncf.io/blog/2025/05/05/announcing-the-automated-governance-maturity-model/), the authors of that document came to a key observation: to build a _Secure Software Factory_ or a fully automated governance program, maturity can be measured in at least four different areas: "Policy, Evaluation, Enforcement, and Audit."

Both the FINOS _Common Cloud Controls (CCC)_ and OpenSSF's _Open Source Project Security Baseline (OSPSB)_ projects began using the language from the AGMM, with the addition of two more areas to describe the difference between documents that give high-level recommendations alongside those which create technology-specific objectives. It was observed that these six different areas build upon eachother, similar to how the OSI Model describes the layers of digital networking.

## Action

Create a community-defined model in a version-controlled project that can be extended with supplemental resources such as schemas and SDKs.

## Consequences

Positive: Can be updated and extended over time
Negative: Requires up-front commitment from a core maintainer group

## Alternatives Considered

An alternative would be to write a white paper following on the AGMM, but this would rule out the possibility of community maintenance and extension with supplemental resources for automation acceleration.
