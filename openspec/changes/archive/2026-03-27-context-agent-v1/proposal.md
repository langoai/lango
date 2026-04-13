## Why

The RetrievalCoordinator (Step 6) only has FactSearchAgent (keyword/FTS5). RAG and GraphRAG run as separate goroutines outside the coordinator, producing opaque section strings. Step 9 brings semantic/vector search inside the coordinator as a new `ContextSearchAgent`, establishing the principle: sqlite-vec/GraphRAG = "related context expansion", not "authoritative truth."

## What Changes

- Add `ContextSearchAgent` implementing `RetrievalAgent` interface — wraps RAGService for vector search
- Add `ContextSearchSource` narrow interface for testability
- Map RAG collections to context layers: knowledge→UserKnowledge, learning→AgentLearnings (v1 scope)
- Skip observation/reflection collections (memory budget boundary, deferred)
- Skip graph expansion (GraphNode has no content field, deferred)
- Score normalization: `max(0, 1.0 - distance)` — naturally lower than FTS5 scores (0-1 vs 1-10+)
- Enrich shadow comparison with factual-vs-new-context split logging
- Register ContextSearchAgent in coordinator when RAG is available

## Capabilities

### New Capabilities
- `context-agent-v1`: ContextSearchAgent for semantic/vector expansion within RetrievalCoordinator

### Modified Capabilities
- `agentic-retrieval`: Shadow comparison enhanced with factual/new-context split, initRetrievalCoordinator accepts embedding components

## Impact

- **NEW**: `internal/retrieval/context_agent.go` — ContextSearchAgent, ContextSearchSource, collectionToLayer, vectorDistanceToScore
- **NEW**: `internal/retrieval/context_agent_test.go` — 11 test cases
- **MODIFY**: `internal/retrieval/shadow.go` — factualLayers map, factual/new-context split in CompareShadowResults
- **MODIFY**: `internal/app/wiring_knowledge.go` — initRetrievalCoordinator expanded to register ContextSearchAgent
