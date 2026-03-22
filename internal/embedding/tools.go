package embedding

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildRAGTools creates tools for RAG retrieval.
func BuildRAGTools(ragSvc *RAGService) []*agent.Tool {
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
				results, err := ragSvc.Retrieve(ctx, query, RetrieveOptions{
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
