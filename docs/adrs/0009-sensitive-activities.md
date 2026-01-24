---
layout: page
title: Formalize Sensitive Activities as a Layer
---

- **ADR:** 0009
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted

## Context

We have often recieved feedback regarding the lack of a "Layer" to describe sensitive activities that live between Policy and Evaluation.

- Layer 1: Guidance
- Layer 2: Controls
- Layer 3: Policy
- (No Layer) Sensitive Activities
- Layer 4: Evaluation
- Layer 5: Enforcement
- Layer 6: Audit

The argument has been that since we don't provide schemas for the diverse array of possible sensitive activities, we shouldn't assign a layer to it. However, as we move closer to a versioned release of the model, it has become clear that the model is not dependent on having schemas for every component.

## Decision

We will assign a layer to sensitive activities in the model.

## Consequences

- This will require +1 renumbering of layers 4, 5, and 6 (to 5, 6, and 7 respectve)
  - Layer 4: Sensitive Activities
  - Layer 5: Evaluation
  - Layer 6: Enforcement
  - Layer 7: Audit
- No impact on schemas and SDKs, as they have already removed any reference to numberic layer identifiers
- All documentation and web content must be updated
- Adopters of previous versions will be disrupted by this change in terminology

## Alternatives Considered

We have previously ruled this out due to the intended stability of the model. However, with the upcoming release of the whitepaper, change of repository locations, and major (v1) changes to the schemas and SDKs, this is an opportunity for a breaking change to the model which will be included in the official whitepaper for long term stability.
