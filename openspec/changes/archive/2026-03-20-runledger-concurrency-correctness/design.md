## Context

RunLedger is the Task OS durable execution engine in Lango. Its core invariant is event sourcing: the append-only journal is the single source of truth, and `RunSnapshot` is a derived cache. Six correctness bugs violate this contract:

1. **Pointer aliasing in snapshot cache** (`MemoryStore.GetRunSnapshot`, `EntStore.GetRunSnapshot`): Both stores return the cached `*RunSnapshot` pointer directly. `ApplyTail` then mutates the snapshot in-place. A concurrent reader that obtained the same pointer observes partially-applied state mid-mutation. The cache update after `ApplyTail` is also a no-op because `cached` already points to the same object.
2. **Input slice mutation in `VerifyAcceptanceCriteria`**: The method writes `criteria[i].Met = true` on the caller's slice. Since `checkRunCompletion` passes `snap.AcceptanceState` directly, this mutates the snapshot's acceptance state outside the journal, breaking the event-sourcing contract. Additionally, the `ctxKeyNow` context-value code is dead — the value is never set anywhere, so `MetAt` is never populated.
3. **Duplicate criterion journaling in `checkRunCompletion`**: The loop journals `EventCriterionMet` for every criterion where `Met == true`, including criteria that were already met in previous invocations. On each completion check, previously-met criteria generate duplicate journal entries.
4. **Silent error swallowing in `marshalPayload`**: Returns `{}` on marshal failure with no logging. A struct with an unexportable field or cyclic reference silently produces an empty payload, making the journal entry meaningless and the root cause invisible.
5. **Silent error swallowing in projection sync**: The `_ = appendProjectionSyncEvent(...)` pattern in `writethrough.go` discards errors from journal appends. When the journal store is degraded (full disk, lock contention), the system silently drops sync-state records.

## Goals / Non-Goals

**Goals:**
- Eliminate data races in snapshot cache access via deep-copy-on-read
- Preserve the event-sourcing contract: `VerifyAcceptanceCriteria` must not mutate its input
- Eliminate duplicate `EventCriterionMet` journal entries
- Make `marshalPayload` and projection sync failures observable via logging
- Remove dead `ctxKeyNow` code and set `MetAt` directly with `time.Now()`
- Pass `go test -race ./internal/runledger/...` without data race reports

**Non-Goals:**
- Changing the `marshalPayload` function signature (log-only, no error propagation)
- Adding mutex-level locking around `ApplyTail` (deep-copy eliminates the need)
- Refactoring `writethrough.go` error handling beyond adding log calls
- Adding metrics or structured logging infrastructure

## Decisions

### 1. Deep-Copy-on-Read for Snapshot Cache

**Decision**: Add a `DeepCopy() *RunSnapshot` method to `RunSnapshot`. Both `MemoryStore.GetRunSnapshot` and `EntStore.GetRunSnapshot` call `DeepCopy()` on the cached snapshot before returning it. `ApplyTail` operates on the copy. The cache is updated with the mutated copy afterward.

**Alternatives considered**:
- Copy-on-write with versioned snapshots: Higher complexity, more allocation tracking
- Read-write mutex with longer critical section: Would serialize all snapshot reads, creating contention
- Return immutable snapshots (frozen after cache insertion): Requires pervasive API changes to prevent mutation

**Rationale**: Deep-copy-on-read is the simplest fix that preserves the existing API. The allocation cost is acceptable because snapshots are small (steps + criteria, not large payloads) and reads are not on a hot path. The copy isolates callers from cache mutations and from each other.

**Implementation details**:
- `DeepCopy` copies all slice fields (`Steps`, `AcceptanceState`) element-by-element
- `Steps[i].Evidence`, `Steps[i].DependsOn`, `Steps[i].ToolProfile` sub-slices are copied
- `Steps[i].Validator.Params` map is copied key-by-key
- `Notes` map is copied key-by-key
- `AcceptanceState[i].MetAt` pointer is copied (new `*time.Time` with same value)
- Scalar fields (`RunID`, `Status`, `LastJournalSeq`, etc.) are copied by value assignment

### 2. Copy Criteria Slice in VerifyAcceptanceCriteria

