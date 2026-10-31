// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/getkin/kin-openapi/openapi3"
)

// Conditionals in a hidden helper are translated when they have a supported
// shape, and dropped and reported like any other conditional when they don't.
func TestPrepareTranslatesHelperConditionals(t *testing.T) {
	tests := []struct {
		name        string
		conditional string
		want        *helperRule // nil: dropped and reported
	}{
		{
			name:        "absent subject",
			conditional: `if a == _|_ { f: #T }`,
			want:        &helperRule{subject: "a", absent: true, field: "f", def: "T"},
		},
		{
			name:        "excluded values",
			conditional: `if k != "x" && k != "y" { f: #T }`,
			want:        &helperRule{subject: "k", values: []string{"x", "y"}, field: "f", def: "T"},
		},
		{
			name:        "non-empty list body",
			conditional: `if k != "x" { l: [#T, ...#T] }`,
			want:        &helperRule{subject: "k", values: []string{"x"}, field: "l", def: "T", list: true},
		},
		{name: "equality to a value", conditional: `if k == "x" { f: #T }`},
		{name: "bytes literal", conditional: `if k != 'x' { f: #T }`},
		{name: "optional body field", conditional: `if k != "x" { f?: #T }`},
		{name: "hidden body field", conditional: `if k != "x" { _f: #T }`},
		{name: "two body fields", conditional: `if k != "x" { f: #T, l: [#T, ...#T] }`},
		{name: "narrowing body", conditional: `if k != "x" { f: "fixed" }`},
		{name: "two subjects", conditional: `if k != "x" && a != "y" { f: #T }`},
		{name: "present subject", conditional: `if a != _|_ { f: #T }`},
		{name: "or chain", conditional: `if k != "x" || k != "y" { f: #T }`},
		{name: "mismatched list tail", conditional: `if k != "x" { l: [#T, ...#U] }`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, info := mustPrepare(t, `package p
#T: string
#U: string
#B: {k: string, a?: string, f?: #T, l?: [#T, ...#T]}
#_BStrict: {
	@go(-)
} & #B & {
	k:  string
	a?: string
	`+tt.conditional+`
}
`)
			rules := info.HelperRules["_BStrict"]
			if tt.want == nil {
				if len(rules) != 0 || len(info.DroppedConditionals) != 1 {
					t.Fatalf("want the conditional dropped and reported; rules = %+v, dropped = %v", rules, info.DroppedConditionals)
				}
				return
			}
			if len(rules) != 1 || len(info.DroppedConditionals) != 0 {
				t.Fatalf("want one rule and nothing dropped; rules = %+v, dropped = %v", rules, info.DroppedConditionals)
			}
			got := rules[0]
			got.pos = ""
			if !reflect.DeepEqual(got, *tt.want) {
				t.Errorf("rule = %+v, want %+v", got, *tt.want)
			}
		})
	}
}

// A condition must refer to the helper's own field. CUE resolves it lexically,
// so when the helper block does not declare the field under that identifier,
// or the conditional sits outside the helper's own conjuncts, it is dropped.
func TestPrepareDropsHelperConditionalsOutsideTheHelperScope(t *testing.T) {
	tests := map[string]string{
		"undeclared subject": `
	if k != "x" { f: #T }`,
		"quoted subject label": `
	"k": string
	if k != "x" { f: #T }`,
		"inside a let": `
	let extra = {
		k: string
		if k != "x" { f: #T }
	}
	_x: extra`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			_, info := mustPrepare(t, `package p
#T: string
k: "outside"
#B: {k: string, f?: #T}
#_BStrict: {
	@go(-)
} & #B & {`+body+`
}
`)
			if len(info.HelperRules) != 0 || len(info.DroppedConditionals) != 1 {
				t.Fatalf("rules = %+v, dropped = %v; want the conditional dropped", info.HelperRules, info.DroppedConditionals)
			}
		})
	}
}

