// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"bytes"
	"fmt"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/literal"
	"cuelang.org/go/cue/token"
	"github.com/getkin/kin-openapi/openapi3"
)

type requiredUnless struct {
	Schema, Field, Selector, Value string
}

// extractRequiredUnless recognizes only presence constraints of the form
// `field?: T; if selector != "value" {field: T}` directly inside a definition.
// The identical T check is important: a narrowed type must not be lost. Other
// comprehensions stay on the existing, explicitly reported omission path.
func extractRequiredUnless(f *ast.File) []requiredUnless {
	var rules []requiredUnless
	astutil.Apply(f, func(c astutil.Cursor) bool {
		s, ok := c.Node().(*ast.StructLit)
		if !ok {
			return true
		}
		name, _, whole, found := enclosingDefinition(c)
		if !found || !whole {
			return true
		}
		out := s.Elts[:0:0]
		for _, decl := range s.Elts {
			if rule, ok := presenceRule(name, s, decl); ok {
				rules = append(rules, rule)
			} else {
				out = append(out, decl)
			}
		}
		s.Elts = out
		return true
	}, nil)
	return rules
}

func presenceRule(name string, s *ast.StructLit, decl ast.Decl) (requiredUnless, bool) {
	comp, ok := decl.(*ast.Comprehension)
	if !ok || len(comp.Clauses) != 1 {
		return requiredUnless{}, false
	}
	clause, ok := comp.Clauses[0].(*ast.IfClause)
	if !ok {
		return requiredUnless{}, false
	}
	cond, ok := clause.Condition.(*ast.BinaryExpr)
	if !ok || cond.Op != token.NEQ {
		return requiredUnless{}, false
	}
	selector, ok := cond.X.(*ast.Ident)
	if !ok {
		return requiredUnless{}, false
	}
	value, ok := cond.Y.(*ast.BasicLit)
	if !ok || value.Kind != token.STRING {
		return requiredUnless{}, false
	}
	unquoted, err := literal.Unquote(value.Value)
	if err != nil {
		return requiredUnless{}, false
	}
	body, ok := comp.Value.(*ast.StructLit)
	if !ok || len(body.Elts) != 1 {
		return requiredUnless{}, false
	}
	required, ok := body.Elts[0].(*ast.Field)
	if !ok || required.Constraint != token.ILLEGAL || required.Optional.IsValid() {
		return requiredUnless{}, false
	}
	field, _, err := ast.LabelName(required.Label)
	if err != nil {
		return requiredUnless{}, false
	}
	for _, d := range s.Elts {
		optional, ok := d.(*ast.Field)
		if !ok || (optional.Constraint != token.OPTION && !optional.Optional.IsValid()) {
			continue
		}
		label, _, err := ast.LabelName(optional.Label)
		if err != nil || label != field {
			continue
		}
		a, errA := format.Node(optional.Value)
		b, errB := format.Node(required.Value)
		if errA == nil && errB == nil && bytes.Equal(a, b) {
			return requiredUnless{name, field, selector.Name, unquoted}, true
		}
	}
	return requiredUnless{}, false
}

func applyRequiredUnlessTyped(doc *openapi3.T, rules []requiredUnless) error {
	for _, rule := range rules {
		ref := doc.Components.Schemas[rule.Schema]
		if ref == nil || ref.Value == nil || declaringSchema(doc.Components.Schemas, rule.Schema, rule.Field) == nil ||
			declaringSchema(doc.Components.Schemas, rule.Schema, rule.Selector) == nil {
			return fmt.Errorf("cannot project required-unless rule for %s.%s", rule.Schema, rule.Field)
		}
		// Requiring the selector in the exception branch matters: an omitted
		// selector uses its normal default and must not create an implicit gap.
		exceptionRequired := []string{rule.Selector}
		holder := declaringSchema(doc.Components.Schemas, rule.Schema, rule.Selector)
		if holder.Properties[rule.Selector].Value.Default == rule.Value {
			exceptionRequired = nil
		}
		ref.Value.AllOf = append(ref.Value.AllOf, &openapi3.SchemaRef{Value: &openapi3.Schema{
			AnyOf: openapi3.SchemaRefs{
				{Value: &openapi3.Schema{
					Required: exceptionRequired,
					Properties: openapi3.Schemas{rule.Selector: {Value: &openapi3.Schema{
						Type: &openapi3.Types{"string"}, Enum: []any{rule.Value},
					}}},
				}},
				{Value: &openapi3.Schema{Required: []string{rule.Field}}},
			},
		}})
	}
	return nil
}
