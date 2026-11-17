# Inline Mapping Prototype Findings

The prototype resolves a MappingDocument and a source ControlCatalog into one
ControlCatalog that preserves per-target evidence, independent assertions about
the same pair of entries, and explicit coverage gaps alongside existing legacy
links. The resolver is test-only. The schema changes are experimental.

## Intentional Difference In Scope

Inline mappings are a methodology-constrained subset of MappingDocument
mappings. They follow the cross-layer linkages prescribed by the Gemara
methodology described in the whitepaper, with a fixed `source.target`
direction. MappingDocument deliberately permits broader source/target
combinations and explicit relationship assertions.

The two formats share evidence semantics, with intentionally different
expressive scope. Full representational parity is not the goal. Inline links
default to `relates-to`; authors can supply explicit relationships and
per-target evidence when warranted. Strength, confidence, rationale, and
applicability retain their meaning in either format.

Only mappings representable by an allowed inline linkage can be folded into a
catalog. A resolver must preserve their direction, relationship, and supplied
evidence. Unsupported mappings must remain external or produce an explicit
unsupported-mapping result; a resolver must not silently reinterpret them to
fit an inline field.

## Shareable Examples

The [source catalog](../../examples/inline-mapping-prototype/source-catalog.yaml),
[mapping document](../../examples/inline-mapping-prototype/mapping-document.yaml),
and [resolved catalog](../../examples/inline-mapping-prototype/resolved-catalog.yaml)
use invented controls and guidance identifiers. Their scores illustrate the
format and make no claims about a real framework.

The example includes two assertions about the same source/target pair, with
different applicability and scores, a multi-target assertion, a legacy link,
and an explicit `no-match` gap. Tests compare the resolved fixture with the
resolver's output after CUE default selection. All three files also validate
independently against CUE and generated OpenAPI.

## Schema Contract

`#MappingTargetMetadata` shares optional `strength` (integer 1–10),
`confidence-level`, `applicability`, `rationale`, and `remarks` between
MappingDocument targets and catalog targets. Evidence has no defaults.
Unreported strength or confidence stays absent.

`#CatalogEntryMapping` extends `#EntryMapping` for individual semantic links.
The existing grouped catalog format needs its own types:
`#CatalogMultiEntryMapping` holds an assertion's relationship and optional ID;
`#CatalogMappingTarget` holds each target's evidence. All seven cross-layer
linkage fields use the grouped type:

| Source entry | Cross-layer linkage fields |
| --- | --- |
| Control | `guidelines`, `threats` |
| Threat | `capabilities`, `vectors` |
| Guideline | `principles`, `vectors` |
| Risk | `threats` |

Inline direction follows `source.target`: the containing object is always the
source, and the linkage field identifies the targets. For `Control.threats`,
the Control is the source and each referenced Threat is a target. An explicit
relationship is read from that Control to those Threats; it does not reverse
direction based on the relationship's wording. Likewise, `Control.guidelines`
with `relationship: implements` means the Control implements the referenced
Guideline. Resolution must preserve this direction or explicitly invert the
relationship when reversing a MappingDocument's endpoints.

The extension is limited to those cross-layer linkages. Catalog `extends` and
`imports`, guideline `extends`, exemption `redirect`, and lifecycle `replaced-by`
pointers retain their existing types.

For compatibility, a group's `reference-id` identifies the target artifact,
while each nested entry's `reference-id` identifies an entry within it.
Resolution renames MappingDocument target `entry-id` to this legacy field:

```yaml
guidelines:
  - id: inventory-manual
    reference-id: DEMO-GUIDANCE
    relationship: implements
    entries:
      - reference-id: DEMO.G01
        strength: 6
        confidence-level: Medium
        applicability: [manual]
        rationale: Manual collection may omit dependencies.
  - id: ownership-gap
    reference-id: DEMO-GUIDANCE
    relationship: no-match
    remarks: The target guidance has no ownership requirement.
```

Omitted relationships resolve to `relates-to`. Other than `no-match`, every
group requires at least one target. A `no-match` assertion can omit entries;
the prototype does not forbid supplied targets, matching MappingDocument's
current behavior. An empty entries list is always invalid.

## Compatibility And Projection Findings

Existing catalog fixtures remain valid with the updated CUE and OpenAPI
schemas. That is input compatibility only. Older closed schemas may reject
new fields, and generated Go consumers must accommodate the new catalog types.
Downstream Go source compatibility has not been verified. CUE and OpenAPI
tests exercise legacy links, supplied evidence,
coverage gaps, and invalid values at each of the seven cross-layer fields.

Two projection fixes were necessary. Defaulted properties could remain required
inside inline `allOf` constraints. The converter now removes those unconditional
requirements without changing alternative branches. It also translates the
narrow CUE presence rule `if relationship != "no-match" {entries: T}` when an
identical optional `entries?: T` declaration exists. OpenAPI then requires
either the explicit exception or entries, accounting for the selector's default.
Unsupported conditionals retain their existing omission warnings.

## Resolver Boundaries

The resolver probe handles Control-to-Guideline mappings into the source catalog
only and rejects other endpoint-type combinations. The schema supports all
seven cross-layer fields listed above; that does not imply support for every
MappingDocument combination.
It expects CUE-validated inputs, checks the source ID/version and source entry
existence, and rejects reference or applicability alias collisions. It copies
target reference declarations and applicability definitions into the result.
It does not merge aliases or resolve catalogs by URL.

The result distributes source content and mapping assertions together. Target
catalog content remains external, so target existence and meaning require a
separate check. Full MappingDocument provenance, including its author and
version, is not retained as assertion provenance. A production resolver needs
an explicit provenance contract and collision policy before supporting arbitrary
catalog pairs or repeat resolution.

## Reproduction

From the repository root, run the full CUE/Go and OpenAPI suites, then regenerate
the local OpenAPI document and lifecycle manifest:

```sh
make test
make genopenapi
```

The generated files live under ignored `generated/`. The versioned OpenAPI
regression snapshot is `cmd/internal/cmd/testdata/openapi.golden.yaml`.
To export the example through the resolver, run from `test/`:

```sh
GEMARA_PROTOTYPE_OUTPUT=/tmp/resolved-catalog.yaml go test -count=1 -run '^TestMappingResolutionPrototype$' .
```

The tests check evidence retention, independent assertion IDs, defaults, and
non-mutation of the source. They also check gap targets and reject invalid
relationship, strength, confidence, and target-presence cases. Schema validation
checks the structure; the assertions still need an author's substantive assessment.
