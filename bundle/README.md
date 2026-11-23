# Gemara bundle format, version 1

Status: experimental.

A Gemara bundle stores one Gemara document, and optionally the documents it
depends on, as an [OCI artifact](https://github.com/opencontainers/image-spec/blob/v1.1.1/manifest.md).
This page defines the format and the check that confirms an unpacked bundle
holds the content a given manifest describes. It is written from the behaviour
of existing producers, so bundles they have already published conform.

[`bundle.cue`](bundle.cue) gives the values and shapes in CUE. This page adds
the rules CUE cannot express. The [conformance test data](conformance/) are
example bundles that any implementation can test against.

The key words MUST, MUST NOT, SHOULD and MAY are used as in
[RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

## Manifest

A bundle is an OCI image manifest with:

| Field | Value |
|---|---|
| `schemaVersion` | `2` |
| `mediaType` | `application/vnd.oci.image.manifest.v1+json` |
| `artifactType` | `application/vnd.gemara.bundle.v1` |
| `config.mediaType` | `application/vnd.gemara.manifest.v1+json` |

### Layers

Each Gemara document is one layer with media type
`application/vnd.gemara.artifact.v1+yaml`. Its content is the document's YAML,
byte for byte. Layer annotations:

| Annotation | Required | Value |
|---|---|---|
| `org.opencontainers.image.title` | yes | The document's file name. One path segment: no `/` or `\`, not `.`, `..`, `bundle.json` or `references`. |
| `org.gemara.artifact.role` | yes | `artifact` or `import`. |
| `org.gemara.artifact.type` | no | The document's Gemara artifact type, for example `ControlCatalog`. Informative only. |

Roles:

- `artifact`: the document the bundle is about. A bundle has exactly one.
- `import`: a document that the `artifact` document, or another import,
  references. A bundle has zero or more.

Within one role, no two layers have the same file name.

A consumer that reads a bundle without a role annotation on any layer: see
[Legacy bundles](#legacy-bundles).

Producers SHOULD put the `artifact` layer first, then the `import` layers.
Consumers MUST NOT depend on layer order.

A layer with any other media type is not part of the format. Consumers MUST
ignore it, and it has no place in the unpacked layout.

### Manifest annotations

The format defines one manifest annotation:

| Annotation | Required | Value |
|---|---|---|
| `org.opencontainers.image.licenses` | no | The publication licence of the bundle, as an SPDX expression. |

Every other manifest annotation, including `org.opencontainers.image.created`,
is left to producers and consumers. The format gives it no meaning. All
manifest annotations are still covered by the manifest digest.

## Config

The config blob is a JSON object. An unpacker may also write it to disk as
`bundle.json` (see [Unpacked layout](#unpacked-layout)).

The config media type is `application/vnd.gemara.manifest.v1+json`. Despite
its name, it identifies the config blob, not the OCI image manifest. It is
fixed for version 1. A future version should use a name that says "config",
such as `application/vnd.gemara.bundle.config.v2+json`.

| Field | Required | Meaning |
|---|---|---|
| `bundle-version` | yes | The `artifact` document's `metadata.version`, exactly. It is **not** the format version, and consumers MUST NOT read it as one. |
| `gemara-version` | yes | The Gemara specification version the bundled documents target. |
| `revision` | no | A revision label the producer chooses. |
| `artifacts` | no | One entry per bundled document: `name`, `type`, `id`, `role`, and optional `dependencies` (names of the documents it references). |

Producers MUST set `bundle-version` to the `artifact` document's
`metadata.version`. Consumers MUST NOT reject a bundle because of its
`bundle-version`: bundles already published carry other values, such as
`1.0`. Consumers MUST ignore fields they don't recognise.

Provenance is not yet part of the format. Signed provenance is better carried
as an OCI referrer of the manifest.

The `artifacts` list describes the layers; it does not define them. The layers
are the source of truth for what a bundle contains.

## Unpacked layout

An unpacked bundle is a directory:

```
<title>                the artifact layer
imports/<title>        each import layer
bundle.json            optional: a copy of the config, for people to read
references/
```

Layer files MUST be the layer's exact bytes.

`bundle.json` is a rendering, not content. An unpacker MAY write the config's
exact bytes, MAY write a re-encoded copy (indented for reading, for example),
or MAY leave it out. The content check never reads it. An implementation that
wants to check the config later, offline, must keep the config's exact bytes
somewhere else, such as its own cache. This keeps existing `bundle.json`
files, and tools that read them, working.

Tools that unpack a bundle should not create or follow symlinks in the
unpacked layout. A symlink in place of a layer could make a tool read or write
a file outside it.

## Content check

The content check confirms that an unpacked bundle holds exactly the content
that a manifest describes. Its inputs are:

- the manifest bytes;
- the digest the caller expects for those bytes, for example a digest whose
  signature the caller has already verified;
- the config blob's bytes, as the implementation holds them (fetched from the
  registry, or kept in a cache), or nothing if it holds none;
- the unpacked directory.

Implementations MUST support `sha256` digests and MAY support other
algorithms. A digest with an algorithm the implementation does not support is
still well formed, as in OCI; the check cannot verify it, so it fails with
`unsupported-digest` rather than `malformed-manifest`.

The check does not fetch, sign or verify signatures. Run these steps in order
and stop at the first failure:

1. **Digest.** If the implementation does not support the algorithm named in
   the expected digest, fail with `unsupported-digest`. Otherwise hash the
   manifest bytes with it; if the result differs, fail with `digest-mismatch`.
2. **Manifest.** Parse the bytes as JSON. Fail with `malformed-manifest` if
   parsing fails, if `schemaVersion` or `mediaType` differs from the values
   above, if `config.mediaType` is not the Gemara config media type, or if
   `artifactType` is present and is not the bundle type. (A missing
   `artifactType` is a [legacy bundle](#legacy-bundles).)
3. **Roles.** For every layer with the Gemara layer media type, if the role
   annotation is present and is neither `artifact` nor `import`, fail with
   `unknown-role`. If it is missing, apply the [legacy rule](#legacy-bundles);
   if that rule does not apply, fail with `unknown-role`.
4. **Structure.** Fail with `malformed-manifest` if a Gemara layer's title is
   missing or not a valid file name, if two layers of the same role share a
   title, or if the number of `artifact` layers is not exactly one. Then, if
   the config's or a Gemara layer's digest uses an algorithm the
   implementation does not support, fail with `unsupported-digest`.
5. **Content.** If the implementation holds no config bytes, fail with
   `missing-config`. If their size or digest differs from `config`, fail with
   `mismatch`. Then, for each Gemara layer, at its path in the unpacked layout:
   - if nothing is there, fail with `missing-file`;
   - if what is there is not a regular file (a symlink, for example), or its
     size or digest differs from the layer's, fail with `mismatch`.
     Implementations MUST NOT follow symlinks.
6. **Extras.** If the directory holds any file other than the layer files,
   `bundle.json` and files under `references/`, fail with `extra-file`. Empty
   directories are ignored.

`references/` holds material related to the bundle that isn't part of its
signed content, such as dependencies, provenance or attestations. Its layout
is up to the tool that writes it. The content check ignores references/, so
its contents aren't covered by the bundle's signature.

If every step passes, the result is `ok`. A consumer SHOULD warn when it
accepted a legacy bundle.

The result names are the names the conformance test data uses. Implementations
may report them in whatever form their language uses, for example as error
types, as long as each one can be told apart.

The config's `artifacts` list is covered by the config digest, so the check
authenticates it. The check does not compare it with the layers.

Implementations SHOULD bound how much they read from a registry for one
bundle, and let callers change the bound.

## Legacy bundles

Every producer the authors know of writes the role annotation and `artifactType`. Bundles from older or other
producers may lack them. So that such bundles stay readable, a consumer
accepts these two cases, and SHOULD warn when it does:

- **No `artifactType`**, when `config.mediaType` is the Gemara config media
  type. The bundle is read as version 1.
- **No role annotation** on exactly one Gemara layer, when no other layer has
  role `artifact` and every other Gemara layer has role `import`. That layer is
  read as the `artifact` layer.

Producers MUST NOT write either shape. `bundle.cue` describes what producers
write, so it rejects both.

## Versioning

The format version is the `v1` in the three media types. It is not
`bundle-version`.

Within version 1, the format may only change in ways a version 1 consumer
already handles: new optional config fields, and new optional manifest or
layer annotations. Anything else needs version 2 and new media types,
including:

- a new layer role, because version 1 consumers reject unknown roles;
- a new required config field or annotation;
- a change to the unpacked layout.

A consumer that meets an `artifactType`, config media type or Gemara layer
media type with a version it does not implement MUST reject the bundle rather
than read it as version 1. A version 1 consumer MUST reject a Gemara layer with
an unknown role (step 3).

## Differences from existing implementations

The rules above follow existing producers: everything they write today passes
the content check. Some existing consumers are more lenient than this spec.

1. **Unknown roles.** Some consumers read any Gemara layer whose role is not
   `import` as the `artifact` layer. This spec rejects an unknown role, because
   reading a future role as the document itself is wrong. A missing role is
   still read under the legacy rule.
2. **`bundle.json`.** Some unpackers write a re-encoded copy of the config.
   This spec allows that and keeps `bundle.json` out of the check; the check
   uses the config's exact bytes instead.
3. **`bundle-version` has two meanings today.** One producer defaults it to the
   source document's `metadata.version`; another always writes `1.0`. This spec
   requires the `artifact` document's `metadata.version`, and requires
   consumers to accept bundles already published with other values.
4. **Config media type.** Some consumers read the config only when its media
   type is the Gemara config type, and otherwise carry on with an empty config.
   This spec rejects any other config media type.
5. **`artifactType`.** Existing consumers do not check it. This spec rejects a
   different value and accepts a missing one as legacy.
6. **File names.** Existing producers do not check layer titles, and one names
   imports after the last segment of their URL, so two imports can share a
   title. This spec rejects duplicate titles within a role and titles that are
   not a single path segment. A producer should report the clash when it
   assembles the bundle, before publishing.
7. **No artifact layer.** Some consumers accept a manifest with no artifact
   layer and return an empty document. This spec requires exactly one.
8. **The `artifacts` list** is not checked against the layers anywhere today.
   This spec keeps it informative.
