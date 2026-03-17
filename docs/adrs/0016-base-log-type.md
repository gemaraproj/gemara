---
layout: page
title: Base Log Type
---

- **ADR:** 0015
- **Proposal Author(s):** @eddie-knight, @jpower432
- **Status:** Accepted

## Context

[ADR-0012](./0012-schema-types) established "Log" as the term for artifacts containing multiple measurements and identified several future log types: Evaluation Log, Enforcement Log, Monitoring Log, and Audit Log.
As new log types are added, each would redeclare those same fields independently — duplicating structure and risking divergence.

## Decision

Introduce a `#Log` base definition in `log.cue` containing the fields common to all log artifacts:

| **Field** | **Type** | **Purpose** |
|:---|:---|:---|
| `metadata` | `#Metadata` | Standard artifact metadata |
| `target` | `#Resource` | Resource being evaluated |

Concrete log types embed `#Log` and add their specific entries. `#EvaluationLog` is refactored to use the base types:

```cue
#EvaluationLog: {
    #Log
    evaluations: [#ControlEvaluation, ...#ControlEvaluation]
}
```

Future types (e.g., `#EnforcementLog`, `#AuditLog`) follow the same pattern.

## Consequences

- `#EvaluationLog` no longer declares `metadata` and `target` directly; changes to shared log fields propagate from one definition
- New log types require only embedding `#Log` plus their own entries
- No breaking change to the serialized YAML/JSON shape — embedded fields are structurally identical to inline fields in CUE

## Alternatives Considered

### Keep fields inline on each log type

Each log type declares its own `metadata` and `target`. Simpler today with one type, but violates single-source-of-truth as the log family grows per ADR-0012.
