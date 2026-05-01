---
layout: page
title: Go SDK
---

The Go SDK provides type-safe APIs for reading, writing, converting, and bundling Gemara documents. Types are generated from CUE schemas using [cuegen](https://github.com/gemaraproj/cuegen).

**[Go Package Reference →](https://pkg.go.dev/github.com/gemaraproj/go-gemara)**

## Installation

```bash
go get github.com/gemaraproj/go-gemara
```

## Loading Documents

`gemara.Load` is a generic loader — substitute the type parameter for any document kind (`gemara.GuidanceCatalog`, `gemara.ControlCatalog`, `gemara.Policy`, `gemara.EvaluationLog`, etc.). Format (YAML or JSON) is detected from the file extension.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gemaraproj/go-gemara"
    "github.com/gemaraproj/go-gemara/fetcher"
)

func main() {
    f := &fetcher.File{}
    ctx := context.Background()

    catalog, err := gemara.Load[gemara.ControlCatalog](ctx, f, "path/to/controls.yaml")
    if err != nil {
        log.Fatal(err)
    }

    for _, control := range catalog.Controls {
        fmt.Printf("Control: %s - %s\n", control.Id, control.Title)
    }
}
```

The `fetcher` package provides three implementations of the `gemara.Fetcher` interface:

- `fetcher.File` — reads from the local filesystem
- `fetcher.HTTP` — fetches over HTTP(S)
- `fetcher.URI` — dispatches by scheme (`file://`, `http(s)://`, or bare paths)

`ControlCatalog` and `GuidanceCatalog` also expose `LoadFiles` for merging multiple sources, and `ControlCatalog.LoadNestedCatalog` for YAML files where the catalog is wrapped in a single key (e.g. `catalog:`).

## Converting to OSCAL

```go
import (
    "github.com/gemaraproj/go-gemara"
    "github.com/gemaraproj/go-gemara/fetcher"
    "github.com/gemaraproj/go-gemara/gemaraconv"
)

f := &fetcher.File{}
ctx := context.Background()

catalog, _ := gemara.Load[gemara.ControlCatalog](ctx, f, "path/to/catalog.yaml")
oscalCatalog, _ := gemaraconv.ControlCatalog(catalog).ToOSCAL()

guidance, _ := gemara.Load[gemara.GuidanceCatalog](ctx, f, "path/to/guidance.yaml")
oscalCat, oscalProfile, _ := gemaraconv.GuidanceCatalog(guidance).ToOSCAL("relative/path/to/catalog.json")
```

A `ControlCatalog` can also be rendered to Markdown via `gemaraconv.ControlCatalog(catalog).ToMarkdown(ctx)`.

## Converting to SARIF

`EvaluationLog` results can be emitted as SARIF for surfacing in code-scanning tools. A `ControlCatalog` is required to resolve control metadata referenced from the log.

```go
catalog, _ := gemara.Load[gemara.ControlCatalog](ctx, f, "path/to/catalog.yaml")

evaluationLog := &gemara.EvaluationLog{ /* populate */ }
sarifBytes, _ := gemaraconv.EvaluationLog(evaluationLog).ToSARIF("path/to/artifact.md", catalog)
```

## Bundling and Distributing via OCI

The `bundle` package assembles the full dependency tree (`extends` + `imports`) of a Gemara document, packs it into an OCI layout, and pushes it to a registry.

```go
import (
    "github.com/gemaraproj/go-gemara/bundle"
    "github.com/gemaraproj/go-gemara/fetcher"
    "oras.land/oras-go/v2"
    "oras.land/oras-go/v2/content/oci"
    "oras.land/oras-go/v2/registry/remote"
)

data, _ := os.ReadFile("policy.yaml")
src := bundle.File{Name: "policy.yaml", Data: data}

m := bundle.Manifest{BundleVersion: "1", GemaraVersion: "v1.0.0"}
asm := bundle.NewAssembler(&fetcher.URI{})
b, _ := asm.Assemble(ctx, m, src)

layoutStore, _ := oci.New("./bundle-output")
desc, _ := bundle.Pack(ctx, layoutStore, b)
_ = layoutStore.Tag(ctx, desc, "v1.0.0")

repo, _ := remote.NewRepository("registry.example.com/org/bundle")
tagDesc, _ := layoutStore.Resolve(ctx, "v1.0.0")
_ = oras.CopyGraph(ctx, layoutStore, repo, tagDesc, oras.DefaultCopyGraphOptions)
_ = repo.Tag(ctx, tagDesc, "v1.0.0")

unpacked, _ := bundle.Unpack(ctx, repo, "v1.0.0")
_ = unpacked
```

## CLI: `oscalexport`

The repository also ships an `oscalexport` command that converts Gemara documents to OSCAL without writing any Go.

```bash
# Build
make build

# Convert a Control Catalog
./bin/oscalexport catalog ./path/to/catalog.yaml --output ./catalog.json

# Convert a Guidance Catalog (emits an OSCAL Catalog and Profile)
./bin/oscalexport guidance ./path/to/guidance.yaml \
    --catalog-output ./guidance.json \
    --profile-output ./profile.json
```

## Relationship to Other Components

### [The Model](../model)
Provides the conceptual foundation. Go SDK types correspond to elements in the model.

### [The Schemas](../schema/)
Go SDK types are generated from the CUE schemas, ensuring consistency between validation and programmatic access. The schema version supported by a given SDK release is exposed as `gemara.SchemaVersion`.
