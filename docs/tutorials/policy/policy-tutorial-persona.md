---
layout: page
title: Organizational Risk & Policy Tutorial (Guinevere Persona)
description: Layer 3 policy document tutorial through the lens of Guinevere GRC, Security Governance Lead
---

This tutorial presents **Layer 3: Risk and Policy** through the lens of the best-fit persona for authoring or consuming organizational policy documents. Use it together with the [Organizational Risk & Policy Guide](policy-guide); here we frame the work as **Guinevere GRC (Security Governance Lead)** would.

## Persona: Guinevere GRC — Security Governance Lead

**Role:** Security Policy / GRC lead (e.g. at an international research institute). She bridges policy and engineering: she defines what “secure enough” means for dev environments, tools, and internal infra. She collaborates across compliance, security, and platform teams to keep things aligned and reduce risk without crushing velocity. She cares about policy-as-code, reusable baselines, and templates that teams actually use.

**Why she fits Layer 3:** Guinevere turns **organizational risk appetite** and **control selections** into a **policy document** that conforms to the Layer 3 schema. She is the **risk manager** who contextualizes **Pang’s Layer 2** threat and control catalogs into a **risk register** (business impact and likelihood), then defines which risks are **accepted** vs **mitigated** and sets **adherence** (evaluation methods, assessment plans) so Pang’s automation can validate controls in scope. A **Compliance** persona may assist with assessment plan and scope. **Governance leads** and **Grear** (for GDPR/HIPAA) ingest the policy to ensure adherence and how often assessments run. See [Tutorial user journeys](../persona-layer-alignment#tutorial-user-journeys-writing-vs-ingesting).

---

## As Guinevere: Your Goal

*As Guinevere, you’re turning organizational risk appetite and control selections into a policy document with scope, imports (guidance and catalogs), optional implementation-plan and risks, and adherence. You’re producing a **policy document** (Layer 3) that conforms to the current Layer 3 schema in `layer-3.cue`, so that evaluations (Layer 5) and audit (Layer 7) have a clear definition of what “compliant” means.*

---

## What You’ll Build

For the policy document structure and full schema overview (fields, imports, adherence), see [Organizational Risk & Policy Guide — What You'll Build](policy-guide#what-youll-build) in the base tutorial.

---

## Walkthrough (with Guinevere in mind)

1. **Use the base tutorial**  
   Follow the structure and steps in the [Organizational Risk & Policy Guide](policy-guide), which is outlined from the current **layer-3.cue** schema (metadata, contacts, scope, imports, implementation-plan, risks, adherence).

2. **Start from risk and controls**  
   Before writing policy, you need:
   - **Risk appetite** (what the organization is willing to accept) and a view of **risks** (informed by threats from Layer 2). In the schema these go under `risks.mitigated` and `risks.accepted` (with justification for accepted risks).
   - **Control catalogs** (from Pang’s world) that define *what* must be satisfied. Your policy references them via `imports.catalogs` and can tailor requirements with **assessment-requirement-modifications** (add, modify, remove, replace, override).
   - Optionally **guidance catalogs** (from Sachiko’s world) via `imports.guidance`, with optional exclusions and **constraints**.

3. **Define scope**  
   Set **scope.in** (and optionally **scope.out**) with dimension fields such as technologies, geopolitical, sensitivity, users, groups. Guinevere keeps scope clear so audits and evaluations are scoped and meaningful.

4. **Define imports and adherence**  
   Use **imports.guidance** and **imports.catalogs** to tie the policy to Layer 1 and Layer 2. Use **adherence** (evaluation-methods, assessment-plans, enforcement-methods, non-compliance) so Layer 5 and Layer 6 know how to evaluate and enforce.

5. **Validate**  
   Validate the policy document against the Gemara Layer 3 schema. From the repo root, run:

   ```bash
   cue vet -c -d '#Policy' ./layer-3.cue ./metadata.cue ./mapping.cue ./base.cue your-policy.yaml
   ```

---

## What’s Next

- **Layer 5** evaluations will test whether implementations conform to this policy (and thus to the controls and guidance it references).
- **Layer 7** audit and continuous monitoring will assess the effectiveness of this policy and of the evaluation and enforcement activities.

Upstream, your policy relies on:
- **Layer 2** control catalogs (and threat catalogs): [Threat Assessment Guide](../controls/threat-assessment-guide), [Control Catalog Guide](../controls/control-catalog-guide), and the [Threat Assessment & Control Catalog Tutorial (Pang)](../controls/threat-and-control-tutorial-persona).
- **Layer 1** guidance catalogs (optional): [Guidance Catalog Guide](../guidance/guidance-guide) and the [Guidance Catalog Tutorial (Sachiko)](../guidance/guidance-tutorial-persona).

See the [Persona–Layer Alignment](../persona-layer-alignment) for how Guinevere maps to Layer 3 and how the definition layers (1, 2, 3) connect.
