---
layout: page
title: Go SDK
---

The Go SDK provides type-safe APIs for reading, writing, and manipulating Gemara documents. Types are generated from CUE schemas using [cuegen](https://github.com/gemaraproj/cuegen).

**[Go Package Reference →](https://pkg.go.dev/github.com/gemaraproj/gemara)**

## Installation

```bash
go get github.com/gemaraproj/gemara
```

## Usage

```go
import "github.com/gemaraproj/gemara"

// Load a control catalog
catalog := &gemara.Catalog{}
catalog, err := catalog.LoadFile("file://controls.yaml")
if err != nil {
    log.Fatal(err)
}

// Access controls
for _, control := range catalog.Controls {
    fmt.Printf("Control: %s - %s\n", control.ID, control.Title)
}
```

## Relationship to Other Components

### [The Model](../model)
Provides the conceptual foundation. Go SDK types correspond to elements in the model.

### [The Schemas](../schema/)
Go SDK types are generated from the CUE schemas, ensuring consistency between validation and programmatic access.
