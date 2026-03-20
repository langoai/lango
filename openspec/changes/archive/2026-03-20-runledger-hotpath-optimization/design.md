## Context

RunLedger's hot paths execute expensive operations on every tool call and every LLM request. Three bottlenecks have been identified through profiling:

1. **ToolProfileGuard snapshot lookup** (`tool_profile_guard.go:27`): Each tool invocation calls `store.GetRunSnapshot()`, which hits the DB (cached snapshot query + tail replay). With 10-20 tool calls per turn, this produces 10-20 identical lookups because the snapshot does not change within a single turn.

2. **assembleRunSummarySection full scan** (`adk/context_model.go:313`): Every LLM request calls `ListRunSummaries()`, which queries the DB, unmarshals JSON, and builds the summary string. Within a session, run summaries change only when journal events are appended, yet the query runs unconditionally on every request.

3. **EntStore global mutex** (`ent_store.go:24`): A single `sync.Mutex` guards all cache map operations across all runs. `AppendJournalEvent` holds this lock for the entire DB transaction duration, blocking unrelated runs' cache reads.

This change depends on change C (`runledger-concurrency-correctness`) being applied first, which ensures deep-copy safety for cached snapshots.

## Goals / Non-Goals

**Goals:**

- Reduce ToolProfileGuard snapshot lookups from N-per-turn to 1-per-turn via context-scoped caching.
- Reduce assembleRunSummarySection DB queries from 1-per-LLM-request to 1-per-session-invalidation-cycle via session-scoped caching with journal-seq-based invalidation.
- Eliminate cross-run lock contention in EntStore by decomposing the global mutex into per-run locks.
- Remove the Go-level mutex from `AppendJournalEvent` where the DB transaction already provides serialization.
- Provide a benchmark suite that demonstrates measurable improvement for each optimization.

**Non-Goals:**

- Changing the RunLedger public API or store interface.
- Adding new configuration knobs (all caches are internal, not user-configurable).
- Optimizing the Ent query layer itself (e.g., prepared statements, connection pooling).
- Changing journal event semantics or snapshot materialization logic.
- Read-path caching for CLI commands (`lango run list/journal`) — these are cold paths.

## Decisions

### D1: Per-turn context-scoped cache for ToolProfileGuard

**Choice**: Store the fetched `RunSnapshot` in a context value keyed by `(runID, turnID)`. First call fetches and stores; subsequent calls in the same turn reuse the cached value.

**Alternatives considered**:
- **Middleware-level field cache**: Store snapshot on the middleware struct. Rejected because the middleware is shared across turns and would require explicit invalidation and locking.
- **sync.Map keyed by runID**: Would cache across turns, risking stale data. Turn-scoped context values are naturally garbage-collected when the context is done.

**Rationale**: Context propagation is already used throughout the tool call chain (session key, run ID). Adding a snapshot cache to context is idiomatic, zero-allocation on cache hit (pointer in context), and automatically scoped to the turn's lifetime. No invalidation logic needed — each turn starts with a fresh context.

**Implementation approach**:
- Define a private context key type `snapshotCacheKey` in `tool_profile_guard.go`.
- Seed a shared `*snapshotCache` into the turn context once in `adk.Agent.Run`.
- On first lookup: fetch snapshot through the cache loader and store it in the shared cache.
- On subsequent lookups in the same turn: retrieve from the shared cache and skip store calls.
- Because tool handlers receive `context.Context` by value, the context must already carry a mutable shared pointer before the first tool call.

### D2: Session-scoped cache for assembleRunSummarySection

**Choice**: Cache the assembled summary string per session with journal-seq-based invalidation. The cache entry stores `(summary string, lastSeq int64)`. Before using the cache, compare the stored seq against the current max journal seq for the session's runs. If unchanged, return cached string.

**Alternatives considered**:
- **TTL-based cache** (e.g., 5s expiry): Simple but introduces staleness window. After a journal event, the summary could be stale for up to 5 seconds.
- **Event-bus subscription**: Subscribe to journal events and invalidate on append. More complex wiring and introduces coupling to the event bus.
- **No cache, just reduce query scope**: Only query active/paused runs. Reduces work but still hits DB every request.

**Rationale**: Journal-seq-based invalidation provides exact consistency — the cache is invalidated precisely when journal events change. The seq check is a single lightweight query (`SELECT MAX(last_journal_seq) FROM run_snapshots WHERE session_key = ?`) compared to the full unmarshal-and-format pipeline. This approach requires no new infrastructure (no event bus, no TTL goroutine).

**Implementation approach**:
- Add a `runSummaryCache` struct to `ContextAwareModelAdapter` with fields: `mu sync.RWMutex`, `entries map[string]summaryCacheEntry` where entry is `{summary string, maxSeq int64}`.
- In `assembleRunSummarySection`: acquire read lock, check if entry exists and seq matches current max seq. If hit, return cached summary. If miss, release read lock, query full summaries, build string, acquire write lock, store entry.
- Add a `MaxJournalSeqForSession(ctx, sessionKey) (int64, error)` method to the store interface (or use existing query capabilities).

