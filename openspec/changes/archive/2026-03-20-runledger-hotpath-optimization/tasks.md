## 1. Prerequisites

- [x] 1.1 Verify change C (`runledger-concurrency-correctness`) is merged and deep-copy safety is in place
- [x] 1.2 Add baseline benchmarks for current ToolProfileGuard, assembleRunSummarySection, and EntStore (before optimization) in `_bench_test.go` files

## 2. ToolProfileGuard Per-Turn Context Cache

- [x] 2.1 Define `snapshotCacheKey` context key type and `snapshotCache` struct (pointer with `sync.Once`-guarded load) in `tool_profile_guard.go`
- [x] 2.2 Modify `ToolProfileGuard` middleware to check context for cached snapshot before calling `store.GetRunSnapshot()`
- [x] 2.3 On cache miss, fetch snapshot via store, store in context cache, and proceed with profile check
- [x] 2.4 Add unit tests: cache hit (2nd+ call skips store), cache miss (1st call fetches), cross-turn isolation (fresh context has no cache)
- [x] 2.5 Add `BenchmarkToolProfileGuard_WithCache` and `BenchmarkToolProfileGuard_NoCache` benchmarks

## 3. Session-Scoped Run Summary Cache

- [x] 3.1 Add `MaxJournalSeqForSession(ctx, sessionKey) (int64, error)` method to `RunLedgerStore` interface and implement in `EntStore` and `MemoryStore`
- [x] 3.2 Define `runSummaryCache` struct with `sync.RWMutex` and `map[string]summaryCacheEntry` on `ContextAwareModelAdapter`
- [x] 3.3 Modify `assembleRunSummarySection` to check cache: compare stored max seq against current max seq from `MaxJournalSeqForSession`
- [x] 3.4 On cache hit (seq unchanged), return cached summary string; on miss/invalidation, query fresh summaries, build string, update cache entry
- [x] 3.5 Add unit tests: cache hit returns cached string without store query, cache invalidation on seq change triggers fresh query, cache miss populates entry
- [x] 3.6 Add `BenchmarkAssembleRunSummary_CacheHit` and `BenchmarkAssembleRunSummary_CacheMiss` benchmarks

## 4. EntStore Lock Decomposition

- [x] 4.1 Replace `mu sync.Mutex` + `cache map[string]*RunSnapshot` with `locks sync.Map` + `cache sync.Map` on `EntStore`
- [x] 4.2 Add `func (s *EntStore) runLock(runID string) *sync.Mutex` helper using `sync.Map.LoadOrStore`
- [x] 4.3 Update `GetCachedSnapshot` to use per-run lock instead of global mutex
- [x] 4.4 Update `UpdateCachedSnapshot` to use per-run lock instead of global mutex
- [x] 4.5 Remove Go-level mutex from `AppendJournalEvent` (DB transaction provides serialization)
- [x] 4.6 Add `TestAppendJournalEvent_ConcurrentSameRun` test verifying correct seq assignment under `-race`
- [x] 4.7 Add `TestGetCachedSnapshot_ConcurrentDifferentRuns` test verifying no cross-run blocking
- [x] 4.8 Add `BenchmarkEntStore_ParallelRuns` and `BenchmarkEntStore_GlobalLock_Baseline` benchmarks

## 5. Integration Verification

- [x] 5.1 Run `go build ./...` to verify no compilation errors
- [x] 5.2 Run `go test ./internal/runledger/... ./internal/adk/... -race -count=1` to verify all tests pass with race detector
- [x] 5.3 Run full benchmark suite and record before/after results in benchmark output
- [x] 5.4 Verify no public API changes (store interface additions are additive, no breaking changes)
