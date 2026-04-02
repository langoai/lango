// Package learning provides self-learning from tool execution results.
//
// Engine analyzes tool execution outcomes to extract reusable patterns,
// categorizes errors (Timeout, Permission, Provider, Tool, General), and
// persists learnings to the knowledge store and agent memory.
//
// GraphEngine extends Engine with graph triple generation and confidence
// propagation. ConversationAnalyzer and SessionLearner analyze conversation
// history for deeper pattern extraction. AnalysisBuffer batches analysis
// with turn/token thresholds.
//
// Related packages:
//   - knowledge: primary store where learnings are persisted
//   - agentmemory: per-agent memory where patterns are stored
//   - memory: observational memory (source of conversation data)
//   - graph: triple store updated with extracted relationships
package learning
