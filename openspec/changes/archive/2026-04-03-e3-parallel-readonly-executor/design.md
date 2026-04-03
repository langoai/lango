## Design: ParallelReadOnlyExecutor

### Architecture

The executor lives in `internal/streamx/` alongside other stream combinators. It depends on `internal/agent` for `Tool` and `ToolCapability` types.

```
internal/agent (Tool, ToolCapability)
        ↑
internal/streamx/parallel_executor.go
```

### Concurrency Model

- Uses `golang.org/x/sync/errgroup` with `SetLimit(maxConcurrency)` for bounded parallelism
- Each eligible tool runs in its own goroutine within the errgroup
- Non-eligible tools are rejected synchronously before goroutine dispatch
- Results stored in a pre-allocated indexed slice to preserve invocation order without channels

### Eligibility Check

A tool is eligible for parallel execution when both conditions hold:
- `Capability.ReadOnly == true` — tool performs no mutations
- `Capability.ConcurrencySafe == true` — tool is safe for concurrent invocation

This is a strict AND; missing either flag defaults to ineligible (fail-safe).

### Error Handling

- Non-eligible tools: error recorded in `ToolResult.Error` with capability values for debugging
- Handler errors: captured per-result, not propagated to the errgroup (no cascading cancellation)
- Context cancellation: checked before each goroutine starts; propagated via `errgroup.WithContext`
- Nil tools: handled as error result at the corresponding index

### Key Decisions

| Decision | Rationale |
|---|---|
| Indexed slice over channels | Preserves invocation order, simpler API contract |
| errgroup over raw goroutines | Built-in concurrency limiting, context propagation, clean lifecycle |
| Per-result errors over group error | Callers need per-tool granularity; one failure shouldn't abort others |
| Pre-check eligibility | Fail fast without goroutine overhead for ineligible tools |
