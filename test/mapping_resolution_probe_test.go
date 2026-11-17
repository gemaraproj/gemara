// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"encoding/json"
	"fmt"
)

// resolvePrototype is deliberately limited to Control -> Guideline mappings.
// It is a throwaway feasibility probe, not a supported resolver API.
func resolvePrototype(catalog, mapping map[string]any) (map[string]any, error) {
	if mapping["source-reference"].(map[string]any)["entry-type"] != "Control" ||
		mapping["target-reference"].(map[string]any)["entry-type"] != "Guideline" {
		return nil, fmt.Errorf("prototype supports only Control -> Guideline")
	}
	// Work on a copy: resolving must not rewrite the author's input artifact.
	data, err := json.Marshal(catalog)
	if err != nil {
		return nil, err
	}
	var output map[string]any
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	metadata := output["metadata"].(map[string]any)
	sourceRef := mapping["source-reference"].(map[string]any)["reference-id"]
	targetRef := mapping["target-reference"].(map[string]any)["reference-id"]
	var sourceDeclaration, targetDeclaration map[string]any
	for _, raw := range mapping["metadata"].(map[string]any)["mapping-references"].([]any) {
		ref := raw.(map[string]any)
		if ref["id"] == sourceRef {
			sourceDeclaration = ref
		}
		if ref["id"] == targetRef {
			targetDeclaration = ref
		}
	}
	if sourceDeclaration == nil {
		return nil, fmt.Errorf("source reference is undeclared")
	}
	if sourceDeclaration["id"] != metadata["id"] || sourceDeclaration["version"] != metadata["version"] {
		return nil, fmt.Errorf("source catalog identity/version does not match")
	}
	if targetDeclaration == nil {
		return nil, fmt.Errorf("target reference is undeclared")
	}
	refs, _ := metadata["mapping-references"].([]any)
	for _, raw := range refs {
		if raw.(map[string]any)["id"] == targetRef {
			return nil, fmt.Errorf("target reference alias %v already exists; prototype will not merge it", targetRef)
		}
	}
	metadata["mapping-references"] = append(refs, targetDeclaration)
	// Applicability ids are local to the MappingDocument and must travel too.
	groups, _ := metadata["applicability-groups"].([]any)
	mappingGroups, _ := mapping["metadata"].(map[string]any)["applicability-groups"].([]any)
	for _, raw := range mappingGroups {
		group := raw.(map[string]any)
		for _, existing := range groups {
			if existing.(map[string]any)["id"] == group["id"] {
				return nil, fmt.Errorf("applicability alias %v already exists", group["id"])
			}
		}
		groups = append(groups, group)
	}
	if len(groups) > 0 {
		metadata["applicability-groups"] = groups
	}
	controls := map[string]map[string]any{}
	for _, raw := range output["controls"].([]any) {
		control := raw.(map[string]any)
		controls[control["id"].(string)] = control
	}
	for _, raw := range mapping["mappings"].([]any) {
		m := raw.(map[string]any)
		control := controls[m["source"].(string)]
		if control == nil {
			return nil, fmt.Errorf("source entry %v not found", m["source"])
		}
		group := map[string]any{"id": m["id"], "reference-id": targetRef, "relationship": m["relationship"]}
		if remarks, exists := m["remarks"]; exists {
			group["remarks"] = remarks
		}
		if targets, exists := m["targets"].([]any); exists {
			entries := make([]any, 0, len(targets))
			for _, raw := range targets {
				target := map[string]any{}
				for key, value := range raw.(map[string]any) {
					if key == "entry-id" {
						key = "reference-id"
					}
					target[key] = value
				}
				entries = append(entries, target)
			}
			group["entries"] = entries
		}
		links, _ := control["guidelines"].([]any)
		control["guidelines"] = append(links, group)
	}
	return output, nil
}
