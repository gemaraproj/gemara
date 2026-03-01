---
layout: page
title: Threat Assessment & Control Catalog Tutorial (Pang Persona)
description: Layer 2 threat and control catalog tutorials through the lens of Pang the Product Security Engineer
---

This tutorial presents **Layer 2: Threats and Controls** through the lens of the best-fit persona for authoring or adopting threat catalogs and control catalogs. Follow the same steps as in the [Threat Assessment Guide](threat-assessment-guide) and the [Control Catalog Guide](control-catalog-guide); here we frame them as **Pang the Product Security Engineer** would.

## Persona: Pang the Product Security Engineer

**Role:** Product Security Engineer (security champion) embedded in engineering. He bakes security into backlogs, uses standards like SLSA and SAMM, and integrates tooling with CI/CD. He’s familiar with secure SDLC, security standards, and controls; he sees himself as the bridge between security policy and practical engineering.

**Why he fits Layer 2:** Pang defines **controls with assessment requirements** so teams can be evaluated against them. He works with **threat-informed** controls—capabilities and threats first (threat catalog), then safeguards with verifiable assessment requirements (control catalog). He uses the **[FINOS CCC](https://www.finos.org/common-cloud-controls-project) Core** reusable catalog whenever possible to import capabilities, threats, or controls. The output is exactly what Layer 5 evaluation tooling (e.g. Privateer) consumes and what **Guinevere** (Layer 3) and **Carl the Consumer** (vendor/manufacturer from the [OSSF Toolbelt personas](https://github.com/ossf/toolbelt/tree/main/personas) ingest: Carl must understand the **security posture** of the OSS his project consumes and wants the software he's building to **comply with regulatory requirements** and **manage risks associated with consumption**. See [Tutorial user journeys](../persona-layer-alignment#tutorial-user-journeys-writing-vs-ingesting).

---

## As Pang: Your Goal

*As Pang, you’re defining controls with assessment requirements so teams can be evaluated against them. You start from “what can go wrong” (threat catalog) and then define “what we require” (control catalog) with clear, checkable conditions. Your catalogs will be used in evaluations and in policy (Layer 3).*

---

## Walkthrough (with Pang in mind)

### Part 1: Threat catalog

Use the full steps in the [Threat Assessment Guide](threat-assessment-guide). While you do:

1. **Step 0 – Scope**  
   Pick the component or technology you’re securing (e.g. container management tool, SEC.SLAM.CM). Same scope will be reused for the control catalog.

2. **Step 1 – Metadata and mapping-references**  
   Point to external catalogs you trust (e.g. FINOS CCC Core) so you can **import** capabilities and threats instead of redefining everything. Pang prefers reusing well-vetted threats and only adding what’s specific to his scope.

3. **Step 2 – Capabilities**  
   Identify what the system *does* (attack surface). Import from CCC where it fits (e.g. Active Ingestion, Resource Versioning); add scope-specific capabilities (e.g. “Image Retrieval by Tag”).

4. **Step 3 – Threats**  
   For each capability, ask “what could go wrong?” Import common threats (e.g. CCC.Core.TH14) and define threats specific to your scope (e.g. image tampering or poisoning). These threat IDs will be referenced in **threat-mappings** in your control catalog.

5. **Validate**  
   Validate the threat catalog against the Layer 2 schema as in the main guide.

### Part 2: Control catalog

Use the full steps in the [Control Catalog Guide](control-catalog-guide). While you do:

1. **Step 0 – Scope and references**  
   Reuse the same scope and threat catalog ID. Add a mapping reference to your threat catalog and, if you import controls, to CCC or another control catalog.

2. **Step 1 – Metadata**  
   Declare the control catalog and its mapping-references (threat catalog + optional CCC).

3. **Step 2 – Families**  
   Group controls by theme (e.g. “Image Integrity and Supply Chain”). These are the same buckets you’d use in backlogs and security reviews.

4. **Step 3 – Import or define controls**  
   **Import** controls from CCC (or elsewhere) when they address your threats; **define** controls when they’re specific to your scope. Each control needs an **objective** and **assessment-requirements**—verifiable conditions that an evaluator (or Layer 5 tool) can check. Pang cares that every requirement is testable (e.g. “MUST resolve to digest,” “MUST disallow tag-only for production”).

5. **Step 4 – Threat-mappings**  
   Link each control to the threat(s) it mitigates via `threat-mappings`. This documents *why* each control exists and ties Layer 2 to the threat catalog.

6. **Step 5 – Validate**  
   Validate the control catalog against the Layer 2 schema. Fix any schema errors so the catalog is ready for Layer 5 evaluations and for Guinevere (Layer 3) to reference in policy.

---

## Summary: From threats to controls (Pang’s view)

| You did (threat assessment) | You use (control catalog) |
|-----------------------------|----------------------------|
| Scope / catalog id (e.g. SEC.SLAM.CM) | Same id in control catalog metadata; reference in mapping-references |
| Threat IDs (e.g. SEC.SLAM.CM.THR01, CCC.Core.TH14) | `threat-mappings` on each control that mitigates those threats |
| Understanding of what goes wrong | Control objectives and **assessment requirements** that evaluators can verify |

---

## What’s Next

- Use the control catalog in **Layer 5 evaluations** (e.g. Privateer) to record which assessment requirements passed or failed.
- Let **Guinevere** (Layer 3) reference your controls in **policy documents**: [Organizational Risk & Policy Guide](../policy/policy-guide) and the [Organizational Risk & Policy Tutorial (Guinevere)](../policy/policy-tutorial-persona).
- Optionally add **guideline-mappings** on controls to link back to **Layer 1 guidance** (Sachiko’s catalogs): [Guidance Catalog Guide](../guidance/guidance-guide) and [Guidance Catalog Tutorial (Sachiko)](../guidance/guidance-tutorial-persona).

See the [Persona–Layer Alignment](../persona-layer-alignment) for how Pang maps to Layer 2.
