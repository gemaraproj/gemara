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

// definitionUses lists every place files name a definition. Prep reads them
// before prepare strips comprehensions, so a name inside a cross-field rule
// counts too.
func definitionUses(files []*ast.File) []definitionUse {
	var out []definitionUse
	for _, f := range files {
		labels := map[*ast.Ident]bool{}
		packageLabels(f.Decls, labels)
		astutil.Apply(f, func(c astutil.Cursor) bool {
			id, ok := c.Node().(*ast.Ident)
			if !ok || !isDefinitionName(id.Name) {
				return true
			}
			// x.#Name selects from x, not from this package.
			if sel, ok := c.Parent().Node().(*ast.SelectorExpr); ok && sel.Sel == id {
				return true
			}
			out = append(out, definitionUse{name: id.Name, file: f.Filename, pos: id.Pos().String(), decl: labels[id]})
			return true
		}, nil)
	}
	return out
}

// packageLabels records the labels that declare package-level fields: those of
// top-level fields, aliased or not, and of fields in a top-level embedded
// struct or comprehension, which CUE merges into the package the same way.
func packageLabels(decls []ast.Decl, out map[*ast.Ident]bool) {
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.Field:
			label := d.Label
			if a, ok := label.(*ast.Alias); ok {
				label, _ = a.Expr.(ast.Label)
			}
			if id, ok := label.(*ast.Ident); ok {
				out[id] = true
			}
		case *ast.EmbedDecl:
			if s, ok := d.Expr.(*ast.StructLit); ok {
				packageLabels(s.Elts, out)
			}
		case *ast.Comprehension:
			if s, ok := d.Value.(*ast.StructLit); ok {
				packageLabels(s.Elts, out)
			}
		}
	}
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
