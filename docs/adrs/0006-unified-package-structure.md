---
layout: page
title: Unified Go SDK Package Structure
---

- **ADR:** 0006
- **Proposal Author(s):** @jpower432
- **Status:** Accepted; Modified by [ADR 0007](./0007-isolate-concepts-from-code) (SDK isolation); ; Modified by [ADR-0009](./0009-sensitive-activities) (numbering for layers 3+)

## Context

The Gemara Go module initially organized code by the conceptual layers of the Gemara model:
- `layer1/` - Guidance documents (Layer 1)
- `layer2/` - Control catalogs (Layer 2)  
- `layer3/` - Policy documents (Layer 3)
- `layer4/` - Evaluation logs and results (Layer 4)

This structure mirrored the conceptual model described in the README, where each layer builds upon lower layers.

However, over time some issues emerged:

1. **Type Sharing**: Many types are shared across layers (e.g., `Metadata`, `Contact`, `Mapping`, `Date`). These were duplicated between packages.

2. **Cross-Layer Usage**: Higher layers frequently reference lower layers:
   - Layer 4 (Evaluation) references Layer 2 (Catalog) controls
   - Converters need types from multiple layers
   - Loaders share common logic

3. **Import Complexity**: Consumers needed to import multiple packages:
   ```go
   import (
       "github.com/ossf/gemara/layer1"
       "github.com/ossf/gemara/layer2"
       "github.com/ossf/gemara/layer4"
   )
   ```

## Decision

We consolidated all layer Go packages into a single unified package: `package gemara` at the module root.

### New Structure

All Go files are now in the root package:
- `generated_types.go` - All types from all layers (generated from CUE schemas)
- `loaders.go` - Unified loader functions for all document types
- `assessment_log.go` - Layer 4 evaluation functionality
- `control_evaluation.go` - Layer 4 control evaluation
- `evaluation_plan.go` - Layer 4 evaluation planning
- `result.go` - Layer 4 result types
- `actor_type.go` - Actor type definitions
- `document_example_test.go` - Layer 1 examples
- `test-data.go` - Shared test data

### Package Organization Principles

1. **Single Import**: Consumers import one package: `github.com/ossf/gemara`
   With unified package, relationships are explicit:
   ```go
   // Cohesive - relationships clear
   import "github.com/ossf/gemara"
   
   var doc gemara.GuidanceDocument
   doc.Metadata = gemara.Metadata{...}  // Same namespace = clear relationship
   ```

   The unified package makes it immediately obvious that `Metadata`, `GuidanceDocument`, `Catalog`, and `EvaluationLog` are all part of the same conceptual model, not separate concerns.

2. **Format Packages Separate**: Format converters remain separate packages (`oscal`, `sarif`) as they have different dependency profiles and are optional features

3. **Internal Utilities**: Shared implementation details go in `internal/`:
   - `internal/loaders/` - Generic file loading utilities
   - `internal/oscal/` - OSCAL-specific utilities

## Consequences

### Positive

1. **Simplified Imports**: Single import path for all Gemara functionality:
   ```go
   import "github.com/ossf/gemara"
   ```

2. **No Code Duplication**: Shared types and utilities defined once

3. **Easier Refactoring**: Changes to shared types automatically propagate throughout the codebase

4. **Simpler Schema Generation**: CUE generates one `generated_types.go` file, no manual splitting needed

5. **Unified API**: All functionality accessible through one package, improving discoverability

6. **Reduced Cognitive Load**: Users don't need to understand layer boundaries to use the library

### Negative

1. **Larger Package**: Single package contains all functionality, which some may consider less organized

2. **Potential Namespace Pollution**: All exported types in one namespace (mitigated by clear naming conventions)

3. **Migration Effort**: Existing consumers needed to update imports (though straightforward find/replace)

### Neutral

1. **Documentation**: Package documentation can reference layers conceptually while types are unified

2. **Testing**: Tests remain organized by functionality, not package boundaries

3. **Schema Organization**: CUE schemas remain organized by layer (`schemas/layer-1.cue`, etc.)

## Alternatives Considered

### Alternative: Keep Layer Packages, Extract Common Types

Create a `common/` package for shared types (e.g., `Metadata`, `Actor`, `Mapping`, `Date`), keep layer-specific code in layer packages.

**Pros:**
- Maintains conceptual alignment with the model
- Clear separation of concerns
- Solves type sharing problem - shared types defined once in `common/`
- Reduces code duplication

**Cons:**
- Requires multiple imports (`layer1`, `layer2`, `common`)
- Semantic clarity tradeoff: Shared types would be prefixed with `common.` (e.g., `common.Metadata`, `common.Actor`), which doesn't add meaning - "common" is organizational, not semantic
- Converters still need multiple imports for cross-layer usage

**Decision**: Rejected - while this solves the type sharing problem, the semantic clarity tradeoff was considered worse than the unified package approach. Types like `Metadata` and `Actor` are core Gemara concepts, not "common utilities" - they deserve the same namespace as layer-specific types. The unified package provides better semantic clarity (`gemara.Metadata` vs `common.Metadata`) while solving all the same problems.

## References

- [Go Package Design](https://go.dev/blog/package-names)
- [Effective Go - Packages](https://go.dev/doc/effective_go#names)
- [Gemara Model Documentation](/model)
