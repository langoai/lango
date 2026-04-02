// Package knowledge provides the primary knowledge store and multi-layer
// context retrieval system.
//
// The store persists user-contributed knowledge entries (rules, definitions,
// preferences, facts, patterns, corrections) with versioning and relevance
// scoring. ContextRetriever implements 8-layer retrieval:
//
//  1. Runtime context
//  2. Tool registry
//  3. User knowledge
//  4. Skill patterns
//  5. External knowledge
//  6. Agent learnings
//  7. Pending inquiries
//  8. Conversation analysis
//
// SetEmbedCallback and SetGraphCallback wire async processing without
// creating import cycles.
//
// Related packages:
//   - memory: session-scoped observations (feeds into knowledge via learning)
//   - agentmemory: per-agent persistent memory
//   - learning: extracts patterns from tool results into knowledge
//   - embedding: vector embeddings for semantic retrieval
//   - graph: triple store for relationship traversal
//   - librarian: proactive knowledge gap analysis
package knowledge
