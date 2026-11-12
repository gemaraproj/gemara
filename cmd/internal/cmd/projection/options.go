// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
)

type convertOpts struct {
	version string
	title   string
}

func defaultProjectionOpts(schemaDir string) convertOpts {
	return convertOpts{
		version: readVersion(schemaDir),
		title:   "Gemara",
	}
}

func (o *convertOpts) apply(options ...ConvertOption) {
	for _, opt := range options {
		opt(o)
	}
}

// ConvertOption configuring the OpenAPI projection
type ConvertOption func(options *convertOpts)

// WithTitle sets a new title on the API document
func WithTitle(title string) ConvertOption {
	return func(o *convertOpts) {
		o.title = title
	}
}

// WithVersion sets a new version on the API document
func WithVersion(version string) ConvertOption {
	return func(o *convertOpts) {
		o.version = version
	}
}

func readVersion(schemaDir string) string {
	path := filepath.Join(schemaDir, "VERSION")
	if data, err := os.ReadFile(path); err == nil {
		version := strings.TrimSpace(string(data))
		if version != "" {
			return version
		}
	}
	return "unknown"
}

// rootDescription returns the doc comment of the named definition, which
// becomes info.description. An empty name leaves the default in place.
//
// A --root naming a definition that does not exist or has no documentation
// comment is an error: an unknown path or undocumented definition would
// silently keep the default description, so `--root '#Metdata'` or an
// unannotated root would produce a plausible, wrong document rather than a
// complaint.
func rootDescription(v cue.Value, root string) (string, error) {
	if root == "" {
		return "", nil
	}
	rv := v.LookupPath(cue.ParsePath(root))
	if err := rv.Err(); err != nil {
		return "", fmt.Errorf("--root %q does not resolve in the CUE package: %w", root, err)
	}
	desc := strings.TrimSpace(docText(rv))
	if desc == "" {
		return "", fmt.Errorf("--root %q has no documentation comment", root)
	}
	return desc, nil
}
