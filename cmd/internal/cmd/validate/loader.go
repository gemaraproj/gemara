// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// LoadPaths resolves a list of file or directory paths and parses all Gemara YAML/JSON documents.
func LoadPaths(paths []string) ([]*Document, error) {
	var docs []*Document

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat path %q: %w", p, err)
		}

		if info.IsDir() {
			err = filepath.Walk(p, func(path string, f os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if f.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".yaml" || ext == ".yml" || ext == ".json" {
					doc, err := LoadDocument(path)
					if err != nil {
						// Skip files that aren't valid Gemara documents
						return nil
					}
					if doc != nil {
						docs = append(docs, doc)
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk directory %q: %w", p, err)
			}
		} else {
			doc, err := LoadDocument(p)
			if err != nil {
				return nil, err
			}
			if doc != nil {
				docs = append(docs, doc)
			}
		}
	}

	return docs, nil
}

// LoadDocument reads a single YAML/JSON file and parses its metadata, declared entities, and references.
func LoadDocument(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}

	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("unmarshal YAML %q: %w", path, err)
	}

	metaRaw, ok := root["metadata"].(map[string]any)
	if !ok {
		// Not a standard Gemara document with metadata
		return nil, nil
	}

	doc := &Document{
		FilePath:          path,
		MappingReferences: make(map[string]MappingReference),
		DeclaredEntities:  make(map[string]bool),
		SubEntities:       make(map[string]map[string]bool),
	}

	if title, ok := root["title"].(string); ok {
		doc.Title = title
	}
	if id, ok := metaRaw["id"].(string); ok {
		doc.ID = id
	}
	if docType, ok := metaRaw["type"].(string); ok {
		doc.Type = docType
	}
	if ver, ok := metaRaw["version"].(string); ok {
		doc.Version = ver
	}

	// Parse metadata.mapping-references
	if mrList, ok := metaRaw["mapping-references"].([]any); ok {
		for _, item := range mrList {
			if mrMap, ok := item.(map[string]any); ok {
				var mr MappingReference
				if id, ok := mrMap["id"].(string); ok {
					mr.ID = id
				}
				if t, ok := mrMap["title"].(string); ok {
					mr.Title = t
				}
				if u, ok := mrMap["url"].(string); ok {
					mr.URL = u
				}
				if v, ok := mrMap["version"].(string); ok {
					mr.Version = v
				}
				if mr.ID != "" {
					doc.MappingReferences[mr.ID] = mr
				}
			}
		}
	}

	// Extract declared entities & references based on document type
	extractDocumentContents(doc, root)

	return doc, nil
}

func extractDocumentContents(doc *Document, root map[string]any) {
	switch doc.Type {
	case "ControlCatalog":
		extractControlCatalog(doc, root)
	case "RiskCatalog":
		extractSimpleList(doc, root, "risks")
	case "ThreatCatalog":
		extractThreatCatalog(doc, root)
	case "CapabilityCatalog":
		extractSimpleList(doc, root, "capabilities")
	case "GuidanceCatalog":
		extractSimpleList(doc, root, "guidelines")
	case "VectorCatalog":
		extractSimpleList(doc, root, "vectors")
	case "PrincipleCatalog":
		extractSimpleList(doc, root, "principles")
	case "MappingDocument":
		extractMappingDocument(doc, root)
	case "Policy":
		extractPolicy(doc, root)
	case "EvaluationLog":
		extractEvaluationLog(doc, root)
	case "EnforcementLog":
		extractEnforcementLog(doc, root)
	case "AuditLog":
		extractAuditLog(doc, root)
	default:
		// Generic extraction
		extractGenericArtifact(doc, root)
	}
}

func extractControlCatalog(doc *Document, root map[string]any) {
	controls, ok := root["controls"].([]any)
	if !ok {
		return
	}
	for _, c := range controls {
		cMap, ok := c.(map[string]any)
		if !ok {
			continue
		}
		cID, ok := cMap["id"].(string)
		if !ok || cID == "" {
			continue
		}
		doc.DeclaredEntities[cID] = true

		// Assessment requirements
		if arList, ok := cMap["assessment-requirements"].([]any); ok {
			if doc.SubEntities[cID] == nil {
				doc.SubEntities[cID] = make(map[string]bool)
			}
			for _, ar := range arList {
				if arMap, ok := ar.(map[string]any); ok {
					if arID, ok := arMap["id"].(string); ok && arID != "" {
						doc.SubEntities[cID][arID] = true
					}
				}
			}
		}

		// Guidelines mapped from control
		extractMultiEntryMappings(doc, cMap["guidelines"], "control.guidelines")
		// Threats mapped from control
		extractMultiEntryMappings(doc, cMap["threats"], "control.threats")
	}
}

