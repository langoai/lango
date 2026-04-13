// Package memory provides observational memory for conversations.
//
// It extracts observations from conversation turns (Observer) and synthesizes
// higher-level reflections from accumulated observations (Reflector).
// Memory is session-scoped and temporal — it captures what happened during
// a conversation, not persistent agent knowledge.
//
// Buffer manages async processing with configurable token thresholds.
// GraphHooks generates temporal/session triples for the graph store.
//
// Related packages:
//   - agentmemory: per-agent persistent memory (cross-session)
//   - knowledge: user-contributed knowledge store with multi-layer retrieval
//   - learning: pattern extraction from tool execution results
//   - graph: triple store for semantic relationships
package memory
