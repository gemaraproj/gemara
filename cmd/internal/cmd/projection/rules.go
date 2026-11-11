// SPDX-License-Identifier: Apache-2.0

package projection

import (
	"fmt"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// rulesExtension lists, on a schema, the rules its CUE enforces that the
// OpenAPI document cannot, such as unique ids and references that must
// resolve to a declared entry.
const rulesExtension = "x-rules"

// applyRuleListTyped lists each conditional prep had to drop under x-rules on
// the schema of the definition that declares it. A rule declared in a hidden
// #_ helper goes on the helper's base, which every use of the helper is
// collapsed onto, so this must run before collapseHelperSchemasTyped.
//
// A dropped rule with no comment fails generation: the list is the only place
// the document still states it.
func applyRuleListTyped(doc *openapi3.T, dropped []droppedConditional) error {
	if len(dropped) == 0 {
		return nil
	}
	bases, err := helperBases(doc.Components.Schemas)
	if err != nil {
		return err
	}
	for _, d := range dropped {
		if d.owner == "" {
			return fmt.Errorf("%s: cross-field conditional is outside any definition, so cue2openapi has no schema to list it on", d.pos)
		}
		if d.doc == "" {
			return fmt.Errorf("%s: #%s has a cross-field rule cue2openapi cannot project; describe it in a // comment directly above it, which is listed under %s", d.pos, d.owner, rulesExtension)
		}
		name := d.owner
		if base, ok := bases[name]; ok {
			name = base
		}
		ref, ok := doc.Components.Schemas[name]
		if !ok || ref == nil || ref.Value == nil {
			return fmt.Errorf("%s: #%s is not in the generated document, so its rule cannot be listed", d.pos, d.owner)
		}
		s := ref.Value
		if s.Extensions == nil {
			s.Extensions = map[string]any{}
		}
		rules, _ := s.Extensions[rulesExtension].([]string)
		// Each sentence is listed once per schema. Uncommented rules in one
		// hidden field all take that field's comment, which is where
		// repeats come from.
		if !slices.Contains(rules, d.doc) {
			s.Extensions[rulesExtension] = append(rules, d.doc)
		}
	}
	return nil
}
