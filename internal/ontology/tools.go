package ontology

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates agent-facing tools for ontology management and data ingestion.
// When reg is non-nil, dynamic tools are generated for each registered action.
func BuildTools(svc OntologyService, reg *ActionRegistry) []*agent.Tool {
	tools := []*agent.Tool{
		buildListTypes(svc),
		buildDescribeType(svc),
		buildQueryEntities(svc),
		buildGetEntity(svc),
		buildAssertFact(svc),
		buildRetractFact(svc),
		buildListConflicts(svc),
		buildResolveConflict(svc),
		buildMergeEntities(svc),
		buildFactsAt(svc),
		buildImportJSON(svc),
		buildImportCSV(svc),
		buildFromMCP(svc),
	}
	if reg != nil {
		tools = append(tools, buildListActions(svc))
		for _, action := range reg.List() {
			tools = append(tools, buildActionTool(svc, action))
		}
	}
	return tools
}

func buildListActions(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_list_actions",
		Description: "List all registered ontology actions with their parameters and required permissions.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, _ map[string]interface{}) (interface{}, error) {
			actions, err := svc.ListActions(ctx)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"actions": actions,
				"count":   len(actions),
			}, nil
		},
	}
}

func buildActionTool(svc OntologyService, action *ActionType) *agent.Tool {
	properties := make(map[string]interface{}, len(action.ParamSchema))
	required := make([]string, 0, len(action.ParamSchema))
	for name, desc := range action.ParamSchema {
		properties[name] = map[string]interface{}{
			"type":        "string",
			"description": desc,
		}
		required = append(required, name)
	}
	return &agent.Tool{
		Name:        "ontology_action_" + action.Name,
		Description: action.Description,
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		},
		Handler: func(ctx context.Context, rawParams map[string]interface{}) (interface{}, error) {
			strParams := make(map[string]string, len(rawParams))
			for k, v := range rawParams {
				strParams[k] = fmt.Sprintf("%v", v)
			}
			result, err := svc.ExecuteAction(ctx, action.Name, strParams)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"logID":   result.LogID.String(),
				"status":  string(result.Status),
				"effects": result.Effects,
				"error":   result.Error,
			}, nil
		},
	}
}

func buildListTypes(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_list_types",
		Description: "List all registered ObjectTypes in the ontology.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, _ map[string]interface{}) (interface{}, error) {
			types, err := svc.ListTypes(ctx)
			if err != nil {
				return nil, fmt.Errorf("list types: %w", err)
			}
			items := make([]map[string]interface{}, len(types))
			for i, t := range types {
				items[i] = map[string]interface{}{
					"name":        t.Name,
					"description": t.Description,
					"status":      string(t.Status),
					"properties":  len(t.Properties),
				}
			}
			return map[string]interface{}{"types": items, "count": len(items)}, nil
		},
	}
}

func buildDescribeType(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_describe_type",
		Description: "Describe an ObjectType with its properties and relevant predicates.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type_name": map[string]interface{}{"type": "string", "description": "ObjectType name to describe"},
			},
			"required": []string{"type_name"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			name, err := toolparam.RequireString(params, "type_name")
			if err != nil {
				return nil, err
			}
			objType, err := svc.GetType(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("describe type: %w", err)
			}
			preds, _ := svc.ListPredicates(ctx)
			var relevant []map[string]interface{}
			for _, p := range preds {
				if containsStr(p.SourceTypes, name) || containsStr(p.TargetTypes, name) || len(p.SourceTypes) == 0 {
					relevant = append(relevant, map[string]interface{}{
						"name":        p.Name,
						"cardinality": string(p.Cardinality),
						"description": p.Description,
					})
				}
			}
			return map[string]interface{}{
				"name":        objType.Name,
				"description": objType.Description,
				"status":      string(objType.Status),
				"properties":  objType.Properties,
				"predicates":  relevant,
			}, nil
		},
	}
}

func buildQueryEntities(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_query_entities",
		Description: "Query entities by type and property filters. Returns matching entities with properties and outgoing triples.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{"type": "string", "description": "ObjectType name to query"},
				"filters": map[string]interface{}{
					"type":        "array",
					"description": "Property filters [{property, op, value}]. Ops: eq, neq, contains",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"property": map[string]interface{}{"type": "string"},
							"op":       map[string]interface{}{"type": "string"},
							"value":    map[string]interface{}{"type": "string"},
						},
					},
				},
				"limit": map[string]interface{}{"type": "integer", "description": "Max results (default 100)"},
			},
			"required": []string{"type"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			typeName, err := toolparam.RequireString(params, "type")
			if err != nil {
				return nil, err
			}
			limit := toolparam.OptionalInt(params, "limit", 100)
			filters := parseFilters(params)
			results, err := svc.QueryEntities(ctx, PropertyQuery{
				EntityType: typeName,
				Filters:    filters,
				Limit:      limit,
			})
			if err != nil {
				return nil, fmt.Errorf("query entities: %w", err)
			}
			return map[string]interface{}{"entities": results, "count": len(results)}, nil
		},
	}
}

