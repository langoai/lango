## Context

Step 6 added FactSearchAgent (FTS5/LIKE) to RetrievalCoordinator. RAG/GraphRAG run separately. Step 9 integrates vector search into the coordinator.

## Goals / Non-Goals

**Goals:**
- ContextSearchAgent as second agent in coordinator, providing semantic expansion
- Score natural priority: FTS5 > vector (no explicit discount)
- Shadow comparison with factual/new-context split

**Non-Goals:**
- Graph expansion (GraphNode has no content field — deferred)
- observation/reflection collections (memory budget boundary — deferred)
- Replacing existing RAG goroutine (shadow mode, existing path untouched)

## Decisions

### DD1: Narrow interface ContextSearchSource
`ContextSearchSource` wraps `*embedding.RAGService`. Avoids direct dependency for testability. `retrieval → embedding` import is safe (no reverse).

### DD2: v1 layer scope — knowledge + learning only
observation/reflection are memory-section signals. Including them would violate budget ownership. Deferred to future step.

### DD3: Score natural priority
Vector distance inverted: `max(0, 1.0 - distance)` → 0-1 range. FTS5 BM25 negated → 1-10+ range. Natural ordering means no discount needed.

### DD4: Shadow factual/new-context split
Prevents structural dilution when new agents add items. factualLayers tracks UserKnowledge + AgentLearnings + ExternalKnowledge.

## Risks / Trade-offs

- **[Shadow mode only]** → ContextSearchAgent doesn't affect user-visible output until shadow mode is turned off.
- **[Collection filtering]** → RAGService is called with `Collections: ["knowledge", "learning"]` — other collections not searched.
