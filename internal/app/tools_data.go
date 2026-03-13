package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/session"
	toolpayment "github.com/langoai/lango/internal/tools/payment"
	x402pkg "github.com/langoai/lango/internal/x402"
)

// buildGraphTools creates tools for graph traversal and querying.
func buildGraphTools(gs graph.Store) []*agent.Tool {
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

				var triples []graph.Triple
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

// buildRAGTools creates tools for RAG retrieval.
func buildRAGTools(ragSvc *embedding.RAGService) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "rag_retrieve",
			Description: "Retrieve semantically similar content from the knowledge base using vector search.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":       map[string]interface{}{"type": "string", "description": "The search query"},
					"limit":       map[string]interface{}{"type": "integer", "description": "Maximum results to return (default: 5)"},
					"collections": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Filter by collections (e.g., knowledge, observation)"},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query, err := toolparam.RequireString(params, "query")
				if err != nil {
					return nil, err
				}
				limit := toolparam.OptionalInt(params, "limit", 5)
				collections := toolparam.StringSlice(params, "collections")
				sessionKey := session.SessionKeyFromContext(ctx)
				results, err := ragSvc.Retrieve(ctx, query, embedding.RetrieveOptions{
					Limit:       limit,
					Collections: collections,
					SessionKey:  sessionKey,
				})
				if err != nil {
					return nil, fmt.Errorf("rag retrieve: %w", err)
				}
				return map[string]interface{}{"results": results, "count": len(results)}, nil
			},
		},
	}
}

// buildMemoryAgentTools creates tools for observational memory management.
func buildMemoryAgentTools(ms *memory.Store) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "memory_list_observations",
			Description: "List observations for a session. Returns compressed notes from conversation history.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key to list observations for (uses current session if empty)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey := toolparam.OptionalString(params, "session_key", session.SessionKeyFromContext(ctx))
				observations, err := ms.ListObservations(ctx, sessionKey)
				if err != nil {
					return nil, fmt.Errorf("list observations: %w", err)
				}
				return map[string]interface{}{"observations": observations, "count": len(observations)}, nil
			},
		},
		{
			Name:        "memory_list_reflections",
			Description: "List reflections for a session. Reflections are condensed observations across time.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key to list reflections for (uses current session if empty)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey := toolparam.OptionalString(params, "session_key", session.SessionKeyFromContext(ctx))
				reflections, err := ms.ListReflections(ctx, sessionKey)
				if err != nil {
					return nil, fmt.Errorf("list reflections: %w", err)
				}
				return map[string]interface{}{"reflections": reflections, "count": len(reflections)}, nil
			},
		},
	}
}

// buildPaymentTools creates blockchain payment tools.
func buildPaymentTools(pc *paymentComponents, x402Interceptor *x402pkg.Interceptor) []*agent.Tool {
	return toolpayment.BuildTools(pc.service, pc.limiter, pc.secrets, pc.chainID, x402Interceptor)
}

// buildLibrarianTools creates proactive librarian agent tools.
func buildLibrarianTools(is *librarian.InquiryStore) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "librarian_pending_inquiries",
			Description: "List pending knowledge inquiries for the current session",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key (uses current session if empty)"},
					"limit":       map[string]interface{}{"type": "integer", "description": "Maximum results (default: 5)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey := toolparam.OptionalString(params, "session_key", session.SessionKeyFromContext(ctx))
				limit := toolparam.OptionalInt(params, "limit", 5)
				inquiries, err := is.ListPendingInquiries(ctx, sessionKey, limit)
				if err != nil {
					return nil, fmt.Errorf("list pending inquiries: %w", err)
				}
				return map[string]interface{}{"inquiries": inquiries, "count": len(inquiries)}, nil
			},
		},
		{
			Name:        "librarian_dismiss_inquiry",
			Description: "Dismiss a pending knowledge inquiry that the user does not want to answer",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"inquiry_id": map[string]interface{}{"type": "string", "description": "UUID of the inquiry to dismiss"},
				},
				"required": []string{"inquiry_id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				idStr, err := toolparam.RequireString(params, "inquiry_id")
				if err != nil {
					return nil, err
				}
				id, err := uuid.Parse(idStr)
				if err != nil {
					return nil, fmt.Errorf("invalid inquiry_id: %w", err)
				}
				if err := is.DismissInquiry(ctx, id); err != nil {
					return nil, fmt.Errorf("dismiss inquiry: %w", err)
				}
				return map[string]interface{}{
					"status":  "dismissed",
					"message": fmt.Sprintf("Inquiry %s dismissed", idStr),
				}, nil
			},
		},
	}
}
