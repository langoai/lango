## 1. Snapshot Deep Copy

- [x] 1.1 Add `DeepCopy() *RunSnapshot` method to `snapshot.go` — copy all scalar fields, deep-copy `Steps` (including `Evidence`, `DependsOn`, `ToolProfile`, `Validator.Params` sub-slices/maps), `AcceptanceState` (including `MetAt` pointer), and `Notes` map
- [x] 1.2 Add unit tests for `DeepCopy` — verify independence of all mutable fields (slice append, map mutation, pointer modification on copy do not affect original)

## 2. MemoryStore Snapshot Race Fix

- [x] 2.1 Update `MemoryStore.GetRunSnapshot` (`store.go:193-223`) — call `cached.DeepCopy()` before `ApplyTail`, update cache with the mutated copy
- [x] 2.2 Add race-condition test — concurrent `GetRunSnapshot` + `AppendJournalEvent` under `go test -race`

## 3. EntStore Snapshot Race Fix

- [x] 3.1 Update `EntStore.GetRunSnapshot` (`ent_store.go:299-328`) — call `cached.DeepCopy()` before `ApplyTail`, update cache with the mutated copy
- [x] 3.2 Add race-condition test — concurrent `GetRunSnapshot` + `AppendJournalEvent` under `go test -race` with EntStore

## 4. VerifyAcceptanceCriteria Input Mutation Fix

- [x] 4.1 Change `VerifyAcceptanceCriteria` signature to return `(unmet []AcceptanceCriterion, evaluated []AcceptanceCriterion, error)` — work on a shallow copy of the input slice, set `Met=true` and `MetAt=time.Now()` on the copy only
- [x] 4.2 Remove dead `ctxKeyNow` type and the unused context-value lookup block in `pev.go`
- [x] 4.3 Update all callers of `VerifyAcceptanceCriteria` to handle the new return signature
- [x] 4.4 Add unit test — verify original criteria slice is unmodified after `VerifyAcceptanceCriteria` call

## 5. Duplicate Criterion Journaling Fix

- [x] 5.1 Update `checkRunCompletion` (`tools.go:557-569`) — capture before-state of `snap.AcceptanceState[i].Met`, compare with evaluated copy, only journal `EventCriterionMet` for criteria that transitioned `false -> true`
- [x] 5.2 Add unit test — call `checkRunCompletion` twice with one criterion met each time, verify no duplicate `EventCriterionMet` entries in the journal

## 6. marshalPayload Error Logging

- [x] 6.1 Add `log.Printf("WARN marshalPayload: %v", err)` in the error branch of `marshalPayload` (`types.go:142-147`) — keep returning `{}` as fallback
- [x] 6.2 Add unit test — pass an unmarshalable value (e.g., `chan int`), verify log output contains warning and return value is `{}`

## 7. Projection Sync Error Logging

- [x] 7.1 Replace all `_ = appendProjectionSyncEvent(...)` patterns in `writethrough.go` with `if err := ...; err != nil { log.Printf("WARN projection sync %s: %v", runID, err) }`
- [x] 7.2 Verify no compile errors and run `go build ./internal/runledger/...`

## 8. Verification

- [x] 8.1 Run `go test -race ./internal/runledger/...` — all tests pass with zero race reports
- [x] 8.2 Run `go build ./...` — full project builds without errors
- [x] 8.3 Run `go test ./...` — full test suite passes
