# SPDX-License-Identifier: Apache-2.0
"""Generate bundle conformance test data into <out>/ (argv[1], default test-data/ beside this script), one JSON file per test case."""
import base64, hashlib, json, os, shutil, sys

OUT = sys.argv[1] if len(sys.argv) > 1 else os.path.join(os.path.dirname(os.path.abspath(__file__)), "test-data")
LAYER = "application/vnd.gemara.artifact.v1+yaml"
CONFIG = "application/vnd.gemara.manifest.v1+json"
ART = "application/vnd.gemara.bundle.v1"
ROLE, TYPE, TITLE = "org.gemara.artifact.role", "org.gemara.artifact.type", "org.opencontainers.image.title"

POLICY = b"""metadata:
  id: example-policy
  type: Policy
  gemara-version: "1.0.0"
  version: "1.0.0"
title: Example policy
"""
CATALOG = b"""metadata:
  id: example-controls
  type: ControlCatalog
  gemara-version: "1.0.0"
  version: "2.1.0"
title: Example control catalog
"""
GUIDANCE = b"""metadata:
  id: example-guidance
  type: GuidanceCatalog
  gemara-version: "1.0.0"
  version: "0.3.0"
title: Example guidance catalog
"""


def d(b):
    return "sha256:" + hashlib.sha256(b).hexdigest()


def desc(mt, b, ann=None):
    x = {"mediaType": mt, "digest": d(b), "size": len(b)}
    if ann:
        x["annotations"] = ann
    return x


def layer(name, data, role, typ=None):
    a = {TITLE: name}
    if role:
        a[ROLE] = role
    if typ:
        a[TYPE] = typ
    return desc(LAYER, data, a)


def config(version, artifacts):
    return json.dumps({"bundle-version": version, "gemara-version": "1.0.0", "artifacts": artifacts},
                      separators=(",", ":")).encode()


def manifest(cfg, layers, artifact_type=ART):
    m = {"schemaVersion": 2, "mediaType": "application/vnd.oci.image.manifest.v1+json",
         "artifactType": artifact_type, "config": desc(CONFIG, cfg), "layers": layers,
         "annotations": {"org.opencontainers.image.created": "2026-01-01T00:00:00Z",
                         "org.opencontainers.image.licenses": "Apache-2.0"}}
    if artifact_type is None:
        del m["artifactType"]
    return json.dumps(m, indent=2).encode() + b"\n"


def render(cfg):
    """A readable rendering of the config, as an unpacker may write it to bundle.json."""
    return json.dumps(json.loads(cfg), indent=2).encode()


SINGLE_CFG = config("2.1.0", [{"name": "catalog.yaml", "type": "ControlCatalog", "id": "example-controls", "role": "artifact"}])
SINGLE_LAYERS = [layer("catalog.yaml", CATALOG, "artifact", "ControlCatalog")]
SINGLE_FILES = {"bundle.json": render(SINGLE_CFG), "catalog.yaml": CATALOG}

IMP_CFG = config("1.0.0", [
    {"name": "policy.yaml", "type": "Policy", "id": "example-policy", "role": "artifact",
     "dependencies": ["catalog.yaml", "guidance.yaml"]},
    {"name": "catalog.yaml", "type": "ControlCatalog", "id": "example-controls", "role": "import"},
    {"name": "guidance.yaml", "type": "GuidanceCatalog", "id": "example-guidance", "role": "import"},
])
IMP_LAYERS = [layer("policy.yaml", POLICY, "artifact", "Policy"),
              layer("catalog.yaml", CATALOG, "import", "ControlCatalog"),
              layer("guidance.yaml", GUIDANCE, "import", "GuidanceCatalog")]
IMP_FILES = {"bundle.json": render(IMP_CFG), "policy.yaml": POLICY,
             "imports/catalog.yaml": CATALOG, "imports/guidance.yaml": GUIDANCE}


def b64(data):
    return base64.b64encode(data).decode()


def add(name, desc_text, man, cfg, files, expect, digest=None, legacy=False):
    """cfg is the config blob the implementation holds (None: it holds none)."""
    v = {"description": desc_text, "expect": expect}
    if legacy:
        v["legacy"] = True
    v["digest"] = digest or d(man)
    v["manifest"] = b64(man)
    if cfg is not None:
        v["config"] = b64(cfg)
    v["files"] = {rel: b64(data) for rel, data in sorted(files.items())}
    with open(os.path.join(OUT, name + ".json"), "w") as f:
        json.dump(v, f, indent=2)
        f.write("\n")


shutil.rmtree(OUT, ignore_errors=True)
os.makedirs(OUT)

M1, MI = manifest(SINGLE_CFG, SINGLE_LAYERS), manifest(IMP_CFG, IMP_LAYERS)
add("valid-single", "One artifact layer; bundle.json is a re-encoded rendering, which the check ignores.",
    M1, SINGLE_CFG, SINGLE_FILES, "ok")
add("valid-imports", "One artifact layer and two import layers.", MI, IMP_CFG, IMP_FILES, "ok")
add("valid-raw-bundle-json", "bundle.json holds the config's exact bytes.",
    M1, SINGLE_CFG, {**SINGLE_FILES, "bundle.json": SINGLE_CFG}, "ok")
