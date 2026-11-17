// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"reflect"
	"testing"
)

func TestPreparePreservesOnlySupportedPresenceRules(t *testing.T) {
	_, info := mustPrepare(t, `package p
#Link: {
 relationship: *"relates-to" | string
 entries?: [string, ...string]
 if relationship != "no-match" {entries: [string, ...string]}
}
#Narrower: {
 relationship: string
 entries?: [...string]
 if relationship != "no-match" {entries: [string, ...string]}
}
#Nested: {
 inner: {
  relationship: string
  entries?: [...string]
  if relationship != "no-match" {entries: [...string]}
 }
}
`)
	want := []requiredUnless{{"Link", "entries", "relationship", "no-match"}}
	if !reflect.DeepEqual(info.RequiredUnless, want) {
		t.Fatalf("rules = %#v, want %#v", info.RequiredUnless, want)
	}
	if len(info.DroppedConditionals) != 2 {
		t.Fatalf("unsupported rules must remain reported: %v", info.DroppedConditionals)
	}
}

func TestRequiredUnlessUsesSelectorDefault(t *testing.T) {
	doc, err := convertSource(t, `@gemara(status="experimental")
package p
#Link: {
 relationship: *"no-match" | "relates-to" @gemara(default="no-match")
 entries?: [string, ...string]
 if relationship != "no-match" {entries: [string, ...string]}
}
`)
	if err != nil {
		t.Fatal(err)
	}
	s := doc.Components.Schemas["Link"].Value
	if err := s.VisitJSON(map[string]any{}); err != nil {
		t.Fatalf("defaulted exception should allow omitted entries: %v", err)
	}
	if err := s.VisitJSON(map[string]any{"relationship": "relates-to"}); err == nil {
		t.Fatal("non-exception relationship accepted without entries")
	}
}
