package graph

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates tools for graph traversal and querying.
func BuildTools(gs Store) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "graph_traverse",
			Description: "Traverse the knowledge graph from a start node using BFS. Returns related triples up to the specified depth.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_node": map[string]interface{}{"type": "string", "description": "The node ID to start traversal from"},
					"max_depth":  map[string]interface{}{"type": "integer", "description": "Maximum traversal depth (default: 2)"},
					"predicates": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Filter by predicate types (empty = all)"},
				},
				"required": []string{"start_node"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				startNode, err := toolparam.RequireString(params, "start_node")
				if err != nil {
					return nil, err
				}
				maxDepth := toolparam.OptionalInt(params, "max_depth", 2)
				predicates := toolparam.StringSlice(params, "predicates")
				triples, err := gs.Traverse(ctx, startNode, maxDepth, predicates)
				if err != nil {
					return nil, fmt.Errorf("graph traverse: %w", err)
				}
				return map[string]interface{}{"triples": triples, "count": len(triples)}, nil
			},
		},
		{
			Name:        "graph_query",
			Description: "Query the knowledge graph by subject or object node. Returns matching triples.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"subject":   map[string]interface{}{"type": "string", "description": "Subject node to query by"},
					"object":    map[string]interface{}{"type": "string", "description": "Object node to query by"},
					"predicate": map[string]interface{}{"type": "string", "description": "Optional predicate filter (used with subject)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				subject := toolparam.OptionalString(params, "subject", "")
				object := toolparam.OptionalString(params, "object", "")
				predicate := toolparam.OptionalString(params, "predicate", "")

				if subject == "" && object == "" {
					return nil, fmt.Errorf("either subject or object is required")
				}

				var triples []Triple
				var err error
				if subject != "" && predicate != "" {
					triples, err = gs.QueryBySubjectPredicate(ctx, subject, predicate)
				} else if subject != "" {
					triples, err = gs.QueryBySubject(ctx, subject)
				} else {
					triples, err = gs.QueryByObject(ctx, object)
				}
				if err != nil {
					return nil, fmt.Errorf("graph query: %w", err)
				}
				return map[string]interface{}{"triples": triples, "count": len(triples)}, nil
			},
		},
	}
}
