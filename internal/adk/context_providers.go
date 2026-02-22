package adk

import (
	"strings"
	"sync"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/knowledge"
)

// ToolRegistryAdapter adapts []*agent.Tool to knowledge.ToolRegistryProvider.
type ToolRegistryAdapter struct {
	tools []knowledge.ToolDescriptor
}

// NewToolRegistryAdapter creates a ToolRegistryAdapter from agent tools.
// The input slice is copied to prevent caller mutation.
func NewToolRegistryAdapter(tools []*agent.Tool) *ToolRegistryAdapter {
	descriptors := make([]knowledge.ToolDescriptor, len(tools))
	for i, t := range tools {
		descriptors[i] = knowledge.ToolDescriptor{
			Name:        t.Name,
			Description: t.Description,
		}
	}
	return &ToolRegistryAdapter{tools: descriptors}
}

// ListTools returns all available tools.
func (a *ToolRegistryAdapter) ListTools() []knowledge.ToolDescriptor {
	out := make([]knowledge.ToolDescriptor, len(a.tools))
	copy(out, a.tools)
	return out
}

// SearchTools returns tools whose name or description contains the query (case-insensitive).
func (a *ToolRegistryAdapter) SearchTools(query string, limit int) []knowledge.ToolDescriptor {
	if limit <= 0 {
		limit = len(a.tools)
	}
	queryLower := strings.ToLower(query)
	var result []knowledge.ToolDescriptor
	for _, t := range a.tools {
		if len(result) >= limit {
			break
		}
		if strings.Contains(strings.ToLower(t.Name), queryLower) ||
			strings.Contains(strings.ToLower(t.Description), queryLower) {
			result = append(result, t)
		}
	}
	return result
}

// RuntimeContextAdapter provides runtime session/system state.
type RuntimeContextAdapter struct {
	mu         sync.RWMutex
	sessionKey string
	channel    string

	toolCount  int
	encryption bool
	knowledge  bool
	memory     bool
}

// NewRuntimeContextAdapter creates a RuntimeContextAdapter with static system info.
func NewRuntimeContextAdapter(toolCount int, encryption, knowledgeEnabled, memoryEnabled bool) *RuntimeContextAdapter {
	return &RuntimeContextAdapter{
		toolCount:  toolCount,
		encryption: encryption,
		knowledge:  knowledgeEnabled,
		memory:     memoryEnabled,
		channel:    "direct",
	}
}

// SetSession updates the session key and derives the channel type.
func (a *RuntimeContextAdapter) SetSession(sessionKey string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sessionKey = sessionKey
	a.channel = deriveChannelType(sessionKey)
}

// GetRuntimeContext returns a snapshot of the current runtime context.
func (a *RuntimeContextAdapter) GetRuntimeContext() knowledge.RuntimeContext {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return knowledge.RuntimeContext{
		SessionKey:        a.sessionKey,
		ChannelType:       a.channel,
		ActiveToolCount:   a.toolCount,
		EncryptionEnabled: a.encryption,
		KnowledgeEnabled:  a.knowledge,
		MemoryEnabled:     a.memory,
	}
}

// deriveChannelType extracts the channel type from a session key.
// Session keys follow the pattern "channel:id:subid" (e.g., "telegram:123:456").
// Returns "direct" if the key has no recognized prefix.
func deriveChannelType(sessionKey string) string {
	if sessionKey == "" {
		return "direct"
	}
	prefix, _, found := strings.Cut(sessionKey, ":")
	if !found {
		return "direct"
	}
	switch prefix {
	case "telegram", "discord", "slack":
		return prefix
	default:
		return "direct"
	}
}
