# Contributing to Gemara

The project welcomes your contributions whether they be:

* reporting an [issue](https://github.com/gemaraproj/gemara/issues/new/choose)
* making a code contribution ([create a fork](https://github.com/gemaraproj/gemara/fork))
* updating our [docs](https://github.com/gemaraproj/gemara/blob/main/README.md)

## PR guidelines

All changes to the repository should be made via PR ([OSPS-AC-03](https://baseline.openssf.org/#osps-ac-03)).

PRs MUST meet the following criteria:

* Clear title that conforms to the [Conventional Commits spec](https://www.conventionalcommits.org/)
* Descriptive commit message
* DCO signoff (via `git commit -s` -- [OSPS-LE-01](https://baseline.openssf.org/#osps-le-01))
* All checks must pass ([OSPS-QA-04](https://baseline.openssf.org/#osps-qa-04))

### Useful make tasks when making schema changes

Make sure to update `schema-nav.yml` in the [website repository](https://github.com/gemaraproj/website) after adding, renaming, or removing a schema definition in a `.cue` file.
For instance, if you add a new `#NewSchema` definition to `controlcatalog.cue`, add it to the matching entry in the website's `schema-nav.yml`:
```yaml
  - title: "Control Catalog"
    filename: "controlcatalog"
    layer: 2
    schemas:
      - "ControlCatalog"
      - "Control"
      - "AssessmentRequirement"
      - "NewSchema" # new schema added
```

Use `cue fmt .` and `make cuefmtcheck` to ensure proper formatting and `make lintcue` to validate the syntax of your changes. If you forget to do this before opening a PR and your changes are invalid, the [CI workflow](.github/workflows/ci.yml) will fail and alert you.

### Adding a new artifact type

Adding a new artifact type requires schema changes, test data, and documentation updates. Follow the checklist below.

| **Step** | **Action**                                                                                    |
|:---------|:----------------------------------------------------------------------------------------------|
| 1        | Define the new `#YourArtifact` definition in its own `yourartifact.cue` file with a file-level `@gemara(status="…")` attribute (see [Projection directives](#projection-directives)) |
| 2        | Add `"YourArtifact"` to the `#ArtifactType` enum in `metadata.cue`                            |
| 3        | Add an entry for the new file to `schema-nav.yml` in the [website repo](https://github.com/gemaraproj/website), listing the artifact and every struct, enum, and alias it adds (unlisted definitions get no docs page), with a `layer:` value where applicable |
| 4        | Create a valid test data file in `test/test-data/` (prefix with `good-`)                      |
| 5        | Add positive and/or negative test cases to `test/schema_test.go`                              |
| 6        | Run `cue vet -d '#YourArtifact' . test/test-data/good-your-artifact.yaml` to validate locally |
| 7        | Run `cue fmt .` and `make cuefmtcheck` to verify formatting                                   |
| 8        | Run `make lintcue` and `make test` to confirm all checks pass                                 |

## Schema compatibility

The schemas form a public contract. A CI gate compares an OpenAPI projection of the current
schema against the latest released `v1` baseline and fails the check on any backward-incompatible
change. The gate enforces compatibility in **both directions**: it is consumer-safe (a response
may not drop a property or a guarantee that existing readers depend on) and producer-safe (a request
may not add a new requirement that existing writers cannot satisfy).

Run the check locally before opening a PR:

```
make breaking-check
```

This generates the OpenAPI projection, fetches the latest `v1` release as the baseline, and reports
any breaking changes with their [oasdiff](https://github.com/oasdiff/oasdiff) check IDs.

### Projection directives

The `cue2openapi` projection reads its Gemara-specific metadata from `@gemara(...)` attributes:

| Directive | Scope | OpenAPI output |
|:----------|:------|:---------------|
| `@gemara(status="stable")` | File | `x-status` on its definitions |
| `@gemara(format="date-time")` | Definition | `format` |
| `@gemara(default=false)` | Field with the same CUE default | `default`, and the field is optional on the wire |

Allowed status values are `experimental`, `stable`, and `deprecated`. A default directive must match
the CUE default so the validation and wire contracts cannot diverge. Keep CUE doc comments as the
source for descriptions and CUE constraints such as `time.Format` as the source for validation; these
are projected only where the upstream encoder cannot represent them.

### Making an intentional breaking change

Sometimes a breaking change is deliberate. There are exactly two sanctioned ways to introduce one:

1. **Bump the module major version to `@v2`.** A major bump establishes a new baseline lineage, so
   `v1` consumers are never silently broken. Use this when the breaking change is part of a broader,
   intentional evolution of the contract.
2. **Allowlist the specific change.** Add the offending exception to a `.oasdiff-allow` file at the
   repository root; `make breaking-check` passes it to the check via `--allow` automatically when the
   file exists. Each entry is a schema-scoped `<check-id> <schema>` pair, one per line, so the
   exception applies only to that one schema. The check ID and schema are both printed for every
   reported change, so you can copy them straight from the output — e.g. a change reported as:

   ```
   ERR[request-property-enum-value-removed] ControlEvaluation: removed the enum value ...
   ```

   is allowlisted with:

   ```
   request-property-enum-value-removed ControlEvaluation
   ```

   This MUST be done in a reviewed PR, and the file MUST include a justification comment (lines
   beginning with `#`) explaining why the change is acceptable. Use this for narrowly scoped,
   well-understood exceptions.

Do not work around the gate by any other means.

## Releases

Releases are automatically created when a PR is merged into `main` with the `release` label. To trigger a release:

1. Add the `release` label to your PR before merging
2. Merge the PR into `main`

The release workflow will automatically:
- Create a GitHub release with the appropriate version tag
- Generate release notes from the PR using release-drafter
- Publish the CUE module to the central registry

**Note:** Only PRs merged into `main` with the `release` label will trigger a release. Other labels (such as `breaking`, `feature`, or `vuln`) will not trigger releases.
