## Why

Knowledge entries have `relevance_score` (default 1.0) used for LIKE search ordering and coordinator merge priority, but the score never changes after creation. Frequently injected items should rank higher; stale items should drift down.

## What Changes

- `RelevanceAdjuster` subscribes to `ContextInjectedEvent`, boosts injected user_knowledge items, decays all scores periodically
- Default mode: shadow (log decisions, no DB writes). Active mode writes to DB.
- Warmup period (50 turns), score cap (5.0), floor (0.1), rollback toggle
- Store gains `BoostRelevanceScore`, `DecayAllRelevanceScores`, `ResetAllRelevanceScores`
- Config: `retrieval.autoAdjust.*` nested under RetrievalConfig

## Effect Scope

Primarily affects LIKE fallback search path + coordinator merge priority. FTS5 BM25 ranking is unaffected. Signal source: old knowledge path structured items only (ContextInjectedEvent.Items).
