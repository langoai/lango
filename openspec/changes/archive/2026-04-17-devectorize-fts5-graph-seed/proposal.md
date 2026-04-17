## Why

Lango's current retrieved-context pipeline still carries vector-specific runtime paths and tooling even though the knowledge/session recall experience is already FTS5-centric. This increases maintenance cost, keeps dead configuration surface alive, and forces GraphRAG to depend on the embedding subsystem for its phase-1 seed results.

## What Changes

- Remove the standalone vector context path from the app and ADK wiring.
- Rewire GraphRAG phase 1 to use hydrated knowledge FTS5 results instead of embedding-backed retrieval.
- Remove the `rag_retrieve` tool and the semantic/vector-specific context-search agent.
- Rename the prompt/event surface from RAG-specific wording to retrieved-context wording where behavior is no longer vector-specific.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `graph-rag`: phase-1 seeding changes from vector retrieval to generic content retrieval backed by FTS5 knowledge hits.
- `context-retriever`: the context-aware adapter no longer runs a standalone vector retrieval path and injects a retrieved-context section sourced from GraphRAG/recall.
- `agentic-retrieval`: the coordinator no longer registers the vector-backed context-search agent.
- `retrieval-feedback-observability`: the event payload renames `RAGTokens` to `RetrievedTokens`.

## Impact

- Affected code: ADK context assembly, graph wiring, retrieval coordinator wiring, graph prompt assembly, tool registration, and knowledge-save hooks.
- Removed runtime surface: `rag_retrieve` tool, `ContextSearchAgent`, embedding wiring path in the application bootstrap.
- No user config key changes in this step: `context.allocation.rag` and existing retrieval config remain accepted for compatibility.