### D3: EntStore per-run lock decomposition

**Choice**: Replace the single `sync.Mutex` with a `sync.Map` of per-run `sync.Mutex` pointers. Each run ID gets its own lock. `AppendJournalEvent` drops the Go-level mutex entirely (DB transaction provides serialization); only cache mutations retain per-run locking.

**Alternatives considered**:
- **sync.RWMutex** on the global lock: Allows concurrent reads but still serializes all writes across runs. Does not solve cross-run contention.
- **Sharded locks** (hash runID to N buckets): Reduces contention but still allows false sharing between runs in the same bucket. More complex than per-run locks with no clear advantage given expected run counts (tens, not millions).
- **Lock-free cache with sync.Map only**: `sync.Map` for the cache map itself, no per-entry locking. Works for simple load/store but `GetRunSnapshot`'s read-modify-write (cache miss -> materialize -> store) needs atomicity per run.

**Rationale**: Per-run locks eliminate all cross-run contention. The number of concurrent runs is small (typically < 100), so the overhead of per-run mutexes is negligible. `sync.Map` provides safe concurrent access to the lock registry itself. Removing the mutex from `AppendJournalEvent` is safe as long as the DB transaction path retries transient lock/constraint failures and the `(run_id, seq)` unique constraint prevents duplicate sequence numbers.

**Implementation approach**:
- Replace `mu sync.Mutex` + `cache map[string]*RunSnapshot` with `locks sync.Map` (values: `*sync.Mutex`) + `cache sync.Map` (values: `*RunSnapshot`).
- Add helper `func (s *EntStore) runLock(runID string) *sync.Mutex` that does `LoadOrStore` on the locks map.
- `GetCachedSnapshot`: lock per-run mutex, read/write cache, unlock.
- `UpdateCachedSnapshot`: lock per-run mutex, write cache, unlock.
- `AppendJournalEvent`: no Go-level lock. DB transaction handles serialization, with retry on transient lock / unique-constraint conflicts.

### D4: Benchmark suite

**Choice**: Add `_bench_test.go` files colocated with each optimized file. Use `testing.B` with realistic data (10-20 tool calls per turn, 3-5 active runs per session, concurrent run operations).

**Benchmarks**:
- `BenchmarkToolProfileGuard_WithCache` vs `BenchmarkToolProfileGuard_NoCache`: Measure N tool calls in a single turn.
- `BenchmarkAssembleRunSummary_CacheHit` vs `BenchmarkAssembleRunSummary_CacheMiss`: Measure repeated LLM requests in a session.
- `BenchmarkEntStore_ParallelRuns`: Measure concurrent `GetRunSnapshot` + `AppendJournalEvent` across M runs with N goroutines.
- `BenchmarkEntStore_GlobalLock_Baseline`: Baseline with current global mutex for comparison.

## Risks / Trade-offs

**[Risk: Stale snapshot in ToolProfileGuard context cache]** If a journal event modifies the active step's tool profile mid-turn (e.g., policy decision changes the step), the cached snapshot would be stale for the remainder of that turn.
- **Mitigation**: Tool profile changes within a single turn are not a supported workflow. Policy decisions occur between turns. The context cache lifetime (single turn) is short enough that staleness is acceptable. If needed, a `InvalidateSnapshotCache(ctx)` escape hatch can be added later.

**[Risk: Summary cache memory growth]** The session-scoped summary cache grows with the number of active sessions.
- **Mitigation**: Entries are small (short summary strings + int64 seq). Eviction is implicit: sessions that stop making LLM requests naturally stop hitting the cache. A simple size cap (e.g., 1000 entries with LRU eviction) can be added if needed, but is not expected to be necessary at current scale.

**[Risk: Per-run lock map unbounded growth]** The `sync.Map` of per-run locks grows with every run that has ever been accessed.
- **Mitigation**: Each entry is a single `*sync.Mutex` (tiny). Completed/failed runs that are no longer accessed can be cleaned up during periodic maintenance (e.g., when `ListRuns` prunes old entries). This is a known minor leak, acceptable for current scale.

**[Risk: AppendJournalEvent without Go-level lock]** Removing the mutex from `AppendJournalEvent` relies entirely on DB transaction isolation for correctness.
- **Mitigation**: The journal's `(run_id, seq)` unique constraint prevents duplicate sequence numbers. The seq calculation (`SELECT MAX(seq) ... FOR UPDATE` within the transaction) is serialized by the DB. This is verified by existing tests and by adding a new `TestAppendJournalEvent_ConcurrentSameRun` test under `-race`.

**[Risk: Dependency on change C]** This change assumes deep-copy safety from `runledger-concurrency-correctness`. If applied without change C, cached snapshots shared across goroutines could be mutated unsafely.
- **Mitigation**: Hard dependency — tasks.md gates implementation on change C being merged. CI should verify change C is present before this branch is merged.
