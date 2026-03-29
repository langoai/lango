package agentmemory

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildTools creates tools that let agents save, recall, and forget
// their own persistent memories (patterns, preferences, facts, skills).
func BuildTools(store Store) []*agent.Tool {
	return []*agent.Tool{
		buildSaveTool(store),
		buildRecallTool(store),
		buildForgetTool(store),
	}
}

func buildSaveTool(store Store) *agent.Tool {
	return &agent.Tool{
		Name:        "memory_agent_save",
		Description: "Save a memory entry for this agent (pattern, preference, fact, or skill). Memories persist across sessions.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key":        map[string]interface{}{"type": "string", "description": "Unique key for this memory entry"},
				"content":    map[string]interface{}{"type": "string", "description": "The memory content to save"},
				"kind":       map[string]interface{}{"type": "string", "description": "Memory kind: pattern, preference, fact, or skill", "enum": []string{"pattern", "preference", "fact", "skill"}},
				"tags":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Optional tags for categorization"},
				"confidence": map[string]interface{}{"type": "number", "description": "Confidence score 0.0-1.0 (default: 0.5)"},
			},
			"required": []string{"key", "content"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			key, err := toolparam.RequireString(params, "key")
			if err != nil {
				return nil, err
			}
			content, err := toolparam.RequireString(params, "content")
			if err != nil {
				return nil, err
			}

			kind := KindFact
			if k := toolparam.OptionalString(params, "kind", ""); k != "" {
				kind = MemoryKind(k)
				if !kind.Valid() {
					return nil, fmt.Errorf("invalid memory kind %q: must be pattern, preference, fact, or skill", k)
				}
			}

			confidence := toolparam.OptionalFloat64(params, "confidence", 0.5)
			if confidence < 0 || confidence > 1 {
				confidence = 0.5
			}

			tags := toolparam.StringSlice(params, "tags")

			agentName := agentNameOrDefault(ctx)

			entry := &Entry{
				AgentName:  agentName,
				Key:        key,
				Content:    content,
				Kind:       kind,
				Scope:      ScopeInstance,
				Confidence: confidence,
				Tags:       tags,
			}

			if err := store.Save(entry); err != nil {
				return nil, fmt.Errorf("save agent memory: %w", err)
			}

			return map[string]interface{}{
				"status":  "saved",
				"key":     key,
				"agent":   agentName,
				"message": fmt.Sprintf("Memory '%s' saved for agent '%s'", key, agentName),
			}, nil
		},
	}
}

func buildRecallTool(store Store) *agent.Tool {
	return &agent.Tool{
		Name:        "memory_agent_recall",
		Description: "Recall memories for this agent. Searches across instance and global scopes.",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{"type": "string", "description": "Search query to find relevant memories"},
				"limit": map[string]interface{}{"type": "integer", "description": "Maximum results to return (default: 10)"},
				"kind":  map[string]interface{}{"type": "string", "description": "Optional kind filter: pattern, preference, fact, or skill", "enum": []string{"pattern", "preference", "fact", "skill"}},
			},
			"required": []string{"query"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			query, err := toolparam.RequireString(params, "query")
			if err != nil {
				return nil, err
			}

			limit := toolparam.OptionalInt(params, "limit", 10)

			agentName := agentNameOrDefault(ctx)

			var kind MemoryKind
			if k := toolparam.OptionalString(params, "kind", ""); k != "" {
				kind = MemoryKind(k)
				if !kind.Valid() {
					return nil, fmt.Errorf("invalid memory kind %q: must be pattern, preference, fact, or skill", k)
				}
			}

			results, err := store.SearchWithContextOptions(agentName, SearchOptions{
				Query: query,
				Kind:  kind,
				Limit: limit,
			})
			if err != nil {
				return nil, fmt.Errorf("recall agent memory: %w", err)
			}

			// Increment use count for returned results.
			for _, r := range results {
				_ = store.IncrementUseCount(r.AgentName, r.Key)
			}

			return map[string]interface{}{
				"results": results,
				"count":   len(results),
				"agent":   agentName,
			}, nil
		},
	}
}

func buildForgetTool(store Store) *agent.Tool {
	return &agent.Tool{
		Name:        "memory_agent_forget",
		Description: "Forget (delete) a specific memory entry for this agent.",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{"type": "string", "description": "The key of the memory entry to forget"},
			},
			"required": []string{"key"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			key, err := toolparam.RequireString(params, "key")
			if err != nil {
				return nil, err
			}

			agentName := agentNameOrDefault(ctx)

			if err := store.Delete(agentName, key); err != nil {
				return nil, fmt.Errorf("forget agent memory: %w", err)
			}

			return map[string]interface{}{
				"status":  "forgotten",
				"key":     key,
				"agent":   agentName,
				"message": fmt.Sprintf("Memory '%s' forgotten for agent '%s'", key, agentName),
			}, nil
		},
	}
}

// agentNameOrDefault extracts the agent name from context, defaulting to "default".
func agentNameOrDefault(ctx context.Context) string {
	if name := ctxkeys.AgentNameFromContext(ctx); name != "" {
		return name
	}
	return "default"
}
