---
layout: page
title: CUE Validation Subpackage
---

- **ADR:** 0016
- **Proposal Author(s):** @jpower432
- **Status:** Accepted

## Context

Gemara base types are defined in root-level CUE files (`package gemara`). CUE tooling such as `cue def` and `cue exp gengotypes` is used against this package for Go type generation. These tools operate on a single package: all `.cue` files in the target directory are unified.

Cross-field constraints are necessary for artifact correctness, but when co-located with the base type definitions, they are included in the unified output.
This causes `cue exp gengotypes` to produce Go structs that use embedded structs instead of common types like `Metadata`.

## Decision

We will define cross-field validation constraints in a CUE subpackage `validation`. 
The subpackage imports the root package CUE files and layers constraints on top of base types.

Each validation definition embeds the base type and adds cross-field rules:
```cue
#ControlCatalog: {
    gemara.#ControlCatalog
    // cross-field constraints (uniqueness, referential integrity, etc.)
}
```

Consumers can validate artifacts against the validation package via `cue vet`:

```bash
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@latest:validation data.yaml
```

## Consequences

- Contributors must understand that type definitions and validation rules live in separate directories 
- Adding a new field to a base type that needs cross-field validation requires edits in both packages

## Alternatives Considered

### Go-side validation

Implement cross-field rules in Go code rather than CUE, but expressing validation in Go would split the source of truth for the spec between CUE and Go.

## References

- [ADR 0006: Unified Go SDK Package Structure](./0006-unified-package-structure.md)
- [ADR 0015: Schema Organization by Artifact Type](./0015-schema-organization-by-artifact-type.md)
- [CUE Modules and Packages](https://cuelang.org/docs/concept/modules-packages-instances/)