add("valid-no-bundle-json", "No bundle.json is written; the check uses the config bytes it holds.",
    M1, SINGLE_CFG, {"catalog.yaml": CATALOG}, "ok")
foreign = desc("text/plain", b"notes\n", {TITLE: "notes.txt"})
add("valid-foreign-layer", "A layer of another media type is ignored and has no file in the unpacked layout.",
    manifest(SINGLE_CFG, SINGLE_LAYERS + [foreign]), SINGLE_CFG, SINGLE_FILES, "ok")
add("legacy-missing-role", "The only Gemara layer has no role annotation; read as the artifact layer.",
    manifest(SINGLE_CFG, [layer("catalog.yaml", CATALOG, None, "ControlCatalog")]), SINGLE_CFG, SINGLE_FILES,
    "ok", legacy=True)
add("legacy-no-artifact-type", "The manifest has no artifactType but has the Gemara config media type.",
    manifest(SINGLE_CFG, SINGLE_LAYERS, artifact_type=None), SINGLE_CFG, SINGLE_FILES, "ok", legacy=True)
add("digest-mismatch", "The manifest bytes do not hash to the requested digest.",
    M1, SINGLE_CFG, SINGLE_FILES, "digest-mismatch", digest=d(b"some other manifest"))
add("malformed-artifact-type", "The manifest's artifactType is not the bundle type.",
    manifest(SINGLE_CFG, SINGLE_LAYERS, artifact_type="application/vnd.example.other.v1"), SINGLE_CFG,
    SINGLE_FILES, "malformed-manifest")
add("malformed-two-artifacts", "Two layers have role artifact.",
    manifest(IMP_CFG, [IMP_LAYERS[0], layer("catalog.yaml", CATALOG, "artifact", "ControlCatalog")]), IMP_CFG,
    {"bundle.json": render(IMP_CFG), "policy.yaml": POLICY, "catalog.yaml": CATALOG}, "malformed-manifest")
add("unknown-role", "A Gemara layer has a role this version does not define.",
    manifest(SINGLE_CFG, SINGLE_LAYERS + [layer("evidence.yaml", GUIDANCE, "evidence")]), SINGLE_CFG,
    SINGLE_FILES, "unknown-role")
add("mismatch-layer", "An import file differs from its layer.",
    MI, IMP_CFG, {**IMP_FILES, "imports/catalog.yaml": CATALOG.replace(b"2.1.0", b"2.1.1")}, "mismatch")
add("mismatch-config", "The config bytes held differ from the config descriptor.",
    M1, SINGLE_CFG.replace(b"1.0.0", b"9.9.9", 1), SINGLE_FILES, "mismatch")
add("missing-file", "An import layer has no file.",
    MI, IMP_CFG, {k: v for k, v in IMP_FILES.items() if k != "imports/guidance.yaml"}, "missing-file")
add("missing-config", "The implementation holds no config bytes.",
    M1, None, SINGLE_FILES, "missing-config")
add("extra-file", "A file in the unpacked layout matches no layer.",
    M1, SINGLE_CFG, {**SINGLE_FILES, "imports/extra.yaml": GUIDANCE}, "extra-file")
add("valid-references", "Files under references/ are ignored by the content check.",
    M1, SINGLE_CFG, {**SINGLE_FILES, "references/provenance.json": b"{}\n"}, "ok")
for title in ["../x", "a/b.yaml", "bundle.json", "references"]:
    add("malformed-title-" + title.replace("../", "dotdot-").replace("/", "-").replace(".", "-"),
        f"A layer title ({title}) is not a valid file name.",
        manifest(SINGLE_CFG, [layer(title, CATALOG, "artifact", "ControlCatalog")]), SINGLE_CFG,
        SINGLE_FILES, "malformed-manifest")
add("malformed-duplicate-title", "Two import layers have the same title.",
    manifest(IMP_CFG, IMP_LAYERS + [layer("catalog.yaml", GUIDANCE, "import", "GuidanceCatalog")]), IMP_CFG,
    IMP_FILES, "malformed-manifest")
add("malformed-config-media-type", "The config media type is not the Gemara config media type.",
    manifest(SINGLE_CFG, SINGLE_LAYERS).replace(CONFIG.encode(), b"application/vnd.oci.image.config.v1+json", 1),
    SINGLE_CFG, SINGLE_FILES, "malformed-manifest")
add("legacy-missing-role-imports", "The artifact layer has no role annotation and the other layers are imports.",
    manifest(IMP_CFG, [layer("policy.yaml", POLICY, None, "Policy")] + IMP_LAYERS[1:]), IMP_CFG, IMP_FILES,
    "ok", legacy=True)
_odd = layer("guidance.yaml", GUIDANCE, "import", "GuidanceCatalog")
_odd["digest"] = "example:" + hashlib.sha256(GUIDANCE).hexdigest()
add("unsupported-digest", "An import layer's digest uses an algorithm no implementation is required to support.",
    manifest(IMP_CFG, IMP_LAYERS[:2] + [_odd]), IMP_CFG, IMP_FILES, "unsupported-digest")
