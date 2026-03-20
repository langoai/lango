## Why

RunLedger performs expensive operations on every tool call and every LLM request. `ToolProfileGuard` calls `GetRunSnapshot()` 10-20 times per turn. `assembleRunSummarySection` queries all run summaries on every LLM request. `EntStore` uses a single global `sync.Mutex` for all runs. At current scale this works; at target scale (parallel runs, 50+ tool calls/turn) it becomes a bottleneck.

## What Changes

- **Per-turn context-scoped cache for `ToolProfileGuard`** (`tool_profile_guard.go`, seeded from `adk/agent.go`): First tool call in a turn fetches and caches the snapshot in context; subsequent calls reuse it. Expected: 10-20x reduction in snapshot lookups per turn.
- **Session-scoped caching for `assembleRunSummarySection`** (`adk/context_model.go:313`): Cache run summaries per session with short TTL or journal-seq-based invalidation. Expected: 3-5x reduction in summary queries per turn.
- **EntStore lock decomposition** (`ent_store.go:24,37,151,177,270`): Replace global `sync.Mutex` with per-run cache locking (`sync.Map` or per-run mutex). Remove Go-level mutex from `AppendJournalEvent` and rely on DB transaction + retry-on-lock/constraint for serialization.
- **Benchmark suite**: Add before/after benchmarks for each optimization.

## Capabilities

### New Capabilities

_(none — all changes are performance optimizations within existing capability boundaries)_

### Modified Capabilities

- `run-ledger`: Snapshot lookup caching, summary caching, store concurrency model

## Impact

- **Code**: `internal/runledger/tool_profile_guard.go`, `internal/adk/context_model.go`, `internal/runledger/ent_store.go`
- **Breaking**: None — all changes are internal performance improvements
- **Dependencies**: Depends on change C (`runledger-concurrency-correctness`) being applied first — cached snapshots must be deep-copied
- **Downstream**: No public API changes; benchmark results documented
- **Risk**: Low — caching is additive; lock decomposition is internal to EntStore
- **Verification**: Benchmark results showing improvement vs baseline; `go test -race` passes
