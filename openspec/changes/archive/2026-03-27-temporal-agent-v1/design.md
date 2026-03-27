# Design: temporal-agent-v1

## Architecture

TemporalSearchAgent is a third RetrievalAgent in the coordinator, complementing FactSearchAgent (keyword relevance) and ContextSearchAgent (semantic similarity) with recency-based ranking.

```
RetrievalCoordinator.Retrieve(query, tokenBudget)
├── FactSearchAgent     → keyword search (FTS5/LIKE), score 1-10+
├── TemporalSearchAgent → recency search (updated_at DESC), score 0-1
├── ContextSearchAgent  → vector search (RAG), score 0-1
└── Dedup, Sort by Score DESC, Truncate
```

## Key Decisions

### KnowledgeEntry.UpdatedAt
Added `UpdatedAt time.Time` to the domain type. 6 conversion sites in store.go updated. Backward compatible — zero value for callers constructing entries.

### SearchRecentKnowledge
New store method: `WHERE is_latest=true ORDER BY updated_at DESC`. Does NOT use FTS5 (which orders by BM25). Optional keyword filter via existing `knowledgeKeywordPredicates`.

### Score Normalization
`recencyScore = max(0, 1.0 - hoursSinceUpdate / 168)` where 168 = 1 week. Natural priority: FTS5 (1-10+) > recency/vector (0-1).

### Content Enrichment
`[v5 | updated 3h ago] Original content` — gives LLM temporal context without Finding struct changes.

### Layer Coverage
v1: UserKnowledge only. Learnings lack version/is_latest — excluded.

### Wiring
Always registered (kStore always available). No conditional gating.

## Dedup Behavior
Same (Layer, Key) from FactSearch + TemporalSearch → highest score wins (usually FactSearch). Temporal agent's main value: surfacing recently-changed items NOT found by keyword search.
