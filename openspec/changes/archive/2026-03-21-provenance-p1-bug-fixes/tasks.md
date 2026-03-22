## 1. Fix EntStore Seq Propagation

- [x] 1.1 Hoist `var nextSeq int64` before retry loop in `EntStore.AppendJournalEvent`
- [x] 1.2 Change `nextSeq := int64(1)` to `nextSeq = int64(1)` (assignment, not declaration)
- [x] 1.3 Add `event.Seq = nextSeq` before hook call on commit success
- [x] 1.4 Add `TestEntStore_AppendHookReceivesSeq` test — hook receives non-zero monotonic Seq

## 2. Post-construction Hook Registration

- [x] 2.1 Add `AppendHookSetter` interface to `internal/runledger/options.go`
- [x] 2.2 Add `SetAppendHook` method on `MemoryStore` with hook chaining
- [x] 2.3 Add `SetAppendHook` method on `EntStore` with hook chaining
- [x] 2.4 Add `TestMemoryStore_SetAppendHook` and `TestMemoryStore_SetAppendHook_Chaining` tests
- [x] 2.5 Add `TestEntStore_SetAppendHook` and `TestEntStore_SetAppendHook_Chaining` tests

## 3. Module Wiring

- [x] 3.1 Wire `cpService.OnJournalEvent` via `SetAppendHook` in `modules_provenance.go`
- [x] 3.2 Add `TestSetAppendHook_Integration` in `checkpoint_test.go`

## 4. EntCheckpointStore

- [x] 4.1 Create `internal/provenance/ent_store.go` implementing `CheckpointStore` with Ent
- [x] 4.2 Implement all 6 methods: SaveCheckpoint, GetCheckpoint, ListByRun, ListBySession, CountBySession, DeleteCheckpoint
- [x] 4.3 Add error mapping: `ent.IsNotFound` → `ErrCheckpointNotFound`, invalid UUID → parse error
- [x] 4.4 Create `internal/provenance/ent_store_test.go` with enttest.Open and full method+error coverage
- [x] 4.5 Add round-trip integration test (save → list → get)

## 5. CLI + App Module Switch

- [x] 5.1 Update `modules_provenance.go` to use `EntCheckpointStore` when DBClient available
- [x] 5.2 Update 3 CLI checkpoint commands to use `NewEntCheckpointStore(boot.DBClient)`
- [x] 5.3 Replace session tree/list commands with "not yet implemented" placeholder
- [x] 5.4 Clean up unused imports in `session.go`

## 6. Verification

- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./internal/runledger/... -count=1` passes
- [x] 6.3 `go test ./internal/provenance/... -count=1` passes
- [x] 6.4 `go test ./internal/app/... -count=1` passes
