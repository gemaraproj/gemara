---
layout: page
title: Tutorials
---

## Find Your Tutorial

Pick your role, then your goal — each path leads to the right guide.

### I am a Product Security Engineer

- **I want to define engineering standards** for my organization
  → [Guidance Catalog Guide](guidance/guidance-guide) — create a guidance catalog (Layer 1) with guidelines, families, and mapping-references to standards like OWASP, NIST, and SLSA.
- **I want to perform a threat assessment** for a system or component
  → [Threat Assessment Guide](controls/threat-assessment-guide) — identify capabilities and threats, map them to attack surfaces (Layer 2).
- **I want to define security controls** that mitigate those threats
  → [Control Catalog Guide](controls/control-catalog-guide) — create a control catalog with assessment requirements and threat-mappings (Layer 2).

### I am a GRC / Security Governance Lead

- **I want to create organizational policy** that captures risk appetite and compliance
  → [Organizational Risk & Policy Guide](policy/policy-guide) — create a policy document with scope, imports, adherence, and risks (Layer 3).
- **I want to define guidance for distributed teams** to follow
  → [Guidance Catalog Guide](guidance/guidance-guide) — create reusable guidance catalogs that feed into policy and controls (Layer 1).
- **I want to understand what threats and controls exist** before writing policy
  → [Threat Assessment Guide](controls/threat-assessment-guide) and [Control Catalog Guide](controls/control-catalog-guide) — review or author threat-informed controls that your policy will reference (Layer 2).

### I am a Solution Architect

- **I want to create secure-by-design guidance** for my organization or domain
  → [Guidance Catalog Guide](guidance/guidance-guide) — share reusable blueprints aligned with SLSA, SAMM, and other frameworks (Layer 1).
- **I want to see how guidance feeds into controls and policy**
  → [Persona–Layer Alignment](persona-layer-alignment) — understand the full flow from guidance to controls to policy.

### I am a Risk Manager / Compliance Analyst

- **I want to define policy and adherence requirements**
  → [Organizational Risk & Policy Guide](policy/policy-guide) — create policy with assessment plans, enforcement methods, and risk acceptance (Layer 3).
- **I want to review the controls my policy will reference**
  → [Control Catalog Guide](controls/control-catalog-guide) — understand the control catalog structure and assessment requirements (Layer 2).

### I am a Vendor / Manufacturer consuming OSS

- **I want to understand the security posture** of software I consume
  → [Threat Assessment Guide](controls/threat-assessment-guide) — review threat catalogs for your dependencies (Layer 2).
- **I want to comply with regulatory requirements** and manage consumption risk
  → [Control Catalog Guide](controls/control-catalog-guide) — use control catalogs (e.g. OSPS, CCC) to assess and communicate posture (Layer 2).

---

## What You'll Build

| Layer | Artifact | Guide |
|-------|----------|-------|
| **Layer 1** — Guidance | Guidance Catalog (guidelines, families, mapping-references) | [Guidance Catalog Guide](guidance/guidance-guide) |
| **Layer 2** — Controls | Threat Catalog + Control Catalog (assessment requirements, threat-mappings) | [Threat Assessment](controls/threat-assessment-guide), [Control Catalog](controls/control-catalog-guide) |
| **Layer 3** — Policy | Policy Document (scope, imports, adherence, risks) | [Policy Guide](policy/policy-guide) |

For who *writes* vs. who *ingests* each layer, see [Persona–Layer Alignment](persona-layer-alignment).

## What You'll Need

* `go` installed
* `cue` installed for validation
* (Optional) [OpenSSF Toolbelt personas](https://github.com/ossf/toolbelt/tree/main/personas) for persona-based framing

## Have Ideas? 

* Reach out via Slack in `#gemara`
* Discuss in one of our bi-weekly meetings on the [OpenSSF calendar](https://calendar.google.com/calendar/u/0?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ)
* Open a [GitHub Issue](https://github.com/gemaraproj/gemara/issues)
