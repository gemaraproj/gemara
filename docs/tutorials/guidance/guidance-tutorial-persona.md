---
layout: page
title: Guidance Catalog Tutorial (Sachiko Persona)
description: Layer 1 guidance catalog tutorial through the lens of Sachiko the Solution Architect
---

This tutorial presents **Layer 1: Vectors and Guidance** through the lens of the best-fit persona for authoring or adopting guidance catalogs. Follow the same steps as in the [Guidance Catalog Guide](guidance-guide); here we frame them as **Sachiko the Solution Architect** would.

## Persona: Sachiko the Solution Architect

**Role:** Principal Software Architect at a cloud infrastructure company. She defines secure design patterns and blueprints, aligns with frameworks like SLSA and SAMM, and shares internal guidance so others can reuse it. She doesn’t write most of the code anymore but is the person teams go to when something needs to scale securely.

**Why she fits Layer 1:** Sachiko captures **secure-by-design guidelines** and maps them to external standards (e.g. OWASP, NIST) so the organization can reuse one coherent guidance catalog internally. She focuses on **optimal outcomes without foreknowledge of implementation details**—exactly what a Gemara guidance catalog is for.

> **Alternative personas for Layer 1:** For **regulator- or governing-body-aligned** guidance (e.g. data protection, regulated industry), use **Grear** (Data Governance Analyst in a government agency) or **Guinevere** (GRC lead developing guidance for globally distributed teams). See [Tutorial user journeys](../persona-layer-alignment#tutorial-user-journeys-writing-vs-ingesting) and [Persona–Layer Alignment](../persona-layer-alignment#layer-1-vectors-and-guidance).

---

## As Sachiko: Your Goal

*As Sachiko, you’re capturing secure-by-design guidelines and mapping them to OWASP/NIST (or other standards) so teams can reuse this guidance internally. You’re producing a **guidance catalog**—families and guidelines with optional `guideline-mappings`—that downstream controls (Layer 2) and policy (Layer 3) can reference.*

---

## Walkthrough (with Sachiko in mind)

Follow the full step-by-step flow in the [Guidance Catalog Guide](guidance-guide) (Step 0: Scope and catalog type → Step 1: Metadata → Step 2: Families → Step 3: Guidelines → Step 4: Validate). As Sachiko, emphasize: scope and type that match how your org thinks about systems; `mapping-references` to the standards you care about (OWASP, NIST, SLSA, etc.); families that align with design reviews and architecture; and **`guideline-mappings`** on each guideline so Pang (Layer 2) and Guinevere (Layer 3) can trace lineage to standards.

---

## What’s Next

- Map your guidance to **Layer 2 controls** via control catalogs’ `guideline-mappings` (Pang’s world): [Threat Assessment Guide](../controls/threat-assessment-guide) and [Control Catalog Guide](../controls/control-catalog-guide).
- Reference this catalog from **Layer 3 policy** (Guinevere’s world): [Organizational Risk & Policy Guide](../policy/policy-guide) and the [Organizational Risk & Policy Tutorial (Guinevere)](../policy/policy-tutorial-persona).

See the [Persona–Layer Alignment](../persona-layer-alignment) for how Sachiko, Grear, and Regulated Ragnar map to Layer 1.
