## Why

Knowledge and learning search currently uses per-keyword OR predicates with `ContentContains`/`KeyContains` (Ent LIKE queries), which cannot handle phrase matching, prefix search, or relevance ranking beyond static `RelevanceScore`. This limits factual retrieval quality and blocks the context engineering roadmap's downstream steps (budget manager, temporal knowledge, agentic retrieval). FTS5 provides BM25 ranking, phrase/prefix search, and significantly better performance on larger corpora — all without requiring an embedding provider.

## What Changes

- Create `internal/search/` package as a shared FTS5 search substrate (new)
  - `FTS5Index`: table creation, insert/update/delete, search with BM25 ranking
  - `ProbeFTS5()`: runtime detection of FTS5 extension availability
  - LIKE-based fallback when FTS5 is unavailable
- Modify `knowledge.Store.SearchKnowledge()` to use injected `FTS5Index` for primary search
- Modify `knowledge.Store.SearchLearnings()` with the same FTS5-first pattern
- Wire FTS5 index creation and injection during app initialization
- Add FTS5 feature status to context health reporting

## Capabilities

### New Capabilities
- `fts5-search-index`: Shared FTS5 full-text search substrate — table lifecycle, CRUD, BM25 search, runtime probe, LIKE fallback
- `knowledge-fts5-integration`: Integration of FTS5 index into knowledge and learning store search paths

### Modified Capabilities
- `knowledge-store`: Search methods gain FTS5-first path with LIKE fallback; `SetFTS5Index()` injection point added
- `feature-status`: FTS5 availability reported in context health checks

## Impact

- **New package**: `internal/search/` — zero external dependencies beyond `database/sql`
- **Modified packages**: `internal/knowledge/` (store search methods), `internal/app/` (wiring), `internal/cli/doctor/checks/` (health check)
- **Schema**: FTS5 virtual tables created via raw SQL (outside Ent migration)
- **Config**: No new config keys — FTS5 is auto-detected and used when available
- **Breaking**: None — LIKE fallback preserves existing behavior when FTS5 is unavailable