func buildGetEntity(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_get_entity",
		Description: "Get a single entity with its properties, outgoing and incoming triples.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"entity_id": map[string]interface{}{"type": "string", "description": "Entity ID to retrieve"},
			},
			"required": []string{"entity_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			id, err := toolparam.RequireString(params, "entity_id")
			if err != nil {
				return nil, err
			}
			entity, err := svc.GetEntity(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("get entity: %w", err)
			}
			return entity, nil
		},
	}
}

func buildAssertFact(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_assert_fact",
		Description: "Assert a fact (triple) with temporal metadata and conflict detection.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"subject":      map[string]interface{}{"type": "string", "description": "Subject entity ID"},
				"predicate":    map[string]interface{}{"type": "string", "description": "Predicate name"},
				"object":       map[string]interface{}{"type": "string", "description": "Object entity ID"},
				"subject_type": map[string]interface{}{"type": "string", "description": "Subject ObjectType (optional)"},
				"object_type":  map[string]interface{}{"type": "string", "description": "Object ObjectType (optional)"},
				"source":       map[string]interface{}{"type": "string", "description": "Source category (default: manual)"},
				"confidence":   map[string]interface{}{"type": "number", "description": "Confidence 0.0-1.0 (default: 0.8)"},
			},
			"required": []string{"subject", "predicate", "object"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			subject, err := toolparam.RequireString(params, "subject")
			if err != nil {
				return nil, err
			}
			predicate, err := toolparam.RequireString(params, "predicate")
			if err != nil {
				return nil, err
			}
			object, err := toolparam.RequireString(params, "object")
			if err != nil {
				return nil, err
			}
			input := AssertionInput{
				Triple: graph.Triple{
					Subject:     subject,
					Predicate:   predicate,
					Object:      object,
					SubjectType: toolparam.OptionalString(params, "subject_type", ""),
					ObjectType:  toolparam.OptionalString(params, "object_type", ""),
				},
				Source:     toolparam.OptionalString(params, "source", "manual"),
				Confidence: toolparam.OptionalFloat64(params, "confidence", 0.8),
			}
			result, err := svc.AssertFact(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("assert fact: %w", err)
			}
			resp := map[string]interface{}{"stored": result.Stored, "message": result.Message}
			if result.ConflictID != nil {
				resp["conflict_id"] = result.ConflictID.String()
			}
			return resp, nil
		},
	}
}

func buildRetractFact(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_retract_fact",
		Description: "Retract a fact (soft delete by setting ValidTo).",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"subject":   map[string]interface{}{"type": "string", "description": "Subject entity ID"},
				"predicate": map[string]interface{}{"type": "string", "description": "Predicate name"},
				"object":    map[string]interface{}{"type": "string", "description": "Object entity ID"},
				"reason":    map[string]interface{}{"type": "string", "description": "Reason for retraction"},
			},
			"required": []string{"subject", "predicate", "object", "reason"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			subject, err := toolparam.RequireString(params, "subject")
			if err != nil {
				return nil, err
			}
			predicate, err := toolparam.RequireString(params, "predicate")
			if err != nil {
				return nil, err
			}
			object, err := toolparam.RequireString(params, "object")
			if err != nil {
				return nil, err
			}
			reason, err := toolparam.RequireString(params, "reason")
			if err != nil {
				return nil, err
			}
			if err := svc.RetractFact(ctx, subject, predicate, object, reason); err != nil {
				return nil, fmt.Errorf("retract fact: %w", err)
			}
			return map[string]interface{}{"status": "retracted"}, nil
		},
	}
}

func buildListConflicts(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_list_conflicts",
		Description: "List all open ontology conflicts awaiting resolution.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, _ map[string]interface{}) (interface{}, error) {
			conflicts, err := svc.OpenConflicts(ctx)
			if err != nil {
				return nil, fmt.Errorf("list conflicts: %w", err)
			}
			return map[string]interface{}{"conflicts": conflicts, "count": len(conflicts)}, nil
		},
	}
}

