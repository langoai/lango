## Decisions

### DD1: Separate from FeedbackProcessor
RelevanceAdjuster is a separate subscriber. FeedbackProcessor stays read-only for logging.

### DD2: Shadow-first, process-local warmup
Default mode="shadow". Warmup counter is atomic.Int64 (process-local, resets on restart).

### DD3: Signal scope — old knowledge path only
ContextInjectedEvent.Items = old ContextRetriever structured items. RAG/memory/coordinator are not in Items.

### DD4: Decay-before-boost ordering
Prevents "boost then immediately decay" anomaly on decay-interval turns.

### DD5: Global cross-session decay
Decay is not per-session. Any session's turn contributes to the global decay trigger counter.

### DD6: Turn-level dedup
Same key appearing multiple times in one event is boosted once only.

### DD7: Effect scope — LIKE fallback + merge priority
FTS5 BM25 ranking unaffected. relevance_score matters in LIKE fallback ORDER BY and SearchKnowledgeScored LIKE branch Score field.
