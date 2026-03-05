package app

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/toolchain"
)

// buildAgentMemoryTools creates tools that let agents save, recall, and forget
// their own persistent memories (patterns, preferences, facts, skills).
func buildAgentMemoryTools(store agentmemory.Store) []*agent.Tool {
	return []*agent.Tool{
		{
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
				key, _ := params["key"].(string)
				content, _ := params["content"].(string)
				if key == "" || content == "" {
					return nil, fmt.Errorf("key and content are required")
				}

				kind := agentmemory.KindFact
				if k, ok := params["kind"].(string); ok && k != "" {
					kind = agentmemory.MemoryKind(k)
				}

				confidence := 0.5
				if c, ok := params["confidence"].(float64); ok && c >= 0 && c <= 1 {
					confidence = c
				}

				var tags []string
				if rawTags, ok := params["tags"].([]interface{}); ok {
					for _, t := range rawTags {
						if s, ok := t.(string); ok {
							tags = append(tags, s)
						}
					}
				}

				agentName := toolchain.AgentNameFromContext(ctx)
				if agentName == "" {
					agentName = "default"
				}

				entry := &agentmemory.Entry{
					AgentName:  agentName,
					Key:        key,
					Content:    content,
					Kind:       kind,
					Scope:      agentmemory.ScopeInstance,
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
		},
		{
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
				query, _ := params["query"].(string)
				if query == "" {
					return nil, fmt.Errorf("query is required")
				}

				limit := 10
				if l, ok := params["limit"].(float64); ok && l > 0 {
					limit = int(l)
				}

				agentName := toolchain.AgentNameFromContext(ctx)
				if agentName == "" {
					agentName = "default"
				}

				kindStr, _ := params["kind"].(string)

				var results []*agentmemory.Entry
				var err error

				if kindStr != "" {
					results, err = store.Search(agentName, agentmemory.SearchOptions{
						Query: query,
						Kind:  agentmemory.MemoryKind(kindStr),
						Limit: limit,
					})
				} else {
					results, err = store.SearchWithContext(agentName, query, limit)
				}
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
		},
		{
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
				key, _ := params["key"].(string)
				if key == "" {
					return nil, fmt.Errorf("key is required")
				}

				agentName := toolchain.AgentNameFromContext(ctx)
				if agentName == "" {
					agentName = "default"
				}

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
		},
	}
}