func buildResolveConflict(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_resolve_conflict",
		Description: "Resolve an ontology conflict by choosing a winner.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"conflict_id":   map[string]interface{}{"type": "string", "description": "Conflict UUID"},
				"winner_object": map[string]interface{}{"type": "string", "description": "Object value of the winning triple"},
				"reason":        map[string]interface{}{"type": "string", "description": "Resolution reason"},
			},
			"required": []string{"conflict_id", "winner_object", "reason"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			idStr, err := toolparam.RequireString(params, "conflict_id")
			if err != nil {
				return nil, err
			}
			conflictID, err := uuid.Parse(idStr)
			if err != nil {
				return nil, fmt.Errorf("invalid conflict_id: %w", err)
			}
			winner, err := toolparam.RequireString(params, "winner_object")
			if err != nil {
				return nil, err
			}
			reason, err := toolparam.RequireString(params, "reason")
			if err != nil {
				return nil, err
			}
			if err := svc.ResolveConflict(ctx, conflictID, winner, reason); err != nil {
				return nil, fmt.Errorf("resolve conflict: %w", err)
			}
			return map[string]interface{}{"status": "resolved"}, nil
		},
	}
}

func buildMergeEntities(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_merge_entities",
		Description: "Merge a duplicate entity into a canonical entity. Moves relationships and registers alias.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"canonical": map[string]interface{}{"type": "string", "description": "Canonical entity ID (target)"},
				"duplicate": map[string]interface{}{"type": "string", "description": "Duplicate entity ID (to be merged)"},
			},
			"required": []string{"canonical", "duplicate"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			canonical, err := toolparam.RequireString(params, "canonical")
			if err != nil {
				return nil, err
			}
			duplicate, err := toolparam.RequireString(params, "duplicate")
			if err != nil {
				return nil, err
			}
			result, err := svc.MergeEntities(ctx, canonical, duplicate)
			if err != nil {
				return nil, fmt.Errorf("merge entities: %w", err)
			}
			return map[string]interface{}{
				"status":          "merged",
				"triples_updated": result.TriplesUpdated,
				"aliases_created": result.AliasesCreated,
			}, nil
		},
	}
}

func buildFactsAt(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_facts_at",
		Description: "Query facts valid at a specific point in time (time-travel query).",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"subject":  map[string]interface{}{"type": "string", "description": "Subject entity ID"},
				"valid_at": map[string]interface{}{"type": "string", "description": "RFC3339 timestamp (e.g., 2026-03-30T12:00:00Z)"},
			},
			"required": []string{"subject", "valid_at"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			subject, err := toolparam.RequireString(params, "subject")
			if err != nil {
				return nil, err
			}
			validAtStr, err := toolparam.RequireString(params, "valid_at")
			if err != nil {
				return nil, err
			}
			validAt, err := time.Parse(time.RFC3339, validAtStr)
			if err != nil {
				return nil, fmt.Errorf("invalid valid_at format (expected RFC3339): %w", err)
			}
			facts, err := svc.FactsAt(ctx, subject, validAt)
			if err != nil {
				return nil, fmt.Errorf("facts at: %w", err)
			}
			return map[string]interface{}{"facts": facts, "count": len(facts)}, nil
		},
	}
}

// --- Ingestion Tools ---

// jsonImportEntity represents a single entity in the JSON import format.
type jsonImportEntity struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
	Relations  []jsonRelation    `json:"relations"`
}

type jsonRelation struct {
	Predicate  string `json:"predicate"`
	Object     string `json:"object"`
	ObjectType string `json:"object_type"`
}

type jsonImportPayload struct {
	Entities []jsonImportEntity `json:"entities"`
}

func buildImportJSON(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_import_json",
		Description: "Import entities from JSON data. Creates properties and relations.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{"type": "string", "description": "JSON string with entities array"},
			},
			"required": []string{"data"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			data, err := toolparam.RequireString(params, "data")
			if err != nil {
				return nil, err
			}
			var payload jsonImportPayload
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				return nil, fmt.Errorf("import json parse: %w", err)
			}
			imported, errors := 0, 0
			for _, e := range payload.Entities {
				entityErr := false
				for k, v := range e.Properties {
					if err := svc.SetEntityProperty(ctx, e.ID, e.Type, k, v); err != nil {
						entityErr = true
						errors++
						break
					}
				}
				if entityErr {
					continue
				}
				for _, r := range e.Relations {
					_, err := svc.AssertFact(ctx, AssertionInput{
						Triple: graph.Triple{
							Subject:     e.ID,
							SubjectType: e.Type,
							Predicate:   r.Predicate,
							Object:      r.Object,
							ObjectType:  r.ObjectType,
						},
						Source:     "import",
						Confidence: 0.9,
					})
					if err != nil {
						errors++
					}
				}
				imported++
			}
			return map[string]interface{}{
				"imported": imported,
				"errors":   errors,
				"total":    len(payload.Entities),
			}, nil
		},
	}
}