// Only a hidden helper's own conditionals are translated; the same shape in an
// ordinary definition still has nowhere to go and is reported as dropped.
func TestPrepareLeavesOrdinaryConditionalsDropped(t *testing.T) {
	_, info := mustPrepare(t, `package p
#T: string
#B: {
	k: string
	f?: #T
	if k != "x" { f: #T }
}
`)
	if len(info.HelperRules) != 0 || len(info.DroppedConditionals) != 1 {
		t.Fatalf("rules = %+v, dropped = %v; want the conditional dropped", info.HelperRules, info.DroppedConditionals)
	}
}

const helperRuleSchema = `@gemara(status="stable")

package p

#Datetime: string
#Ref: {id: string}
#Result: "Passed" | "Failed" | "Not Run" | "Unknown"

#Log: {
	result:   #Result
	start?:   #Datetime
	payload?: _
	source?:  #Ref
}

#_LogStrict: {
	@go(-)
} & #Log & {
	result:   #Result
	payload?: _
	if result != "Not Run" && result != "Unknown" {
		start: #Datetime
	}
	if payload == _|_ {
		source: #Ref
	}
}

#Kind: "link" | "none"

#Pair: {
	kind:     #Kind
	targets?: [#Ref, ...#Ref]
}

#_PairStrict: {
	@go(-)
} & #Pair & {
	kind: #Kind
	if kind != "none" {
		targets: [#Ref, ...#Ref]
	}
}

#Doc: {
	logs?:  [#_LogStrict, ...#_LogStrict]
	pairs?: [#_PairStrict, ...#_PairStrict]
}
`

// The projected rules must accept exactly what the CUE accepts: each entry is
// checked against both, on its own and inside a document, so a mismatch in
// either direction fails. On its own, the base schema must agree with the
// helper, because the helper's rules are the base type's real rules.
func TestHelperRulesMatchTheCUE(t *testing.T) {
	doc, err := convertSource(t, helperRuleSchema)
	if err != nil {
		t.Fatal(err)
	}
	schemas := reloaded(t, doc).Components.Schemas
	ctx := cuecontext.New()
	v := ctx.CompileString(helperRuleSchema)
	if err := v.Err(); err != nil {
		t.Fatal(err)
	}
	agree := func(t *testing.T, schema string, def cue.Value, value any, valid bool) {
		t.Helper()
		cueErr := def.Unify(ctx.Encode(value)).Validate(cue.Concrete(true))
		if (cueErr == nil) != valid {
			t.Fatalf("CUE valid = %v, want %v (%v); the case itself is wrong", cueErr == nil, valid, cueErr)
		}
		s := schemas[schema]
		if s == nil {
			t.Fatalf("%s is missing from the generated document", schema)
		}
		if oaErr := s.Value.VisitJSON(value); (oaErr == nil) != valid {
			t.Errorf("%s: OpenAPI valid = %v, CUE valid = %v:\n%v", schema, oaErr == nil, valid, oaErr)
		}
	}
	uses := map[string]struct{ helper, list string }{
		"Log":  {"#_LogStrict", "logs"},
		"Pair": {"#_PairStrict", "pairs"},
	}

	tests := []struct {
		name  string
		base  string
		entry map[string]any
		valid bool
	}{
		{"executed with start and source", "Log", map[string]any{"result": "Passed", "start": "t", "source": map[string]any{"id": "s"}}, true},
		{"executed without start", "Log", map[string]any{"result": "Passed", "source": map[string]any{"id": "s"}}, false},
		{"not run without start", "Log", map[string]any{"result": "Not Run", "source": map[string]any{"id": "s"}}, true},
		{"unknown without start", "Log", map[string]any{"result": "Unknown", "source": map[string]any{"id": "s"}}, true},
		{"failed without start", "Log", map[string]any{"result": "Failed", "source": map[string]any{"id": "s"}}, false},
		{"neither payload nor source", "Log", map[string]any{"result": "Not Run"}, false},
		{"payload without source", "Log", map[string]any{"result": "Not Run", "payload": "p"}, true},
		{"link without targets", "Pair", map[string]any{"kind": "link"}, false},
		{"link with targets", "Pair", map[string]any{"kind": "link", "targets": []any{map[string]any{"id": "t"}}}, true},
		{"none without targets", "Pair", map[string]any{"kind": "none"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			use := uses[tt.base]
			agree(t, tt.base, hiddenDef(t, v, use.helper), tt.entry, tt.valid)
			agree(t, "Doc", v.LookupPath(cue.ParsePath("#Doc")), map[string]any{use.list: []any{tt.entry}}, tt.valid)
		})
	}
}

