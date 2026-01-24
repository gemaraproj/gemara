---
layout: page
title: Dual-ladders Within Each Layer
---

- **ADR:** 0010
- **Proposal Author(s):** @eddie-knight
- **Status:** Accepted

## Context

We have been treating each layer as cumulative, with a single descriptor for the layer. In many cases, we have allowed for additional artificts with a single layer, such as a list of threats that lives alongside controls in a Layer 2 Control Catalog, or a list of Risks within a Policy Document.

Here is an overview of some common phrasing we've been using:

- Layer 1: Guidance (informed by known attack/negligence vectors)
- Layer 2: Controls (catalog includes technology-specific threats)
- Layer 3: Policy (document links to organizational risk considerations)
- Layer 4: Sensitive Activities
- Layer 5: Evaluation (config scans and behavior scans)
- Layer 6: Enforcement (gates and remediation)
- Layer 7: Audit (manual and continuous monitoring)

## Decision

We will allow for two primary artifacts within each layer. For current development conversations we will informally refer to these as "Red" and "Blue" until a better term is found, because these are _similar_ but not perfectly aligned with cybersecurity "Red Team" and "Blue Team" concepts.

The specific phrasing or terms may be refined or adjusted, but the meaning shall be roughly as follows:

| Layer | Red | Blue |
| 1 | Vector | Guidance |
| 2 | Threat | Control |
| 3 | Risk | Policy |
| 4 | Risk Actualization | Sensitive Activities |
| 5 | Behavioral Evaluation | Intent Evaluation |
| 6 | Remediative Enforcement | Preventive Enforcement |
| 7 | Continuous Compliance Monitoring | Point-in-Time Audit |

## Consequences

- All documentation and web content must be updated
- Adopters of previous versions will be disrupted by this change in terminology

## Alternatives Considered

We could not do this, but it leaves the relevant parts open to potential confusion.