func extractThreatCatalog(doc *Document, root map[string]any) {
	threats, ok := root["threats"].([]any)
	if !ok {
		return
	}
	for _, t := range threats {
		tMap, ok := t.(map[string]any)
		if !ok {
			continue
		}
		tID, ok := tMap["id"].(string)
		if !ok || tID == "" {
			continue
		}
		doc.DeclaredEntities[tID] = true

		extractMultiEntryMappings(doc, tMap["capabilities"], "threat.capabilities")
		extractMultiEntryMappings(doc, tMap["vectors"], "threat.vectors")
	}
}

func extractSimpleList(doc *Document, root map[string]any, key string) {
	items, ok := root[key].([]any)
	if !ok {
		return
	}
	for _, item := range items {
		if iMap, ok := item.(map[string]any); ok {
			if id, ok := iMap["id"].(string); ok && id != "" {
				doc.DeclaredEntities[id] = true
			}
		}
	}
}

func extractMappingDocument(doc *Document, root map[string]any) {
	var srcRef, tgtRef string

	if src, ok := root["source-reference"].(map[string]any); ok {
		if rID, ok := src["reference-id"].(string); ok && rID != "" {
			srcRef = rID
			doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
				ReferenceID: rID,
				Context:     "source-reference",
			})
		}
	}

	if tgt, ok := root["target-reference"].(map[string]any); ok {
		if rID, ok := tgt["reference-id"].(string); ok && rID != "" {
			tgtRef = rID
			doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
				ReferenceID: rID,
				Context:     "target-reference",
			})
		}
	}

	mappings, ok := root["mappings"].([]any)
	if !ok {
		return
	}

	for _, m := range mappings {
		mMap, ok := m.(map[string]any)
		if !ok {
			continue
		}
		mID, _ := mMap["id"].(string)
		if mID != "" {
			doc.DeclaredEntities[mID] = true
		}

		entry := MappingEntry{
			ID:        mID,
			SourceRef: srcRef,
			TargetRef: tgtRef,
		}
		if s, ok := mMap["source"].(string); ok {
			entry.Source = s
		}
		if r, ok := mMap["relationship"].(string); ok {
			entry.Relationship = r
		}

		if targets, ok := mMap["targets"].([]any); ok {
			for _, tgt := range targets {
				if tMap, ok := tgt.(map[string]any); ok {
					if entryID, ok := tMap["entry-id"].(string); ok && entryID != "" {
						entry.Targets = append(entry.Targets, entryID)
					}
				}
			}
		}

		doc.Mappings = append(doc.Mappings, entry)
	}
}

func extractPolicy(doc *Document, root map[string]any) {
	if imports, ok := root["imports"].(map[string]any); ok {
		for _, catKey := range []string{"catalogs", "policies", "guidance"} {
			if list, ok := imports[catKey].([]any); ok {
				for _, item := range list {
					if iMap, ok := item.(map[string]any); ok {
						if rID, ok := iMap["reference-id"].(string); ok && rID != "" {
							doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
								ReferenceID: rID,
								Context:     "imports." + catKey,
							})
						}
					}
				}
			}
		}
	}

	if risks, ok := root["risks"].(map[string]any); ok {
		for _, rk := range []string{"mitigated", "accepted"} {
			if list, ok := risks[rk].([]any); ok {
				for _, item := range list {
					if rMap, ok := item.(map[string]any); ok {
						rID, _ := rMap["reference-id"].(string)
						entryID, _ := rMap["entry-id"].(string)
						if entryID == "" {
							entryID, _ = rMap["risk-id"].(string)
						}
						if rID != "" {
							doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
								ReferenceID: rID,
								Context:     "risks." + rk,
							})
							if entryID != "" {
								doc.EntityUsages = append(doc.EntityUsages, EntityUsage{
									ReferenceID: rID,
									EntryID:     entryID,
									Context:     "risks." + rk,
								})
							}
						}
					}
				}
			}
		}
	}
}

