// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/literal"
)

// prepInfo records what prepare rewrote, so the post-pass can restore the
// information in a form OpenAPI can express.
type prepInfo struct {
	// TimeFormats maps a definition name (no leading '#') to the Go layout
	// string from the time.Format call that was rewritten away.
	TimeFormats map[string]string
	// DroppedConditionals records the source position of every cross-field
	// conditional removed by stripComprehensions. OpenAPI cannot express
	// these, so dropping them is unavoidable — but the caller reports them
	// so a newly added conditional announces that it will not reach the
	// generated document, rather than vanishing silently.
	DroppedConditionals []string
}

// prepare rewrites a CUE file into the subset cuelang.org/go/encoding/openapi
// accepts. Every rewrite here is a rewrite of the *input*; none of them is
// allowed to change what the generated schema means.
//
// Order matters: comprehensions go first, because removing them is what
// orphans the aliases and imports the later steps clean up.
func prepare(f *ast.File) (*prepInfo, error) {
	info := &prepInfo{TimeFormats: map[string]string{}}
	var err error

	f.Decls = stripComprehensions(f.Decls, info)
	astutil.Apply(f, func(c astutil.Cursor) bool {
		if err != nil {
			return false
		}
		switch n := c.Node().(type) {
		case *ast.StructLit:
			// OpenAPI cannot express a cross-field conditional, and the
			// encoder errors rather than guessing. The walker this replaces
			// ignored these too, so no fidelity is lost — but it is a
			// deliberate omission, recorded here rather than incidental.
			n.Elts = stripComprehensions(n.Elts, info)
		case *ast.Field:
			// X="name": the alias exists to be read by the comprehensions
			// just removed. CUE rejects an unreferenced alias, so unwrap it
			// and keep the real label. This is also the fix gemara#473 asks
			// for: the walker returned "" for an *ast.Alias label and
			// dropped the field.
			if al, ok := n.Label.(*ast.Alias); ok {
				if lbl, ok := al.Expr.(ast.Label); ok {
					n.Label = lbl
				}
			}
		case *ast.CallExpr:
			// WORKAROUND (cuelang.org/go, all versions through v0.18.0-alpha.1):
			// encoding/openapi panics on a time.Format() reachable inside a
			// list — build.go:432 type-asserts v.Syntax(cue.Concrete(true))
			// to ast.Expr without checking, and gets an *ast.File:
			//
			//   panic: interface conversion: *ast.File is not ast.Expr
			//
			// #Datetime is reachable from a list in #AuditResult,
			// #ControlEvaluation, #AssessmentLog and #ActionResult, so the
			// schema cannot be generated at all without this. Rewrite the
			// call to `string` and record the layout; applyTimeFormats
			// restores `format` on the way out.
			//
			// Upstream issue not yet filed. When it is fixed, delete this
			// case and applyTimeFormats — TestUpstreamPanicsOnTimeFormatInList
			// will skip, which is the signal.
			if layout, ok := timeFormatLayout(n); ok {
				name, whole, found := enclosingDefinition(c)
				if !found {
					err = fmt.Errorf("%s: time.Format(%q) is not inside a definition; "+
						"cue2openapi cannot record its format and would emit a bare string",
						n.Pos(), layout)
					return false
				}
				// The layout is recorded against the definition name and
				// restored onto that definition's schema, so the call must BE
				// the definition's value. On an inner field the rewrite would
				// turn the field into a bare string with no format and stamp
				// the format onto the enclosing object schema instead —
				// gemara#466, silently, which is the whole point of this work.
				if !whole {
					err = fmt.Errorf("%s: time.Format(%q) is only supported as a whole-definition "+
						"body (e.g. `#Datetime: time.Format(...)`); on a field it would be emitted "+
						"as a bare string with no format, and its layout attributed to the "+
						"enclosing definition #%s",
						n.Pos(), layout, name)
					return false
				}
				// TimeFormats is keyed by definition name, so two differing
				// layouts under one name would overwrite each other.
				if prev, ok := info.TimeFormats[name]; ok && prev != layout {
					err = fmt.Errorf("%s: #%s has two different time.Format layouts (%q and %q); "+
						"only one format can be recorded per definition", n.Pos(), name, prev, layout)
					return false
				}
				info.TimeFormats[name] = layout
				c.Replace(ast.NewIdent("string"))
				return false
			}
		}
		return true
	}, nil)
	if err != nil {
		return nil, err
	}
	pruneImports(f)

	return info, nil
}

// stripComprehensions removes *ast.Comprehension entries from a declaration
// list. astutil's Cursor.Delete does not support this position, so the slice
// is rebuilt directly.
func stripComprehensions(decls []ast.Decl, info *prepInfo) []ast.Decl {
	out := decls[:0:0]
	for _, d := range decls {
		if _, ok := d.(*ast.Comprehension); ok {
			info.DroppedConditionals = append(info.DroppedConditionals, d.Pos().String())
			continue
		}
		out = append(out, d)
	}
	return out
}

// pruneImports drops imports left unused by prepare's rewrites. CUE rejects a
// file with an unused import, so this is required, not cosmetic.
func pruneImports(f *ast.File) {
	used := map[string]bool{}
	ast.Walk(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	}, nil)

	out := f.Decls[:0:0]
	for _, d := range f.Decls {
		imp, ok := d.(*ast.ImportDecl)
		if !ok {
			out = append(out, d)
			continue
		}
		specs := imp.Specs[:0:0]
		for _, sp := range imp.Specs {
			if used[importName(sp)] {
				specs = append(specs, sp)
			}
		}
		if len(specs) == 0 {
			continue
		}
		imp.Specs = specs
		out = append(out, d)
	}
	f.Decls = out
}

// importName returns the identifier an import is referenced by.
func importName(sp *ast.ImportSpec) string {
	if sp.Name != nil {
		return sp.Name.Name
	}
	path := strings.Trim(sp.Path.Value, `"`)
	if i := strings.LastIndexAny(path, "/:"); i >= 0 {
		path = path[i+1:]
	}
	return path
}

// timeFormatLayout reports whether call is time.Format("layout"), and the
// unquoted layout if so.
func timeFormatLayout(call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || len(call.Args) != 1 {
		return "", false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "time" {
		return "", false
	}
	id, ok := sel.Sel.(*ast.Ident)
	if !ok || id.Name != "Format" {
		return "", false
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return "", false
	}
	layout, err := literal.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return layout, true
}

// enclosingDefinition walks up from the cursor to the nearest definition field
// and returns its name without the leading '#'.
//
// whole reports whether that definition field is also the *nearest* field
// ancestor — that is, whether the cursor's node is the definition's own value
// rather than sitting on some field inside it. Callers that attribute
// information to the definition name must require it: crossing an inner field
// on the way up means the information belongs to that field, not the
// definition.
func enclosingDefinition(c astutil.Cursor) (name string, whole, found bool) {
	whole = true
	for p := c; p != nil; p = p.Parent() {
		fld, ok := p.Node().(*ast.Field)
		if !ok {
			continue
		}
		if id, ok := fld.Label.(*ast.Ident); ok && strings.HasPrefix(id.Name, "#") {
			return strings.TrimPrefix(id.Name, "#"), whole, true
		}
		// A field that is not the definition: anything found above it encloses
		// that field rather than holding the cursor's own value.
		whole = false
	}
	return "", false, false
}
