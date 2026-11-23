// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"fmt"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
)

// definitionUse is one place a file names a definition of the package: the
// label of a top-level declaration, or a reference.
type definitionUse struct {
	name string // as written, such as "#Evidence"
	file string // filename, as fileStatuses keys it
	pos  string // source position, for reports
	decl bool   // the label of a top-level declaration
}

// definitionUses lists every place files name a package-level definition.
// Prep reads them before prepare strips comprehensions, so a name inside a
// cross-field rule counts too. A definition declared inside another one is
// part of it, so neither its label nor a reference the parser bound to it is
// listed. It fails when a file embeds anything but a struct literal at the
// top level, since which package-level definitions that adds to can't be read
// from the syntax.
func definitionUses(files []*ast.File) ([]definitionUse, error) {
	var out []definitionUse
	var opaque []string
	for _, f := range files {
		labels := map[*ast.Ident]bool{}
		allLabels := fieldLabels(f)
		opaque = append(opaque, packageLabels(f.Decls, labels)...)
		astutil.Apply(f, func(c astutil.Cursor) bool {
			id, ok := c.Node().(*ast.Ident)
			if !ok || !isDefinitionName(id.Name) {
				return true
			}
			// x.#Name selects from x, not from this package.
			if sel, ok := c.Parent().Node().(*ast.SelectorExpr); ok && sel.Sel == id {
				return true
			}
			// A label nested in a struct declares a nested definition.
			if allLabels[id] && !labels[id] {
				return true
			}
			// The parser binds a reference to a nested declaration in scope
			// to the struct that declares it. A reference to a package-level
			// definition is bound to its file, or left unbound when another
			// file declares it.
			if _, nested := id.Scope.(*ast.StructLit); nested {
				return true
			}
			out = append(out, definitionUse{name: id.Name, file: f.Filename, pos: id.Pos().String(), decl: labels[id]})
			return true
		}, nil)
	}
	if len(opaque) > 0 {
		return nil, fmt.Errorf("only a struct literal can be embedded at the top level of a schema file, so the stable-reference check can see which definitions it adds to:\n  %s", strings.Join(opaque, "\n  "))
	}
	return out, nil
}

// fieldLabels records the label of every field in f, at any depth.
func fieldLabels(f *ast.File) map[*ast.Ident]bool {
	out := map[*ast.Ident]bool{}
	ast.Walk(f, func(n ast.Node) bool {
		if field, ok := n.(*ast.Field); ok {
			if id, ok := unaliased(field.Label).(*ast.Ident); ok {
				out[id] = true
			}
		}
		return true
	}, nil)
	return out
}

// unaliased returns the label an alias such as X=#A: names.
func unaliased(label ast.Label) ast.Label {
	if a, ok := label.(*ast.Alias); ok {
		if l, ok := a.Expr.(ast.Label); ok {
			return l
		}
	}
	return label
}

// packageLabels records the labels that declare package-level fields: those of
// top-level fields, aliased or not, and of fields in a top-level embedded
// struct or comprehension, which CUE merges into the package the same way. It
// returns the position of each top-level embedding that is anything other than
// a struct literal, such as a reference, a call or a unification, since the
// fields that adds to the package can't be read from the syntax.
func packageLabels(decls []ast.Decl, out map[*ast.Ident]bool) []string {
	var opaque []string
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.Field:
			if id, ok := unaliased(d.Label).(*ast.Ident); ok {
				out[id] = true
			}
		case *ast.EmbedDecl:
			expr := d.Expr
			for p, ok := expr.(*ast.ParenExpr); ok; p, ok = expr.(*ast.ParenExpr) {
				expr = p.X
			}
			if s, ok := expr.(*ast.StructLit); ok {
				opaque = append(opaque, packageLabels(s.Elts, out)...)
			} else {
				opaque = append(opaque, d.Pos().String())
			}
		case *ast.Comprehension:
			if s, ok := d.Value.(*ast.StructLit); ok {
				opaque = append(opaque, packageLabels(s.Elts, out)...)
			}
		}
	}
	return opaque
}

func isDefinitionName(name string) bool {
	return strings.HasPrefix(name, "#") || strings.HasPrefix(name, "_#")
}

// checkStableReferences holds the schemas to the versioning policy: a stable
// schema may only reference stable types
// (https://gemara.openssf.org/schema/#stable-status).
// CUE merges every declaration of a definition across the package, so a
// definition is only stable when every file declaring part of it is. Each
// definition a stable file names, whether it refers to it or declares part of
// it, must be declared in stable files only. Otherwise a change to a
// non-stable file could change a stable schema, and x-status, which is read
// from a single declaring file, would hide that from breaking-check.
func checkStableReferences(uses []definitionUse, statuses map[string]string) error {
	declaredIn := map[string][]string{}
	for _, u := range uses {
		if u.decl {
			declaredIn[u.name] = append(declaredIn[u.name], u.file)
		}
	}
	reported := map[[3]string]bool{}
	var problems []string
	for _, u := range uses {
		if statuses[u.file] != "stable" {
			continue
		}
		for _, file := range declaredIn[u.name] {
			status := statuses[file]
			key := [3]string{u.file, u.name, file}
			if status == "stable" || reported[key] {
				continue
			}
			reported[key] = true
			problems = append(problems, fmt.Sprintf("%s: stable %s names %s, which %s %s declares", u.pos, filepath.Base(u.file), u.name, status, filepath.Base(file)))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("a stable schema may only reference stable types (https://gemara.openssf.org/schema/#stable-status), so each definition below must be declared in stable files only:\n  %s", strings.Join(problems, "\n  "))
}
