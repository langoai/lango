## Why

The existing retrieval pipeline (`ContextRetriever.Retrieve()`) uses a monolithic approach: one query, same search across all layers, no scoring, no provenance tracking. There is no way to plug in specialized search agents, compare retrieval strategies, or measure retrieval quality. The `ContextItem.Score` field has been unused since creation.

This step introduces an agent-based retrieval architecture with `Finding`, `RetrievalAgent`, `RetrievalCoordinator`, and `FactSearchAgent` — running in shadow mode (`shadow=true` default) alongside the existing path to enable quality comparison without risk.

## What Changes

- New `internal/retrieval/` package with Finding type (scored, provenance-tracked), RetrievalAgent interface, RetrievalCoordinator (parallel agent execution, (Layer,Key) dedup, token-budget truncation), FactSearchAgent (wraps scored store APIs), shadow comparison logging
- New scored store APIs: `SearchKnowledgeScored()` and `SearchLearningsScored()` on `knowledge.Store` — preserve FTS5 BM25 rank (negated for higher=better normalization) and SearchSource ("fts5"/"like")
- New `ScoredKnowledgeEntry` and `ScoredLearningEntry` types with Score + SearchSource fields
- `FactSearchSource` narrow interface in retrieval package — satisfied by `knowledge.Store`, enables mock testing
- `RetrievalCoordinator.ToRetrievalResult()` populates `ContextItem.Score` for the first time
- Shadow mode: coordinator runs as fire-and-forget goroutine after existing retrieval completes, logs overlap metrics via `CompareShadowResults()`
- Config: `retrieval.enabled` (default false), `retrieval.shadow` (default true)
- v1 coordinator covers 3 factual layers only (UserKnowledge, AgentLearnings, ExternalKnowledge)

## Capabilities

### New Capabilities
- `agentic-retrieval`: Finding type, RetrievalAgent interface, RetrievalCoordinator, FactSearchAgent, shadow comparison, scored search APIs

### Modified Capabilities
- `knowledge-store`: SearchKnowledgeScored/SearchLearningsScored with BM25 rank preservation and SearchSource tracking

## Impact

- **NEW**: `internal/retrieval/` — finding.go, agent.go, coordinator.go, fact_agent.go, shadow.go, *_test.go (7 files)
- **MODIFY**: `internal/knowledge/store.go` — SearchKnowledgeScored, SearchLearningsScored methods
- **MODIFY**: `internal/knowledge/types.go` — ScoredKnowledgeEntry, ScoredLearningEntry types
- **MODIFY**: `internal/config/types.go` — RetrievalConfig struct
- **MODIFY**: `internal/config/loader.go` — retrieval defaults
- **MODIFY**: `internal/adk/context_model.go` — coordinator field, WithCoordinator(), shadow goroutine
- **MODIFY**: `internal/app/wiring_knowledge.go` — initRetrievalCoordinator()
- **MODIFY**: `internal/app/wiring.go` — coordinator wiring into adapter
- **MODIFY**: `README.md` — Retrieval Coordinator section
- **No impact on**: existing ContextRetriever (unchanged, remains primary), SearchKnowledge/SearchLearnings (unchanged)