// A rule that cannot be projected faithfully is left out and reported, and
// generation carries on, as it did when every conditional was dropped.
func TestHelperRulesSkipUnfaithfulProjections(t *testing.T) {
	tests := []struct {
		name, base, helper, wantReason string
	}{
		{
			name:       "optional subject compared by value",
			base:       `k?: #K, f?: #T`,
			helper:     `k?: #K` + "\n\t" + `if k != "x" { f: #T }`,
			wantReason: "k is optional in #B",
		},
		{
			name:       "subject with a default",
			base:       `k: *"x" | "y" @gemara(default="x"), f?: #T`,
			helper:     `k: *"x" | "y"` + "\n\t" + `if k != "y" { f: #T }`,
			wantReason: "k has a default in #B",
		},
		{
			name:       "body narrows the base field",
			base:       `k: #K, f?: string`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { f: #T }`,
			wantReason: "#B declares f differently",
		},
		{
			name:       "nullable base field",
			base:       `k: #K, f?: #T | null`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { f: #T }`,
			wantReason: "#B declares f differently",
		},
		{
			name:       "field the base does not declare",
			base:       `k: #K, g?: string`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { f: #T }`,
			wantReason: "#B does not declare f",
		},
		{
			name:       "body type with a default",
			base:       `k: #K, f?: #D`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { f: #D }`,
			wantReason: "#D is satisfied without data",
		},
		{
			name:       "body type with only optional fields",
			base:       `k: #K, f?: #S`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { f: #S }`,
			wantReason: "#S is satisfied without data",
		},
		{
			name:       "required field with a default",
			base:       `k: #K, f?: *"d" | #E @gemara(default="d")`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { f: #E }`,
			wantReason: "f has a default in #B",
		},
		{
			name:       "base list may be empty",
			base:       `k: #K, l?: [...#T]`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { l: [#T, ...#T] }`,
			wantReason: "#B declares l differently",
		},
		{
			name:       "nullable base list",
			base:       `k: #K, l?: [#T, ...#T] | null`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { l: [#T, ...#T] }`,
			wantReason: "#B declares l differently",
		},
		{
			name:       "base field is not a list",
			base:       `k: #K, l?: #T`,
			helper:     `k: #K` + "\n\t" + `if k != "x" { l: [#T, ...#T] }`,
			wantReason: "#B declares l differently",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := `@gemara(status="stable")

package p

#T: {id: string}
#K: "x" | "y"
#D: *"d" | "e"
#E: "d" | "e"
#S: {note?: string}
#B: {` + tt.base + `}
#_BStrict: {
	@go(-)
} & #B & {
	` + tt.helper + `
}
#Doc: {bs?: [#_BStrict, ...#_BStrict]}
`
			skipped := skippedRules(t, src)
			if len(skipped) != 1 || !strings.Contains(skipped[0], tt.wantReason) {
				t.Fatalf("skipped = %q, want one note containing %q", skipped, tt.wantReason)
			}
			requireNoRules(t, src, "B")
		})
	}
}

// Every helper of a base collapses onto it, so with two helpers a rule on the
// base would also bind the other helper's uses. Neither rule is projected.
func TestHelperRulesSkipASharedBase(t *testing.T) {
	src := `@gemara(status="stable")

package p

#T: {id: string}
#K: "x" | "y"
#B: {k: #K, f?: #T, g?: #T}
#_BStrict: {
	@go(-)
} & #B & {
	k: #K
	if k != "x" { f: #T }
}
#_BOther: {
	@go(-)
} & #B & {
	k: #K
	if k != "y" { g: #T }
}
#Doc: {
	strict?: [#_BStrict, ...#_BStrict]
	other?: [#_BOther, ...#_BOther]
}
`
	skipped := skippedRules(t, src)
	if len(skipped) != 2 {
		t.Fatalf("skipped = %q, want both rules", skipped)
	}
	for _, s := range skipped {
		if !strings.Contains(s, "#B is extended by more than one helper") {
			t.Errorf("skipped %q for the wrong reason", s)
		}
	}
	requireNoRules(t, src, "B")
}

// The repository's bad fixtures for the three strict helpers must now be
// rejected by the generated document too, and only for the rule's field. The
// entry that breaks the rule is rejected on its own as well, by its base
// schema.
func TestStrictHelperBadFixturesFailTheDocument(t *testing.T) {
	doc := generatedDocument(t)
	// A new required field on #AssessmentLog, #Evidence or #Mapping also has to
	// be added to these fixtures, or they fail here for the wrong reason.
	tests := []struct {
		file, schema, base string
		entry              func(*testing.T, map[string]any) map[string]any // the entry that breaks the rule
		fix                func(map[string]any)                            // adds what the rule asks for
	}{
		{
			"test/test-data/bad-evaluation-log-missing-start.yaml", "EvaluationLog", "AssessmentLog",
			func(t *testing.T, d map[string]any) map[string]any {
				return first(t, first(t, d["evaluations"], "evaluations")["assessment-logs"], "assessment-logs")
			},
			func(e map[string]any) { e["start"] = "2025-08-22T16:01:00Z" },
		},
		{
			"test/test-data/bad-audit-log-evidence-neither.yaml", "AuditLog", "Evidence",
			func(t *testing.T, d map[string]any) map[string]any {
				for _, r := range sliceOf(d["results"]) {
					for _, e := range sliceOf(mapOf(r)["evidence"]) {
						if ev := mapOf(e); ev["payload"] == nil && ev["source"] == nil {
							return ev
						}
					}
				}
				t.Fatal("the fixture has no evidence without payload and source")
				return nil
			},
			func(e map[string]any) { e["payload"] = "inline evidence" },
		},
		{
			"test/test-data/bad-mapping-no-target.yaml", "MappingDocument", "Mapping",
			func(t *testing.T, d map[string]any) map[string]any { return first(t, d["mappings"], "mappings") },
			func(e map[string]any) { e["targets"] = []any{map[string]any{"entry-id": "T1"}} },
		},
	}
	for _, tt := range tests {
		t.Run(filepath.Base(tt.file), func(t *testing.T) {
			value := jsonValue(t, filepath.Join("../../../..", tt.file))
			entry := tt.entry(t, mapOf(value))
			checks := []struct {
				schema string
				value  any
			}{{tt.schema, value}, {tt.base, entry}}
			for _, c := range checks {
				if err := doc.Components.Schemas[c.schema].Value.VisitJSON(c.value); err == nil {
					t.Errorf("the CUE rejects this but %s accepts it", c.schema)
				}
			}
			tt.fix(entry)
			for _, c := range checks {
				if err := doc.Components.Schemas[c.schema].Value.VisitJSON(c.value); err != nil {
					t.Errorf("after supplying the rule's field %s still rejects it, so it failed for another reason:\n%v", c.schema, err)
				}
			}
		})
	}
}

// The start rule must still accept the logs it exempts. The good fixture for
// it has Not Run, Unknown and Not Applicable logs without start, but it is in
// knownFixtureGaps because the CUE fills each log's requirement.reference-id
// from its control, so this supplies that the same way before checking.
func TestStrictHelperGoodFixturePassesTheDocument(t *testing.T) {
	doc := generatedDocument(t)
	value := mapOf(jsonValue(t, "../../../../test/test-data/good-evaluation-log-unstarted.yaml"))
	for _, e := range sliceOf(value["evaluations"]) {
		evaluation := mapOf(e)
		for _, l := range sliceOf(evaluation["assessment-logs"]) {
			log := mapOf(l)
			mapOf(log["requirement"])["reference-id"] = mapOf(evaluation["control"])["reference-id"]
			if err := doc.Components.Schemas["AssessmentLog"].Value.VisitJSON(log); err != nil {
				t.Errorf("AssessmentLog rejects the %v log:\n%v", log["result"], err)
			}
		}
	}
	if err := doc.Components.Schemas["EvaluationLog"].Value.VisitJSON(value); err != nil {
		t.Errorf("EvaluationLog rejects the fixture:\n%v", err)
	}
}

// generatedDocument generates the repository's schema and loads it the way a
// consumer of openapi.yaml would.
func generatedDocument(t *testing.T) *openapi3.T {
	t.Helper()
	out := filepath.Join(t.TempDir(), "projection.yaml")
	writeGenerated(t, out, "", "")
	doc, err := openapi3.NewLoader().LoadFromFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

// skippedRules runs the helper-rule pass alone on the document generated from
// src, and returns its notes on the rules it left out.
func skippedRules(t *testing.T, src string) []string {
	t.Helper()
	v, info, _, err := loadPrepared(writeSourceModule(t, src))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := generateRaw(v, "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	skipped, err := applyHelperRulesTyped(raw, info.HelperRules, v)
	if err != nil {
		t.Fatal(err)
	}
	return skipped
}

// requireNoRules fails unless generation succeeds on src and leaves schema
// without a projected rule.
func requireNoRules(t *testing.T, src, schema string) {
	t.Helper()
	doc, err := convertSource(t, src)
	if err != nil {
		t.Fatalf("generation must not fail on a rule it cannot project: %v", err)
	}
	for _, branch := range doc.Components.Schemas[schema].Value.AllOf {
		if branch.Value != nil && len(branch.Value.AnyOf) > 0 {
			t.Errorf("a skipped rule was projected onto %s anyway", schema)
		}
	}
}

// hiddenDef returns the hidden definition name in v; cue.ParsePath does not
// accept hidden labels.
func hiddenDef(t *testing.T, v cue.Value, name string) cue.Value {
	t.Helper()
	iter, err := v.Fields(cue.Hidden(true), cue.Definitions(true))
	if err != nil {
		t.Fatal(err)
	}
	for iter.Next() {
		if iter.Selector().String() == name {
			return iter.Value()
		}
	}
	t.Fatalf("%s is not defined", name)
	return cue.Value{}
}

// first returns the first element of a fixture list as a map.
func first(t *testing.T, list any, name string) map[string]any {
	t.Helper()
	items := sliceOf(list)
	if len(items) == 0 {
		t.Fatalf("the fixture has no %s; it no longer has the shape this test expects", name)
	}
	return mapOf(items[0])
}

// writeSourceModule writes src as the only file of a throwaway CUE module and
// returns its directory.
func writeSourceModule(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "cue.mod"), 0o755); err != nil {
		t.Fatal(err)
	}
	module := "module: \"example.com/p@v0\"\nlanguage: version: \"v0.15.1\"\n"
	if err := os.WriteFile(filepath.Join(dir, "cue.mod", "module.cue"), []byte(module), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.cue"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// reloaded round-trips doc through the loader so every reference is resolved
// the same way a consumer of the written file would see it.
func reloaded(t *testing.T, doc *openapi3.T) *openapi3.T {
	t.Helper()
	data, err := doc.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	out, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
