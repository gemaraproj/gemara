// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/literal"
	"cuelang.org/go/cue/token"
)

// Prep loads the package and applies required upstream-generator
// workarounds while recording every omitted cross-field conditional.
func (c *Converter) Prep() error {
	if c.schemaDir == "" {
		return fmt.Errorf("load a CUE package before preprocessing")
	}
	v, prep, inst, err := loadPrepared(c.schemaDir)
	if err != nil {
		return err
	}
	c.value, c.prep, c.instance = v, prep, inst
	if len(prep.DroppedConditionals) > 0 {
		fmt.Fprintf(os.Stderr, "cue2openapi: %d cross-field conditional(s) were omitted because the CUE OpenAPI projector cannot translate them:\n", len(prep.DroppedConditionals))
		for _, pos := range prep.DroppedConditionals {
			fmt.Fprintf(os.Stderr, "  %s\n", pos)
		}
	}
	return nil
}

// prepInfo records what prepare rewrote, so the post-pass can restore the
// information in a form OpenAPI can express.
type prepInfo struct {
	// CUE CONVERTER WORKAROUND: DroppedConditionals records the source position
	// of every cross-field conditional removed by stripComprehensions. The
	// current CUE OpenAPI projector cannot translate these, so the caller reports their omission
	// so a newly added conditional announces that it will not reach the
	// generated document, rather than vanishing silently.
	DroppedConditionals []string

	// CUE CONVERTER WORKAROUND: HelperRules holds the conditionals of hidden #_
	// helpers (such as #_AssessmentLogStrict) that stripComprehensions
	// translated instead of dropping, keyed by helper name without the leading
	// '#'. Post attaches each rule to the base schema the helper extends, which
	// every use of the helper is collapsed onto.
	HelperRules map[string][]helperRule
}

