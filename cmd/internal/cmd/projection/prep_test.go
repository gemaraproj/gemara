// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"strings"
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/encoding/openapi"
)

// mustPrepare parses src, runs prepare, and returns the formatted result.
func mustPrepare(t *testing.T, src string) (string, *prepInfo) {
	t.Helper()
	f, err := parser.ParseFile("test.cue", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	info, err := prepare(f)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	b, err := format.Node(f)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	return string(b), info
}

func TestPrepareStripsComprehensions(t *testing.T) {
	got, _ := mustPrepare(t, `package p
#Control: {
	state: string
	if state == "Retired" {
		recommendation?: _|_
	}
}
`)
	if strings.Contains(got, "if state") {
		t.Errorf("comprehension survived:\n%s", got)
	}
	if !strings.Contains(got, "state: string") {
		t.Errorf("sibling field was removed:\n%s", got)
	}
}

func TestPrepareUnwrapsAliasLabels(t *testing.T) {
	got, _ := mustPrepare(t, `package p
#Metadata: {
	MR="mapping-references"?: [string, ...string]
}
`)
	if strings.Contains(got, "MR=") {
		t.Errorf("alias survived:\n%s", got)
	}
	if !strings.Contains(got, `"mapping-references"?:`) {
		t.Errorf("underlying label lost:\n%s", got)
	}
}

// Stripping comprehensions orphans imports that only they used. CUE rejects
// an unused import, so prepare must drop it.
func TestPrepareDropsOrphanedImports(t *testing.T) {
	got, _ := mustPrepare(t, `package p

import "list"

#A: {
	xs: [...string]
	if len(xs) > 0 {
		ys: list.MinItems(1)
	}
}
`)
	if strings.Contains(got, `"list"`) {
		t.Errorf("orphaned import survived:\n%s", got)
	}
}

// An import still in use must be kept.
func TestPrepareKeepsUsedImports(t *testing.T) {
	got, _ := mustPrepare(t, `package p

import "list"

#A: {
	xs: list.MinItems(1)
}
`)
	if !strings.Contains(got, `"list"`) {
		t.Errorf("used import was dropped:\n%s", got)
	}
}

// The rewritten file must still build.
func TestPrepareOutputBuilds(t *testing.T) {
	f, err := parser.ParseFile("test.cue", `package p

import "list"

#Metadata: {
	MR="mapping-references"?: [string, ...string]
	if MR != _|_ {
		_n: list.MinItems(1)
	}
}
`, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, err = prepare(f)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	v := cuecontext.New().BuildFile(f)
	if err := v.Err(); err != nil {
		t.Fatalf("prepared file does not build: %v", err)
	}
}

// A dropped conditional must be recorded, not silently discarded: the caller
// reports these so a newly added conditional cannot vanish unnoticed.
func TestPrepareRecordsDroppedConditionals(t *testing.T) {
	_, info := mustPrepare(t, `package p
#Control: {
	state: string
	if state == "Retired" {
		recommendation?: _|_
	}
}
#Guideline: {
	state: string
	if state == "Retired" {
		recommendations?: _|_
	}
}
`)
	if len(info.DroppedConditionals) != 2 {
		t.Errorf("DroppedConditionals = %v, want 2 entries", info.DroppedConditionals)
	}
	for _, d := range info.DroppedConditionals {
		if !strings.Contains(d.pos, "test.cue:") {
			t.Errorf("position %q does not name the source file", d.pos)
		}
	}
}

// This reproduces the upstream panic the rewrite exists to avoid. If it ever
// stops panicking without the rewrite, the workaround can be deleted.
func TestUpstreamPanicsOnTimeFormatInList(t *testing.T) {
	f, err := parser.ParseFile("test.cue", `package p

import "time"

#DT: time.Format("2006-01-02T15:04:05Z07:00")
#B:  {list?: [#DT, ...#DT]}
`, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	v := cuecontext.New().BuildFile(f)
	if err := v.Err(); err != nil {
		t.Fatalf("build: %v", err)
	}

	panicked := func() (p bool) {
		defer func() {
			if recover() != nil {
				p = true
			}
		}()
		_, _ = openapi.Generate(v, &openapi.Config{
			Info:    map[string]string{"title": "t", "version": "v"},
			Version: "3.0.0",
		})
		return false
	}()
	if !panicked {
		t.Skip("upstream fixed the panic; the time.Format rewrite in prepare() can be removed")
	}
}

func TestPrepareRewritesTimeFormat(t *testing.T) {
	got, _ := mustPrepare(t, `package p

import "time"

#Datetime: time.Format("2006-01-02T15:04:05Z07:00") @gemara(format="date-time")
`)
	if strings.Contains(got, "time.Format") {
		t.Errorf("time.Format survived:\n%s", got)
	}
	if !strings.Contains(got, "#Datetime: string @gemara(format=\"date-time\")") {
		t.Errorf("expected #Datetime to become string:\n%s", got)
	}
}

// The rewrite must not fire on an unrelated call.
func TestPrepareLeavesOtherCallsAlone(t *testing.T) {
	got, _ := mustPrepare(t, `package p

import "strings"

#A: {s: strings.MinRunes(3)}
`)
	if !strings.Contains(got, "strings.MinRunes") {
		t.Errorf("unrelated call was rewritten:\n%s", got)
	}
}

// A time.Format on a field rather than as a whole definition body would be
// rewritten to a bare string with its layout attributed to the enclosing
// definition — reintroducing the exact silent degradation of gemara#466.
func TestPrepareRejectsFieldLevelTimeFormat(t *testing.T) {
	f, err := parser.ParseFile("test.cue", `package p

import "time"

#A: {
	ts: time.Format("2006-01-02")
}
`, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := prepare(f); err == nil {
		t.Error("expected an error for a field-level time.Format, got nil")
	}
}

// A time.Format outside any definition has nowhere to record its layout, so it
// must be an error rather than a silent rewrite to a bare string.
func TestPrepareRejectsTimeFormatOutsideDefinition(t *testing.T) {
	f, err := parser.ParseFile("test.cue", `package p

import "time"

plain: time.Format("2006-01-02")
`, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := prepare(f); err == nil {
		t.Error("expected an error for time.Format outside a definition, got nil")
	}
}

// Without @gemara(format=...), the rewrite would leave a bare string and the
// layout would be lost from the document.
func TestPrepareRejectsTimeFormatWithoutFormatDirective(t *testing.T) {
	f, err := parser.ParseFile("test.cue", `package p

import "time"

#Date: time.Format("2006-01-02")
`, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, err = prepare(f)
	if err == nil || !strings.Contains(err.Error(), "#Date") || !strings.Contains(err.Error(), "@gemara(format=") {
		t.Fatalf("err = %v, want an error naming #Date and the missing format directive", err)
	}
}
