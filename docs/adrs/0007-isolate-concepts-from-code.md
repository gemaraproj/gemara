---
layout: page
title: Isolate Concepts from Code
---

- **ADR:** 0007
- **Proposal Author(s):** @eddie-knight, @jpower432
- **Status:** Accepted

## Context

While [ADR-0006](./0006-unified-package-structure) served to simplify the Go SDK for
developer users, added cognitive overhead was observed for visitors to the project repo,
caused by the large number of Go files. The project had begun to appear as if it is
primarily a software project, rather than a specification project with support utilities.

Additionally, the previous implementation of ADR-0006 did not account for SDKs in other
languages beyond Go, which is highly likely if Gemara evolves like other similar projects.

## Decision

All SDK code should be hosted in secondary repositories.

## Consequences

Complexity on first move — this will delay us a bit on our pursuit toward v1.

Lowered complexity over time — schema changes can be incremental, and SDK doesn't need to
be updated until a schema release is made.

No additional overhead needed to manage the release processes and versioning for the schemas and different SDKs.

SDKs can be adjusted to support multiple schema versions if needed later.

## Alternatives Considered

We could create a /go or /sdk subdirectory instead. We'd then use the website, language-specific package repositories, and some CI magic to handle the release cycles for the different types of assets.