**Decision**: `VerifyAcceptanceCriteria` creates a shallow copy of the input `[]AcceptanceCriterion` slice and works on the copy. The caller's slice is never modified. When a criterion passes validation, `Met` is set to `true` and `MetAt` is set to `time.Now()` on the copy. The method returns two values: the list of unmet criteria (as before) and the full evaluated copy (new return).

**Alternatives considered**:
- Return only newly-met indices: Requires callers to correlate indices back to the original slice
- Accept a `*[]AcceptanceCriterion` and document mutation: Violates the event-sourcing contract by design

**Rationale**: Returning the full evaluated copy lets `checkRunCompletion` diff the before/after states to determine which criteria are newly met, without mutating the snapshot. The `ctxKeyNow` mechanism is removed entirely because it was never wired — `time.Now()` is used directly.

**Signature change**:
```
Before: VerifyAcceptanceCriteria(ctx, criteria) ([]AcceptanceCriterion, error)
After:  VerifyAcceptanceCriteria(ctx, criteria) (unmet []AcceptanceCriterion, evaluated []AcceptanceCriterion, error)
```

### 3. Diff-Based Criterion Journaling

**Decision**: `checkRunCompletion` captures the `Met` state of each criterion before calling `VerifyAcceptanceCriteria`. After the call, it compares before vs. after: only criteria that transitioned from `Met=false` to `Met=true` generate an `EventCriterionMet` journal entry.

**Alternatives considered**:
- Track met criteria in a separate set on the snapshot: Adds state outside the journal
- Have `VerifyAcceptanceCriteria` return a `[]int` of newly-met indices: Couples the PEV engine to journaling concerns

**Rationale**: The diff approach is local to `checkRunCompletion`, requires no changes to the snapshot schema, and naturally prevents duplicates. The before-state snapshot is taken from `snap.AcceptanceState[i].Met` before the verify call.

### 4. Log-Only Error Handling for marshalPayload

**Decision**: `marshalPayload` logs the error at warning level using `log.Printf` and continues to return `{}`. The function signature remains `func marshalPayload(v interface{}) json.RawMessage`.

**Alternatives considered**:
- Return `(json.RawMessage, error)`: Would require changes at ~30 call sites across tools.go, writethrough.go, and snapshot.go
- Panic on marshal failure: Too aggressive for a helper used in best-effort contexts

**Rationale**: Marshal failures indicate a programming error (unexported field, nil interface with non-nil type). Logging makes the error visible in operator logs without changing the call-site contract. The `{}` fallback preserves journal append flow — a degraded payload is better than a failed append.

### 5. Warning-Level Logging for Projection Sync Errors

**Decision**: Replace `_ = appendProjectionSyncEvent(...)` with `if err := appendProjectionSyncEvent(...); err != nil { log.Printf("WARN projection sync: %v", err) }`. Keep best-effort semantics — projection sync failures do not fail the outer operation.

**Alternatives considered**:
- Propagate errors to callers: Would change write-through semantics; projection sync is intentionally best-effort
- Structured logging with error counters: Good future enhancement but out of scope for this correctness fix

**Rationale**: The `_ = err` pattern is the root cause of silent degradation. Warning-level logging makes failures visible without changing the control flow. Operators can detect journal-store issues from logs.

## Risks / Trade-offs

| Risk | Severity | Mitigation |
|------|----------|------------|
| Deep-copy allocation overhead | Low | Snapshots are small (tens of steps, single-digit criteria). Benchmark before/after to confirm < 1us per copy. |
| `VerifyAcceptanceCriteria` signature change breaks callers | Medium | Only two call sites: `checkRunCompletion` (tools.go) and `PEVEngine.Verify` (pev.go, does not call it). Grep for all callers before changing. |
| Log noise from projection sync warnings | Low | These should be rare in healthy systems. If noisy, rate-limit in a follow-up. |
| `marshalPayload` log calls in hot path | Low | Journal appends are not hot-path. Marshal failures are exceptional. |
| Removing `ctxKeyNow` breaks test injection | None | Confirmed: `ctxKeyNow` is never set in production or test code. Dead code removal only. |
