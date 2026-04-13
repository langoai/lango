## Why

RunLedger's `MemoryStore` and `EntStore` return cached snapshot pointers directly, allowing `ApplyTail` to mutate them in-place. Concurrent readers can observe partially-applied state. Additionally, `VerifyAcceptanceCriteria` mutates its input slice (violating the event-sourcing contract), `checkRunCompletion` journals already-met criteria on every call (creating duplicates), and `marshalPayload`/projection sync silently swallow errors (making degraded state invisible).

## What Changes

- **Deep-copy cached snapshots** (`store.go:193-222`, `ent_store.go:299-328`): `GetRunSnapshot` returns a `DeepCopy()` of the cached snapshot; `ApplyTail` operates on the copy; cache updated atomically. Add `RunSnapshot.DeepCopy()` method.
- **Fix `VerifyAcceptanceCriteria` input mutation** (`pev.go:97-124`): Work on a copy of the criteria slice instead of modifying `criteria[i].Met` in-place. Remove dead `ctxKeyNow` type and related code. Set `MetAt` directly with `time.Now()`.
- **Fix duplicate criterion journaling** (`tools.go:557-566`): Only journal `EventCriterionMet` for criteria that are **newly** met (were false before, true after `VerifyAcceptanceCriteria`).
- **Log `marshalPayload` errors** (`types.go:142-147`): Add error logging instead of silently returning `{}`. Keep signature stable (log, don't propagate) to minimize blast radius.
- **Log projection sync errors** (`writethrough.go:140,144,156,161` etc.): Replace `_ = appendProjectionSyncEvent(...)` with warning-level logging. Keep best-effort semantics.

## Capabilities

### New Capabilities

_(none — all fixes are within existing capability boundaries)_

### Modified Capabilities

- `run-ledger`: Snapshot deep-copy safety, acceptance criteria verification contract, journal deduplication, error observability

## Impact

- **Code**: `internal/runledger/store.go`, `internal/runledger/ent_store.go`, `internal/runledger/pev.go`, `internal/runledger/tools.go`, `internal/runledger/types.go`, `internal/runledger/writethrough.go`, `internal/runledger/snapshot.go` (new `DeepCopy` method)
- **Breaking**: None — all changes are internal correctness fixes
- **Downstream**: No public API changes; docs updated only if behavioral semantics change
- **Risk**: Medium — deep-copy adds allocation cost (acceptable for correctness); `marshalPayload` change touches ~30 call sites if signature changes
- **Verification**: Must pass `go test -race ./internal/runledger/...`
