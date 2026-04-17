## Context

The codebase already uses FTS5 as the primary searchable substrate for knowledge and session recall. The remaining vector-specific path exists in two places: a standalone context injection path in the ADK adapter and the phase-1 seed adapter used by GraphRAG. The implementation in this change removes both while keeping the existing graph traversal and FTS5-backed stores intact.

## Goals / Non-Goals

**Goals:**
- Remove standalone vector retrieval from prompt assembly.
- Keep GraphRAG behavior, but seed it from hydrated knowledge FTS5 results.
- Remove vector-only tool/retrieval surfaces that no longer have an active runtime path.
- Keep config compatibility for existing `context.allocation.rag` keys.

**Non-Goals:**
- Removing the entire `internal/embedding` package from the repository.
- Introducing brokered storage, payload encryption, or modernc driver changes.
- Renaming user-facing config keys in this change.

## Decisions

- Use a generic `graph.ContentRetriever` interface instead of a vector-specific adapter.
  Rationale: GraphRAG phase 2 only needs `(collection, sourceID, content)` seeds and does not require vector semantics.
- Use `knowledge.SearchKnowledgeScored` for GraphRAG phase 1.
  Rationale: it already provides hydrated top-k results with FTS5/BM25 ordering and LIKE fallback.
- Remove `ContextSearchAgent` entirely rather than replacing it.
  Rationale: the factual coordinator remains useful with Fact + Temporal agents, and GraphRAG still provides a second retrieval path for prompt enrichment.
- Rename prompt/event wording from RAG-specific to retrieved-context wording only where behavior is now backend-agnostic.
  Rationale: avoid claiming semantic/vector retrieval when the active path is FTS5-backed.

## Risks / Trade-offs

- [Risk] Dead embedding package code remains in-tree. → Mitigation: runtime wiring and tools no longer expose it; full package removal is deferred to later changes.
- [Risk] Prompt/test expectations may still assert RAG-specific labels. → Mitigation: update prompt/event labels and parity tests in the same change.
- [Risk] GraphRAG seed quality may differ from vector results. → Mitigation: use scored hydrated knowledge results and keep phase-2 graph expansion unchanged.
