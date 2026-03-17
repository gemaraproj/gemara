---
layout: page
title: Nest Imported Fields Under imports Struct
---

- **ADR:** 0015
- **Proposal Author(s):** @jpower432
- **Status:** Accepted

## Context

`ControlCatalog`, and `ThreatCatalog` used top-level hyphenated fields (`imported-controls`, `imported-threats`, `imported-capabilities`) to declare entries sourced from external artifacts.
CUE cannot reference hyphenated sibling fields from hidden validation structs because quoted identifiers in expression context resolve as string literals, not field references.

## Decision

Replace top-level `imported-*` fields with a nested `imports` struct on each catalog type, mirroring the Policy pattern.

| **Before** | **After** |
|:---|:---|
| `imported-controls: [...]` | `imports.controls: [...]` |
| `imported-threats: [...]` | `imports.threats: [...]` |
| `imported-capabilities: [...]` | `imports.capabilities: [...]` |

New CUE definitions introduced:

- `#ControlCatalogImports` — wraps `controls`
- `#ThreatCatalogImports` — wraps `threats` and `capabilities`

## Consequences

Breaking change for any adopter using `imported-controls`, `imported-threats`, or `imported-capabilities` in YAML/JSON artifacts

