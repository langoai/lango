// Package agentmemory provides per-agent persistent memory.
//
// Unlike the memory package (session-scoped observations), agentmemory stores
// entries that persist across sessions and are scoped by agent name. Each entry
// has a kind (Pattern, Preference, Fact, Skill) and a confidence score.
//
// Memory scopes control visibility:
//   - Instance: visible only to one agent instance
//   - Type: visible to all instances of the same agent type
//   - Global: visible to all agents
//
// Related packages:
//   - memory: session-scoped observational memory (temporal, ephemeral)
//   - knowledge: user-contributed knowledge store
//   - learning: feeds extracted patterns into agentmemory
package agentmemory
