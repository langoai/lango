package orchestration

import (
	"strings"

	"github.com/langowarny/lango/internal/agent"
)

// RoleToolSet defines which tools belong to each sub-agent role.
type RoleToolSet struct {
	Executor      []*agent.Tool
	Researcher    []*agent.Tool
	Planner       []*agent.Tool
	MemoryManager []*agent.Tool
}

// executorPrefixes are tool name prefixes assigned to the Executor sub-agent.
var executorPrefixes = []string{"exec", "fs_", "browser_", "crypto_", "skill_", "payment_"}

// researcherPrefixes are tool name prefixes assigned to the Researcher sub-agent.
var researcherPrefixes = []string{"search_", "rag_", "graph_", "save_knowledge", "save_learning"}

// memoryPrefixes are tool name prefixes assigned to the Memory Manager sub-agent.
var memoryPrefixes = []string{"memory_", "observe_", "reflect_"}

// PartitionTools splits tools into role-specific sets based on tool name prefixes.
// Tools not matching any prefix are assigned to the Executor (default).
func PartitionTools(tools []*agent.Tool) RoleToolSet {
	var rs RoleToolSet
	for _, t := range tools {
		switch {
		case matchesPrefix(t.Name, researcherPrefixes):
			rs.Researcher = append(rs.Researcher, t)
		case matchesPrefix(t.Name, memoryPrefixes):
			rs.MemoryManager = append(rs.MemoryManager, t)
		case matchesPrefix(t.Name, executorPrefixes):
			rs.Executor = append(rs.Executor, t)
		default:
			// Unmatched tools default to Executor.
			rs.Executor = append(rs.Executor, t)
		}
	}
	// Planner has no tools (LLM-only reasoning).
	return rs
}

// matchesPrefix returns true if name starts with any of the given prefixes.
func matchesPrefix(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// capabilityMap maps tool name prefixes to human-readable capability descriptions.
// This prevents the LLM from seeing raw tool names (e.g. "browser_navigate")
// and hallucinating agent names like "browser_agent".
var capabilityMap = map[string]string{
	"exec":           "command execution",
	"fs_":            "file operations",
	"browser_":       "web browsing",
	"crypto_":        "cryptography",
	"skill_":         "skill management",
	"payment_":       "blockchain payments (USDC on Base)",
	"search_":        "information search",
	"rag_":           "knowledge retrieval (RAG)",
	"graph_":         "knowledge graph traversal",
	"save_knowledge": "knowledge persistence",
	"save_learning":  "learning persistence",
	"memory_":        "memory storage and recall",
	"observe_":       "event observation",
	"reflect_":       "reflection and summarization",
}

// toolCapability returns a human-readable capability for a tool name based
// on its prefix. Returns an empty string if no mapping exists.
func toolCapability(name string) string {
	for prefix, cap := range capabilityMap {
		if strings.HasPrefix(name, prefix) {
			return cap
		}
	}
	return ""
}

// capabilityDescription builds a deduplicated, comma-separated capability
// string from a tool list. Tool names are mapped to natural-language
// descriptions so the LLM never sees raw tool name prefixes.
func capabilityDescription(tools []*agent.Tool) string {
	seen := make(map[string]struct{}, len(tools))
	var caps []string
	for _, t := range tools {
		c := toolCapability(t.Name)
		if c == "" {
			c = "general actions"
		}
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			caps = append(caps, c)
		}
	}
	return strings.Join(caps, ", ")
}
