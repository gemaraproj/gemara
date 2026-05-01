# Migration Guide: v1.0.0

This guide describes every breaking change introduced in `v1.0.0` and provides the schema mappings and concrete YAML/JSON rewrites needed to update your catalogs, policies, and SDK integrations. 

---

## Summary of Recent Changes

| PR                                                                              | Title                                                                                       | Breaking Change                                                                                                                                  |
| ------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| [#230](#230-controls-and-guidance-structures-flattened)                         | Flatten controls/guidance structures                                                        | `ControlFamily` removed; controls lifted out of family objects                                                                                   |
| [#239](#239-imported-guidelines-removed-from-layer-1)                           | Remove `imported-guidelines` from Layer 1                                                   | `imported-guidelines` and `imported-principles` fields removed from `GuidanceDocument`                                                           |
| [#247](#247-mapping-strength-is-now-optional)                                   | Mapping `strength` is now optional                                                          | `strength` on `#MappingEntry` is no longer required                                                                                              |
| [#299](#299-family-and-category-unified-into-group)                             | Unify `#Family` and `#Category` into `#Group`                                               | `#Family` and `#Category` types removed; replaced by `#Group`                                                                                    |
| [#312](#312-enum-values-normalized-to-title-case)                               | Normalize enum casing to Title Case                                                         | All enum values changed from lowercase/kebab-lower to Title Case                                                                                 |
| [#320](#320-metadata-fields-and-import-structure-updated)                       | Update metadata fields and import structure                                                 | `imported-controls`, `imported-threats`, `imported-capabilities` moved under `imports`; new required `type` and `gemara-version` metadata fields |
| [#359](#359-families-categories-and-applicability-categories-renamed) | Rename `families`/`categories`/`applicability-categories` → `groups`/`applicability-groups` | Top-level grouping field names and per-entry membership fields renamed across all artifact types                                                 |

---

## Schema Mapping (CUE)

### #230 Controls and Guidance Structures Flattened

**Context:** Controls were previously nested inside `#ControlFamily` objects (a `controls` list inside each family). Families are now a separate top-level list and controls reference their family by `id`. An identical flattening was applied to `GuidanceDocument`, where guidelines were previously nested inside `#Category` objects.

#### Field Name Changes

| Before                                  | After                                                                         |
| --------------------------------------- | ----------------------------------------------------------------------------- |
| `#ControlFamily`                        | `#Family` (shape: `id`, `title`, `description` only no embedded `controls`)   |
| `#Category` (guidance)                  | `#Family` (shape: `id`, `title`, `description` only no embedded `guidelines`) |
| `catalog.control-families[].controls[]` | Top-level `controls[]` with per-control `family: <id>`                        |
| `document.categories[].guidelines[]`    | Top-level `guidelines[]` with per-guideline `family: <id>`                    |

#### Structural Changes

Controls and guidelines move **out** of their parent family/category objects and into a top-level list. Each entry gains a `family` field that references its parent group by `id`.

**Control Catalog before (#230):**

```yaml
control-families:
  - id: AC
    title: Access Control
    description: Controls for access management
    controls:
      - id: AC-01
        title: Access Control Policy
        objective: Develop and document an access control policy.
        assessment-requirements:
          - id: AC-01.1
            text: Develop and document access control policy
```

**Control Catalog after (#230):**

```yaml
families:
  - id: AC
    title: Access Control
    description: Controls for access management

controls:
  - id: AC-01
    family: AC
    title: Access Control Policy
    objective: Develop and document an access control policy.
    assessment-requirements:
      - id: AC-01.1
        text: Develop and document access control policy
```

**Guidance Catalog before (#230):**

```yaml
categories:
  - id: DET
    title: Detective
    description: Detection and Continuous Improvement
    guidelines:
      - id: AIR-DET-011
        title: Human Feedback Loop for AI Systems
        objective: ...
```

**Guidance Catalog after (#230):**

```yaml
families:
  - id: DET
    title: Detective
    description: Detection and Continuous Improvement

guidelines:
  - id: AIR-DET-011
    family: DET
    title: Human Feedback Loop for AI Systems
    objective: ...
```

---

### #239 `imported-guidelines` Removed from Layer 1

**Context:** The `imported-guidelines` and `imported-principles` top-level fields on `#GuidanceDocument` have been removed. External guideline references must now be expressed via a guideline's `extends` field (pointing to an external `reference-id` and `entry-id`), which feeds into the OSCAL profile's `modify` / alterations mechanism.

#### Field Name Changes
| Before                       | After                                                                        |
| ---------------------------- | ---------------------------------------------------------------------------- |
| `imported-guidelines: [...]` | Removed. Use `guideline.extends.reference-id` + `guideline.extends.entry-id` |
| `imported-principles: [...]` | Removed. No direct replacement                                               |

#### Structural Changes

Remove any top-level `imported-guidelines` or `imported-principles` blocks from your guidance documents. If you were using `imported-guidelines` to pull external controls into an OSCAL profile, express that relationship as an `extends` on the individual guideline instead.

---

### #247 Mapping `strength` Is Now Optional

**Context:** `#MappingEntry.strength` was previously required (`int & >=1 & <=10`). It is now optional. A zero value (absence of the field) means "not yet quantified."

#### Field Name Changes

*No field renames. The field name `strength` is unchanged.*

---

### #299 `#Family` and `#Category` Unified into `#Group`

**Context:** `#Family` and `#Category` were structurally identical types. They have been merged into a single generic `#Group` type. The YAML field name `applicability-categories` was not yet changed here. That rename happens in #359.

---

### #312 Enum Values Normalized to Title Case

**Context:** Enum values across the schema used inconsistent casing (some lowercase, some kebab-lower). The convention is now Title Case for all enum values. Multi-word values use spaces, not hyphens.

---

### #320 Metadata Fields and Import Structure Updated

**Context:** Two sets of changes were made. First, `imported-controls`, `imported-threats`, and `imported-capabilities` were moved from top-level fields into a nested `imports` struct (mirroring the existing `Policy` pattern). Second, two new required fields, `type` and `gemara-version`, were added to `#Metadata`.

---

### #359 `families`, `categories`, and `applicability-categories` Renamed

**Context:** The final unification step (ADR-0020). All remaining uses of "family" and "category" as YAML field names or per-entry membership keys are replaced with "group" across every artifact type.