// prepare rewrites a CUE file into the subset cuelang.org/go/encoding/openapi
// accepts. Every rewrite here is a rewrite of the *input*; none of them is
// allowed to change what the generated schema means.
//
// Order matters: comprehensions go first, because removing them is what
// orphans the aliases and imports the later steps clean up.
func prepare(f *ast.File) (*prepInfo, error) {
	info := &prepInfo{HelperRules: map[string][]helperRule{}}
	var err error

	f.Decls = stripComprehensions(f.Decls, info, "")
	astutil.Apply(f, func(c astutil.Cursor) bool {
		if err != nil {
			return false
		}
		switch n := c.Node().(type) {
		case *ast.StructLit:
			// CUE CONVERTER WORKAROUND: The current CUE OpenAPI projector cannot
			// translate a cross-field conditional and errors rather than guessing. The walker this replaces
			// ignored these too, so no fidelity is lost — but it is a
			// deliberate omission, recorded here rather than incidental.
			// Conditionals in a hidden #_ helper are translated where
			// their shape allows, so Post can restore them.
			n.Elts = stripComprehensions(n.Elts, info, hiddenHelperName(c))
		case *ast.Field:
			// CUE CONVERTER WORKAROUND: X="name": the alias exists to be read by the comprehensions
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
			// CUE CONVERTER WORKAROUND (cuelang.org/go, all versions through v0.18.0-alpha.1):
			// encoding/openapi panics on a time.Format() reachable inside a
			// list — build.go:432 type-asserts v.Syntax(cue.Concrete(true))
			// to ast.Expr without checking, and gets an *ast.File:
			//
			//   panic: interface conversion: *ast.File is not ast.Expr
			//
			// #Datetime is reachable from a list in #AuditResult,
			// #ControlEvaluation, #AssessmentLog and #ActionResult, so the
			// schema cannot be generated at all without this. Rewrite the
			// call to `string`; @gemara(format="...") declares the OpenAPI
			// format applied after generation.
			//
			// Upstream issue not yet filed. When it is fixed, delete this
			// case — TestUpstreamPanicsOnTimeFormatInList
			// will skip, which is the signal.
			if layout, ok := timeFormatLayout(n); ok {
				name, def, whole, found := enclosingDefinition(c)
				if !found {
					err = fmt.Errorf("%s: time.Format(%q) is not inside a definition; "+
						"cue2openapi cannot record its format and would emit a bare string",
						n.Pos(), layout)
					return false
				}
				// The OpenAPI format comes from @gemara(format="...") on the
				// definition, so the call must be that definition's value. An
				// inner field would otherwise be rewritten to a bare string.
				if !whole {
					err = fmt.Errorf("%s: time.Format(%q) is only supported as a whole-definition "+
						"body (e.g. `#Datetime: time.Format(...)`); on a field it would be emitted "+
						"as a bare string with no format, and its layout attributed to the "+
						"enclosing definition #%s",
						n.Pos(), layout, name)
					return false
				}
				// Without the directive the rewrite leaves a bare string and
				// the layout is lost from the document.
				if !declaresFormat(def) {
					err = fmt.Errorf("%s: #%s is time.Format(%q) but has no @gemara(format=\"...\"); "+
						"cue2openapi rewrites it to string and would emit no format",
						n.Pos(), name, layout)
					return false
				}
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
// is rebuilt directly. Inside a hidden helper (helper is non-empty), a
// conditional that translateHelperConditional recognises is recorded as a
// helper rule instead of as dropped.
func stripComprehensions(decls []ast.Decl, info *prepInfo, helper string) []ast.Decl {
	out := decls[:0:0]
	for _, d := range decls {
		c, ok := d.(*ast.Comprehension)
		if !ok {
			out = append(out, d)
			continue
		}
		if helper != "" {
			// CUE resolves the condition's field lexically, so it only means
			// the helper's own field when this block declares it, as the
			// helpers do (`result: #Result`); otherwise it could bind to
			// something outside the helper.
			if rule, ok := translateHelperConditional(c); ok && declaresField(decls, rule.subject) {
				info.HelperRules[helper] = append(info.HelperRules[helper], rule)
				continue
			}
		}
		info.DroppedConditionals = append(info.DroppedConditionals, d.Pos().String())
	}
	return out
}

// declaresField reports whether decls declares a field that an unquoted
// reference to name resolves to. A quoted label does not bind the identifier.
func declaresField(decls []ast.Decl, name string) bool {
	for _, d := range decls {
		if f, ok := d.(*ast.Field); ok {
			if label, isIdent, err := ast.LabelName(f.Label); err == nil && isIdent && label == name {
				return true
			}
		}
	}
	return false
}

// hiddenHelperName returns the name (without '#') of the hidden #_ helper the
// cursor's struct is a conjunct of, as in `#_X: {@go(-)} & #Base & {...}`, or
// "" when the struct is anywhere else: inside one of the helper's fields, a
// let, a disjunction or a list.
func hiddenHelperName(c astutil.Cursor) string {
	for p := c.Parent(); p != nil; p = p.Parent() {
		switch n := p.Node().(type) {
		case *ast.BinaryExpr:
			if n.Op != token.AND {
				return ""
			}
		case *ast.ParenExpr:
		case *ast.Field:
			if id, ok := n.Label.(*ast.Ident); ok && strings.HasPrefix(id.Name, "#_") {
				return strings.TrimPrefix(id.Name, "#")
			}
			return ""
		default:
			return ""
		}
	}
	return ""
}

// CUE CONVERTER WORKAROUND: pruneImports drops imports left unused by prepare's
// rewrites. CUE rejects a file with an unused import, so this is required, not cosmetic.
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
// and returns it along with its name without the leading '#'.
//
// whole reports whether that definition field is also the *nearest* field
// ancestor — that is, whether the cursor's node is the definition's own value
// rather than sitting on some field inside it. Callers that attribute
// information to the definition name must require it: crossing an inner field
// on the way up means the information belongs to that field, not the
// definition.
func enclosingDefinition(c astutil.Cursor) (name string, def *ast.Field, whole, found bool) {
	whole = true
	for p := c; p != nil; p = p.Parent() {
		fld, ok := p.Node().(*ast.Field)
		if !ok {
			continue
		}
		if id, ok := fld.Label.(*ast.Ident); ok && strings.HasPrefix(id.Name, "#") {
			return strings.TrimPrefix(id.Name, "#"), fld, whole, true
		}
		// A field that is not the definition: anything found above it encloses
		// that field rather than holding the cursor's own value.
		whole = false
	}
	return "", nil, false, false
}

// declaresFormat reports whether a definition field carries a well-formed
// @gemara(format="...") directive. A malformed one is left for
// definitionFormats to report with its full context.
func declaresFormat(def *ast.Field) bool {
	for _, attr := range def.Attrs {
		if key, _ := attr.Split(); key != "gemara" {
			continue
		}
		if directive, err := parseFileGemaraDirective(attr); err == nil && directive.format != nil {
			return true
		}
	}
	return false
}