func buildImportCSV(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_import_csv",
		Description: "Import entities from CSV data. First row is header (property names), each subsequent row is an entity.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{"type": "string", "description": "CSV string (first row = headers)"},
				"type": map[string]interface{}{"type": "string", "description": "ObjectType name for all entities"},
			},
			"required": []string{"data", "type"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			data, err := toolparam.RequireString(params, "data")
			if err != nil {
				return nil, err
			}
			typeName, err := toolparam.RequireString(params, "type")
			if err != nil {
				return nil, err
			}
			reader := csv.NewReader(strings.NewReader(data))
			records, err := reader.ReadAll()
			if err != nil {
				return nil, fmt.Errorf("import csv parse: %w", err)
			}
			if len(records) < 2 {
				return nil, fmt.Errorf("import csv: need header row + at least one data row")
			}
			header := records[0]
			if len(header) < 2 {
				return nil, fmt.Errorf("import csv: need at least entity_id column + one property column")
			}
			imported, errors := 0, 0
			for _, row := range records[1:] {
				if len(row) < len(header) {
					errors++
					continue
				}
				entityID := row[0] // first column is entity_id
				entityErr := false
				for j := 1; j < len(header); j++ {
					if err := svc.SetEntityProperty(ctx, entityID, typeName, header[j], row[j]); err != nil {
						entityErr = true
						errors++
						break
					}
				}
				if !entityErr {
					imported++
				}
			}
			return map[string]interface{}{
				"imported": imported,
				"errors":   errors,
				"total":    len(records) - 1,
			}, nil
		},
	}
}

func buildFromMCP(svc OntologyService) *agent.Tool {
	return &agent.Tool{
		Name:        "ontology_from_mcp",
		Description: "Convert an MCP tool result into ontology entities and facts. Explicit mapping only.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tool_name":   map[string]interface{}{"type": "string", "description": "MCP tool name (used as relation source)"},
				"result_json": map[string]interface{}{"type": "string", "description": "JSON string of the MCP tool result"},
				"entity_type": map[string]interface{}{"type": "string", "description": "ObjectType to create for the result entity"},
				"predicate":   map[string]interface{}{"type": "string", "description": "Predicate linking entity to tool"},
			},
			"required": []string{"tool_name", "result_json", "entity_type", "predicate"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			toolName, err := toolparam.RequireString(params, "tool_name")
			if err != nil {
				return nil, err
			}
			resultJSON, err := toolparam.RequireString(params, "result_json")
			if err != nil {
				return nil, err
			}
			entityType, err := toolparam.RequireString(params, "entity_type")
			if err != nil {
				return nil, err
			}
			predicate, err := toolparam.RequireString(params, "predicate")
			if err != nil {
				return nil, err
			}

			var result map[string]interface{}
			if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
				return nil, fmt.Errorf("from mcp json parse: %w", err)
			}

			// Generate entity ID from tool name + first available ID field or hash.
			entityID := fmt.Sprintf("mcp:%s:%s", toolName, entityType)
			if id, ok := result["id"].(string); ok && id != "" {
				entityID = id
			}

			// Store each string field as a property.
			propsSet := 0
			for k, v := range result {
				strVal, ok := v.(string)
				if !ok {
					continue
				}
				if err := svc.SetEntityProperty(ctx, entityID, entityType, k, strVal); err != nil {
					continue // skip invalid properties
				}
				propsSet++
			}

			// Assert fact linking entity to tool.
			_, factErr := svc.AssertFact(ctx, AssertionInput{
				Triple: graph.Triple{
					Subject:     entityID,
					SubjectType: entityType,
					Predicate:   predicate,
					Object:      fmt.Sprintf("tool:%s", toolName),
					ObjectType:  "Tool",
				},
				Source:     "mcp",
				Confidence: 0.7,
			})

			resp := map[string]interface{}{
				"entity_id":      entityID,
				"properties_set": propsSet,
			}
			if factErr != nil {
				resp["fact_error"] = factErr.Error()
			} else {
				resp["fact_asserted"] = true
			}
			return resp, nil
		},
	}
}

// --- Helpers ---

func parseFilters(params map[string]interface{}) []PropertyFilter {
	raw, ok := params["filters"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var filters []PropertyFilter
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		f := PropertyFilter{
			Property: fmt.Sprintf("%v", m["property"]),
			Op:       FilterOp(fmt.Sprintf("%v", m["op"])),
			Value:    fmt.Sprintf("%v", m["value"]),
		}
		if f.Property != "" && f.Op != "" {
			filters = append(filters, f)
		}
	}
	return filters
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
