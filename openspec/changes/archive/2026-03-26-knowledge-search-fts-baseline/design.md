## Context

Knowledge and learning search currently uses Ent-generated LIKE predicates (`ContentContains`/`KeyContains` per keyword, OR-combined). This works for small corpora but cannot rank results by query relevance (only static `RelevanceScore`), handle phrase queries, or do prefix matching. The search module lives entirely within `internal/knowledge/store.go`.

This is Step 2 of the Context Engineering Endgame Roadmap. Step 1 (`context-foundation-zero-config`) established `LoadResult`, `FeatureStatus`, `StatusCollector`, and `contextProfile` as the common foundation.

**Current search flow:**
1. `ContextAwareModelAdapter.GenerateContent()` extracts last user message
2. `ContextRetriever.Retrieve()` calls `Store.SearchKnowledge()` / `Store.SearchLearnings()`
3. Store methods split query into keywords, create per-keyword LIKE predicates, ORDER BY relevance_score DESC

**Constraint (Cross-Cutting Principle 4):** The shared FTS5 package SHALL NOT know domain semantics. `is_latest` or any future temporal filter is enforced by the caller's write-time indexing policy (insert/delete), not by the search package.

## Goals / Non-Goals

**Goals:**
- Provide FTS5-based full-text search with BM25 ranking for knowledge and learning entries
- Auto-detect FTS5 availability at runtime; fall back to existing LIKE search transparently
- Establish `internal/search/` as a shared, domain-agnostic search substrate
- Set search performance baseline (benchmarks for 1k/10k corpora)

**Non-Goals:**
- Temporal versioning or `is_latest` filtering (Step 4)
- Token budget management for search results (Step 3)
- Semantic/vector search integration (Step 9+)
- Category taxonomy changes (Step 5)
- New config keys — FTS5 is auto-detected, not configured

## Decisions

### D1: Separate `internal/search/` package (not `internal/knowledge/fts5.go`)

FTS5 goes in a new `internal/search/` package, not inside `internal/knowledge/`.

**Rationale:** Both knowledge and learning stores need FTS5. Putting it in knowledge would create a knowledge→learning dependency or force code duplication. A shared substrate avoids both.

**Alternative considered:** `internal/knowledge/fts5.go` — simpler initially, but Step 4 (temporal) and future Learning temporal would require extracting it anyway.

### D2: FTS5 virtual tables via raw SQL, not Ent

FTS5 tables are created and queried via `database/sql` directly, not through Ent schema definitions.

**Rationale:** Ent does not support FTS5 virtual tables. Attempting to model them as Ent schemas would fight the ORM. Raw SQL for FTS5 is standard practice.

### D3: Injection via `SetFTS5Index()`, not constructor parameter

The `FTS5Index` is injected into `knowledge.Store` via a setter (`SetFTS5Index()`), not a constructor parameter.

**Rationale:** The knowledge store is created before FTS5 probing happens (Ent client must exist first for the raw DB handle). Setter injection matches the existing `SetEventBus()` pattern in the store.

### D4: FTS5 content table = separate (not external content)

FTS5 tables maintain their own copy of searchable text, not as external-content tables pointing to Ent tables.

**Rationale:** External-content FTS5 requires manual rowid management and `rebuild` commands that couple FTS5 to Ent's internal row IDs. Separate content is simpler, and the storage overhead is negligible for knowledge/learning corpora (typically <10k entries). Sync is maintained by insert/delete/update calls from the knowledge store. Each FTS5 table includes a `source_id UNINDEXED` column for row identification — UNINDEXED so it is not included in text search (avoids false positives from UUID matching).

### D5: One FTS5 table per collection, not a shared table

Separate FTS5 tables for knowledge and learning (e.g., `knowledge_fts`, `learning_fts`) rather than one shared table with a `collection` column.

**Rationale:** Different collections have different searchable columns (knowledge: key+content, learning: trigger+error_pattern+fix). Separate tables allow column-specific FTS5 configuration and avoid cross-collection noise in BM25 ranking.

### D6: LIKE fallback is the existing code path, not a reimplementation

When FTS5 is unavailable, `SearchKnowledge`/`SearchLearnings` use the current Ent LIKE predicates unchanged. The fallback is not a new LIKE implementation in `internal/search/`.

**Rationale:** The Ent-based LIKE path is tested and working. Reimplementing LIKE in `internal/search/` would duplicate logic. The store methods simply check `if fts5Index != nil` and branch.

## Risks / Trade-offs

- **[FTS5 unavailable on some SQLite builds]** → Mitigated by `ProbeFTS5()` runtime check + LIKE fallback. The system degrades gracefully.
- **[FTS5 data out of sync with Ent tables]** → Mitigated by insert/update/delete calls from store methods at write time. No async sync required.
- **[BM25 ranking differs from RelevanceScore ordering]** → Intentional trade-off. FTS5 path uses BM25 for query relevance; LIKE fallback retains RelevanceScore ordering. Results may differ between paths. This is acceptable — BM25 is strictly better for query-relevant ranking.
- **[Storage overhead from FTS5 content copies]** → Negligible for expected corpus sizes. Knowledge entries are typically <10k rows.

## Migration Plan

- **Deploy:** FTS5 tables are created on first app startup when FTS5 is available. Existing entries are bulk-indexed from Ent tables during initialization.
- **Rollback:** Delete FTS5 virtual tables (`DROP TABLE IF EXISTS knowledge_fts; DROP TABLE IF EXISTS learning_fts`). Search automatically falls back to LIKE. No data loss.
- **No config changes required.** FTS5 is auto-detected.
