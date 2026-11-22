// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
)

const testDataDir = "../bundle/conformance/test-data"

// TestBundleTestData runs a minimal reference implementation of the content
// check in bundle/README.md against every conformance test case, and checks that
// the CUE schema accepts exactly the manifests the check finds well formed.
func TestBundleTestData(t *testing.T) {
	insts := load.Instances([]string{"./bundle"}, &load.Config{Dir: ".."})
	bundleSchema := schemaCtx.BuildInstance(insts[0])
	if err := bundleSchema.Err(); err != nil {
		t.Fatalf("building bundle schema: %v", err)
	}

	paths, err := filepath.Glob(filepath.Join(testDataDir, "*.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no test data in %s: %v", testDataDir, err)
	}

	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// []byte fields decode from base64.
			var v struct {
				Digest, Expect string
				Legacy         bool
				Manifest       []byte
				Config         []byte
				Files          map[string][]byte
			}
			if err := json.Unmarshal(raw, &v); err != nil {
				t.Fatal(err)
			}
			manifest, config := v.Manifest, v.Config // config is nil when the test case holds none
			unpacked := t.TempDir()
			for rel, data := range v.Files {
				p := filepath.Join(unpacked, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := checkBundle(manifest, v.Digest, config, unpacked); got != v.Expect {
				t.Errorf("check = %q, want %q", got, v.Expect)
			}

			// The schema describes what producers write: it rejects the manifests the
			// check calls malformed or unknown-role, and the legacy shapes the check still reads.
			cueErr := vetJSON(bundleSchema, "#Manifest", manifest)
			wantInvalid := v.Expect == "malformed-manifest" || v.Expect == "unknown-role" || v.Legacy
			if wantInvalid != (cueErr != nil) {
				t.Errorf("#Manifest validation error = %v, want invalid=%v", cueErr, wantInvalid)
			}
			if config != nil {
				if err := vetJSON(bundleSchema, "#Config", config); err != nil {
					t.Errorf("#Config: %v", err)
				}
			}
		})
	}
}

func vetJSON(schema cue.Value, def string, data []byte) error {
	v := schemaCtx.CompileBytes(data)
	if err := v.Err(); err != nil {
		return err
	}
	return schema.LookupPath(cue.ParsePath(def)).Unify(v).Validate(cue.Concrete(true))
}

type descriptor struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations"`
}

// checkBundle follows the steps of the content check in bundle/README.md and
// returns "ok" or the first failure. config is the config blob the caller
// holds, or nil. It supports sha256 digests only.
func checkBundle(manifest []byte, want string, config []byte, unpacked string) string {
	// Step 1: the manifest bytes hash to the requested digest.
	if sha256Digest(manifest) != want {
		return "digest-mismatch"
	}

	// Step 2: the manifest is a version 1 bundle manifest.
	var m struct {
		SchemaVersion int          `json:"schemaVersion"`
		MediaType     string       `json:"mediaType"`
		ArtifactType  string       `json:"artifactType"`
		Config        descriptor   `json:"config"`
		Layers        []descriptor `json:"layers"`
	}
	if json.Unmarshal(manifest, &m) != nil || m.SchemaVersion != 2 ||
		m.MediaType != "application/vnd.oci.image.manifest.v1+json" ||
		(m.ArtifactType != "" && m.ArtifactType != "application/vnd.gemara.bundle.v1") ||
		m.Config.MediaType != "application/vnd.gemara.manifest.v1+json" {
		return "malformed-manifest"
	}

	// Step 3: every Gemara layer has a known role, except the legacy case of a
	// single unannotated layer when no layer has role "artifact".
	roles := map[int]string{}
	var unannotated []int
	marked := false
	for i, l := range m.Layers {
		if l.MediaType != "application/vnd.gemara.artifact.v1+yaml" {
			continue
		}
		role, ok := l.Annotations["org.gemara.artifact.role"]
		switch {
		case !ok:
			unannotated = append(unannotated, i)
		case role == "artifact" || role == "import":
			roles[i] = role
			marked = marked || role == "artifact"
		default:
			return "unknown-role"
		}
	}
	if len(unannotated) > 0 {
		if marked || len(unannotated) > 1 {
			return "unknown-role"
		}
		roles[unannotated[0]] = "artifact"
	}

	// Step 4: names are valid and unique, and there is exactly one artifact layer.
	expected := map[string]descriptor{}
	artifacts := 0
	for i, l := range m.Layers {
		if l.MediaType != "application/vnd.gemara.artifact.v1+yaml" {
			continue
		}
		name := l.Annotations["org.opencontainers.image.title"]
		if name == "" || name == "." || name == ".." || name == "bundle.json" || strings.ContainsAny(name, `/\`) {
			return "malformed-manifest"
		}
		path := name
		if roles[i] == "import" {
			path = "imports/" + name
		} else {
			artifacts++
		}
		if _, dup := expected[path]; dup {
			return "malformed-manifest"
		}
		expected[path] = l
	}
	if artifacts != 1 {
		return "malformed-manifest"
	}

	// Step 5: the config bytes, then every layer file, match their descriptors.
	if config == nil {
		return "missing-layer"
	}
	if int64(len(config)) != m.Config.Size || sha256Digest(config) != m.Config.Digest {
		return "mismatch"
	}
	for path, d := range expected {
		info, err := os.Lstat(filepath.Join(unpacked, filepath.FromSlash(path)))
		if err != nil {
			return "missing-layer"
		}
		if !info.Mode().IsRegular() {
			return "mismatch"
		}
		data, err := os.ReadFile(filepath.Join(unpacked, filepath.FromSlash(path)))
		if err != nil || int64(len(data)) != d.Size || sha256Digest(data) != d.Digest {
			return "mismatch"
		}
	}

	// Step 6: nothing else is present.
	result := "ok"
	_ = filepath.WalkDir(unpacked, func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			result = "extra-layer"
			return err
		}
		if e.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(unpacked, p)
		if _, ok := expected[filepath.ToSlash(rel)]; !ok && rel != "bundle.json" {
			result = "extra-layer"
			return fs.SkipAll
		}
		return nil
	})
	return result
}

func sha256Digest(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