func extractEvaluationLog(doc *Document, root map[string]any) {
	evals, ok := root["evaluations"].([]any)
	if !ok {
		return
	}

	for _, e := range evals {
		eMap, ok := e.(map[string]any)
		if !ok {
			continue
		}

		var ctrlRef, ctrlID string
		if ctrl, ok := eMap["control"].(map[string]any); ok {
			ctrlRef, _ = ctrl["reference-id"].(string)
			ctrlID, _ = ctrl["entry-id"].(string)
			if ctrlRef != "" {
				doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
					ReferenceID: ctrlRef,
					Context:     "evaluations.control",
				})
				if ctrlID != "" {
					// Index the evaluated control so downstream references (e.g. enforcement logs, ref #499) resolve
					doc.DeclaredEntities[ctrlID] = true
					doc.EntityUsages = append(doc.EntityUsages, EntityUsage{
						ReferenceID: ctrlRef,
						EntryID:     ctrlID,
						Context:     "evaluations.control",
					})
				}
			}
		}

		if aLogs, ok := eMap["assessment-logs"].([]any); ok {
			for _, al := range aLogs {
				alMap, ok := al.(map[string]any)
				if !ok {
					continue
				}
				if req, ok := alMap["requirement"].(map[string]any); ok {
					reqRef, _ := req["reference-id"].(string)
					reqID, _ := req["entry-id"].(string)
					if reqRef == "" {
						reqRef = ctrlRef // Inherited from control
					} else {
						doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
							ReferenceID: reqRef,
							Context:     "evaluations.assessment-logs.requirement",
						})
					}
					if reqRef != "" && reqID != "" {
						doc.EntityUsages = append(doc.EntityUsages, EntityUsage{
							ReferenceID: reqRef,
							EntryID:     ctrlID,
							SubEntryID:  reqID,
							Context:     "evaluations.assessment-logs.requirement",
						})
					}
				}
			}
		}
	}
}

func extractEnforcementLog(doc *Document, root map[string]any) {
	actions, ok := root["actions"].([]any)
	if !ok {
		return
	}

	for _, a := range actions {
		aMap, ok := a.(map[string]any)
		if !ok {
			continue
		}
		if just, ok := aMap["justification"].(map[string]any); ok {
			if assess, ok := just["assessments"].([]any); ok {
				for _, as := range assess {
					asMap, ok := as.(map[string]any)
					if !ok {
						continue
					}
					if logMap, ok := asMap["log"].(map[string]any); ok {
						lRef, _ := logMap["reference-id"].(string)
						lID, _ := logMap["entry-id"].(string)
						if lRef != "" {
							doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
								ReferenceID: lRef,
								Context:     "actions.justification.assessments.log",
							})
							if lID != "" {
								doc.EntityUsages = append(doc.EntityUsages, EntityUsage{
									ReferenceID: lRef,
									EntryID:     lID,
									Context:     "actions.justification.assessments.log",
								})
							}
						}
					}
				}
			}
		}
	}
}

func extractAuditLog(doc *Document, root map[string]any) {
	if criteria, ok := root["criteria"].([]any); ok {
		for _, c := range criteria {
			if cMap, ok := c.(map[string]any); ok {
				if rID, ok := cMap["reference-id"].(string); ok && rID != "" {
					doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
						ReferenceID: rID,
						Context:     "auditlog.criteria",
					})
				}
			}
		}
	}

	if results, ok := root["results"].([]any); ok {
		for _, r := range results {
			if rMap, ok := r.(map[string]any); ok {
				if critRef, ok := rMap["criteria-reference"].(map[string]any); ok {
					if rID, ok := critRef["reference-id"].(string); ok && rID != "" {
						doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
							ReferenceID: rID,
							Context:     "auditlog.results.criteria-reference",
						})
					}
				}
			}
		}
	}
}

func extractGenericArtifact(doc *Document, root map[string]any) {
	// Fallback extraction
	for _, key := range []string{"controls", "risks", "threats", "capabilities", "guidelines", "vectors", "principles"} {
		extractSimpleList(doc, root, key)
	}
}

func extractMultiEntryMappings(doc *Document, raw any, context string) {
	mappings, ok := raw.([]any)
	if !ok {
		return
	}
	for _, m := range mappings {
		mMap, ok := m.(map[string]any)
		if !ok {
			continue
		}
		rID, _ := mMap["reference-id"].(string)
		if rID != "" {
			doc.ReferenceUsages = append(doc.ReferenceUsages, ReferenceUsage{
				ReferenceID: rID,
				Context:     context,
			})
			if entries, ok := mMap["entries"].([]any); ok {
				for _, e := range entries {
					if eMap, ok := e.(map[string]any); ok {
						if entryID, ok := eMap["entry-id"].(string); ok && entryID != "" {
							doc.EntityUsages = append(doc.EntityUsages, EntityUsage{
								ReferenceID: rID,
								EntryID:     entryID,
								Context:     context,
							})
						}
					}
				}
			}
		}
	}
}
