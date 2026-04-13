## Context

Steps 1-5 complete. Knowledge has versioning (Step 4), 6-category taxonomy with temporal tags (Step 5), content-dedup, FTS5+LIKE search. The retrieval pipeline is monolithic — single ContextRetriever handles all layers with no scoring, provenance, or pluggable agents. ContextItem.Score exists but is unused.

## Goals / Non-Goals

**Goals:**
- Agent-based retrieval with pluggable RetrievalAgent interface
- Scored search with FTS5 BM25 rank preservation
- Shadow mode for safe quality comparison
- FactSearchAgent as first concrete agent (factual layers only)
- ContextItem.Score populated for the first time

**Non-Goals:**
- Replacing the existing ContextRetriever (shadow mode only in v1)
- Semantic/vector search as primary path (later step)
- Non-factual layer coverage (tools, runtime, skills, inquiries)
- TurnID/RequestID tracking in events (later step)

## Decisions

### D1. Scored APIs with normalized score + SearchSource
Add `SearchKnowledgeScored` / `SearchLearningsScored` returning `ScoredKnowledgeEntry{Entry, Score, SearchSource}`. FTS5 BM25 rank is negated (`Score = -rank`) for higher=better convention. LIKE path uses `RelevanceScore` directly. `SearchSource` is `"fts5"` or `"like"`. ExternalRefs: Score=0, SearchSource="like".

Existing unscored APIs unchanged — avoids touching 10+ call sites.

### D2. Dedup by (Layer, Key) composite
Same key in different layers = independent findings. Same (layer, key) from multiple agents = keep highest Score. Knowledge `Key`, Learning `Trigger`, ExternalRef `Name` are separate namespaces.

### D3. Budget ownership: coordinator truncates, agents don't
`RetrievalAgent.Search(ctx, query, limit)` receives item-count limit only. Coordinator handles token budget via `TruncateFindings()` after merge. Single truncation point.

### D4. Shadow mode: fire-and-forget goroutine
Launches AFTER `g.Wait()` completes (existing retrieval done). Uses `context.Background()` for independent lifetime. Logs overlap % and unique finding counts. Does not block LLM call.

### D5. ToRetrievalResult propagates Score to ContextItem.Score
`ContextItem.Score = Finding.Score` during conversion. First time this field has meaning.

### D6. FactSearchAgent uses narrow FactSearchSource interface
Decouples `internal/retrieval/` from `*knowledge.Store` concrete type. Enables mock testing. Interface: `SearchKnowledgeScored`, `SearchLearningsScored`, `SearchExternalRefs`.

### D7. v1 coordinator covers factual layers only
3 layers: UserKnowledge, AgentLearnings, ExternalKnowledge. Tools/Runtime/Skills/Inquiries remain in ContextRetriever. Shadow mode compares factual retrieval quality only.

## Risks / Trade-offs

**[Risk] Shadow goroutine lifetime** → Uses `context.Background()`, may outlive request. Acceptable for logging-only work; no state mutation.

**[Trade-off] Score normalization is approximate** → `-rank` and `RelevanceScore` are different scales. Sufficient for within-source ranking; cross-source comparison is informational only.

**[Trade-off] v1 doesn't cover all layers** → Intentional scope control. Full coverage deferred to shadow=false migration step.
