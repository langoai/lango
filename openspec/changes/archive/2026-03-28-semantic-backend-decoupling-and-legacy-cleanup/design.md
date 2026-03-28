# Design: Semantic Backend Decoupling + Legacy Path Cleanup

## Step 14: semantic-backend-decoupling

### D1: Build-tag isolation
`//go:build vec` on `sqlite_vec.go`, `//go:build !vec` on `sqlite_vec_stub.go`. The `init()` calling `sqlite_vec.Auto()` only runs when `vec` tag is present.

### D2: Factory function
Both tagged files export `NewVectorStore(db *sql.DB, dimensions int) (VectorStore, error)`:
- `sqlite_vec.go` (vec): delegates to `NewSQLiteVecStore`
- `sqlite_vec_stub.go` (!vec): returns `nil, ErrVecNotCompiled`

### D3: Wiring graceful degradation
`wiring_embedding.go` calls `NewVectorStore` instead of `NewSQLiteVecStore`. On error (vec not compiled), returns nil `embeddingComponents` with descriptive FeatureStatus.

### D4: Auto-enable policy
Auto-enable logic (`ResolveContextAutoEnable`, `ProbeEmbeddingProvider`) is not changed. When embedding provider auto-detects but vec is not compiled, `initEmbedding` returns nil and FeatureStatus warns. Wiring-level graceful degradation only.

## Step 15: legacy-path-cleanup

### D5: Shadow removal
Delete `retrieval/shadow.go`. Remove `shadow` field, `SetShadow()`, `Shadow()` from `RetrievalCoordinator`. Remove `Shadow` from `RetrievalConfig`.

### D6: Coordinator promotion
Coordinator runs in Phase 1 alongside retriever:
- **Retriever** → non-factual layers: RuntimeContext, ToolRegistry, SkillPatterns, PendingInquiries
- **Coordinator** → factual layers: UserKnowledge, AgentLearnings, ExternalKnowledge
Both as parallel goroutines in Phase 1.

### D7: Result merge
`mergeRetrievalResults(a, b *knowledge.RetrievalResult)` combines Items maps from disjoint layer sets. No key conflict expected.

### D8: Event signal change
`ContextInjectedEvent.Items` now contains merged factual+non-factual items. RelevanceAdjuster (Step 13) already filters by `user_knowledge` layer only — behavior unchanged, but input meaning broadened.

### D9: Run summary cache refactor
`summaryCacheEntry` stores `[]RunSummaryContext` instead of formatted string. `retrieveRunSummaryData` returns cached summaries directly. Eliminates `assembleRunSummarySection` fallback.

## Files Changed

### Step 14
| File | Change |
|------|--------|
| `internal/embedding/sqlite_vec.go` | Add `//go:build vec`, add `NewVectorStore` factory |
| `internal/embedding/sqlite_vec_stub.go` | NEW — stub with `ErrVecNotCompiled` |
| `internal/embedding/sqlite_vec_test.go` | Add `//go:build vec` |
| `internal/embedding/rag_test.go` | Add `//go:build vec` |
| `internal/app/wiring_embedding.go` | Use `NewVectorStore`, updated FeatureStatus |
| `Makefile` | `-tags "fts5,vec"` |
| `Dockerfile` | `-tags "fts5,vec"` |
| `README.md` | vec tag documentation |
| `docs/development/build-test.md` | Build tag table |

### Step 15
| File | Change |
|------|--------|
| `internal/retrieval/shadow.go` | DELETED |
| `internal/retrieval/coordinator.go` | Shadow field/methods removed |
| `internal/retrieval/coordinator_test.go` | Shadow test removed |
| `internal/config/types.go` | Shadow removed from RetrievalConfig |
| `internal/config/loader.go` | Shadow default removed |
| `internal/app/wiring_knowledge.go` | Shadow wiring removed |
| `internal/adk/context_model.go` | Coordinator promotion, shadow removal, merge helper |
| `internal/adk/context_assembly.go` | Dead code removal, cache refactor |
| `internal/adk/context_model_test.go` | Test updates |
| `internal/adk/context_model_bench_test.go` | Benchmark updates |
