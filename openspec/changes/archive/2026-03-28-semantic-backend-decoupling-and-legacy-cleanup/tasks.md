# Tasks: Semantic Backend Decoupling + Legacy Path Cleanup

## Step 14: semantic-backend-decoupling

- [x] 14.1 Add `//go:build vec` to `internal/embedding/sqlite_vec.go`
- [x] 14.2 Add `NewVectorStore` factory to `sqlite_vec.go` (delegates to `NewSQLiteVecStore`)
- [x] 14.3 Create `internal/embedding/sqlite_vec_stub.go` with `//go:build !vec`, `ErrVecNotCompiled`, stub `NewVectorStore`
- [x] 14.4 Add `//go:build vec` to `sqlite_vec_test.go` and `rag_test.go`
- [x] 14.5 Update `wiring_embedding.go`: `NewSQLiteVecStore` → `NewVectorStore`, update FeatureStatus suggestion
- [x] 14.6 Update `Makefile` build/test targets with `-tags "fts5,vec"`
- [x] 14.7 Update `Dockerfile` build command with `-tags "fts5,vec"`
- [x] 14.8 Update `README.md` Embedding & RAG section with vec tag requirement
- [x] 14.9 Update `docs/development/build-test.md` with build tag table
- [x] 14.10 Verify build: `go build -tags fts5 ./...` (FTS5-only) succeeds
- [x] 14.11 Verify build: `go build -tags "fts5,vec" ./...` (full) succeeds
- [x] 14.12 Verify tests: `go test -tags fts5 ./internal/embedding/...` (vec tests excluded)
- [x] 14.13 Verify tests: `go test -tags "fts5,vec" ./internal/embedding/...` (all pass)

## Step 15: legacy-path-cleanup

- [x] 15.1 Delete `internal/retrieval/shadow.go`
- [x] 15.2 Remove `shadow` field, `SetShadow()`, `Shadow()` from `RetrievalCoordinator`
- [x] 15.3 Remove `TestRetrievalCoordinator_Shadow` from `coordinator_test.go`
- [x] 15.4 Remove `Shadow` from `RetrievalConfig` in `config/types.go`
- [x] 15.5 Remove `Shadow: true` default from `config/loader.go`
- [x] 15.6 Remove `coordinator.SetShadow()` and shadow log field from `wiring_knowledge.go`
- [x] 15.7 Promote coordinator to Phase 1: add coordinator goroutine, reduce retriever to non-factual layers
- [x] 15.8 Add `mergeRetrievalResults` helper in `context_model.go`
- [x] 15.9 Remove shadow goroutine (lines 295-309) from `GenerateContent`
- [x] 15.10 Remove `oldRetrieved` variable from `GenerateContent`
- [x] 15.11 Update `buildContextInjectedItems` to use merged result
- [x] 15.12 Delete `assembleMemorySection` (zero callers) from `context_assembly.go`
- [x] 15.13 Refactor `summaryCacheEntry` to store `[]RunSummaryContext`
- [x] 15.14 Update `retrieveRunSummaryData` to return cached summaries directly
- [x] 15.15 Delete `assembleRunSummarySection` from `context_assembly.go`
- [x] 15.16 Remove run summary fallback call from `context_model.go`
- [x] 15.17 Update `context_model_test.go` tests
- [x] 15.18 Update `context_model_bench_test.go` benchmarks
- [x] 15.19 Verify build and all tests pass
