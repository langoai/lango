## 1. Config + Core Types

- [x] 1.1 Create `internal/config/types_provenance.go` — ProvenanceConfig and CheckpointConfig structs
- [x] 1.2 Add `Provenance ProvenanceConfig` field to Config struct in `types.go`
- [x] 1.3 Add provenance defaults in `DefaultConfig()` in `loader.go`
- [x] 1.4 Create `internal/provenance/types.go` — Checkpoint, TokenSummary, SessionNode, Attribution domain types
- [x] 1.5 Create `internal/provenance/errors.go` — sentinel errors (ErrCheckpointNotFound, etc.)
- [x] 1.6 Create `internal/provenance/store.go` — CheckpointStore interface

## 2. RunLedger Append Hook

- [x] 2.1 Create `internal/runledger/options.go` — StoreOptions, StoreOption, WithAppendHook, applyStoreOptions
- [x] 2.2 Modify `MemoryStore` to accept `...StoreOption` and call hook after append (outside lock)
- [x] 2.3 Modify `EntStore` to accept `...StoreOption` and call hook after successful commit

## 3. Checkpoint Store + Service

- [x] 3.1 Create `internal/provenance/memory_store.go` — in-memory CheckpointStore implementation
- [x] 3.2 Create `internal/provenance/memory_store_test.go` — store tests (CRUD, list, count)
- [x] 3.3 Create `internal/provenance/checkpoint.go` — CheckpointService (manual + auto creation via hook)
- [x] 3.4 Create `internal/provenance/checkpoint_test.go` — service tests + append hook integration test

## 4. Ent Schemas

- [x] 4.1 Create `internal/ent/schema/provenance_checkpoint.go` — ProvenanceCheckpoint schema
- [x] 4.2 Create `internal/ent/schema/session_provenance.go` — SessionProvenance schema
- [x] 4.3 Run `go generate ./internal/ent/...` to generate entity code

## 5. Session Tree

- [x] 5.1 Create `internal/provenance/session_tree.go` — SessionTreeStore interface, MemoryTreeStore, SessionTree service
- [x] 5.2 Create `internal/provenance/session_tree_test.go` — tree registration, hierarchy, close tests

## 6. Session Lifecycle Hook

- [x] 6.1 Add SessionLifecycleEvent type and WithLifecycleHook option to `internal/session/child_store.go`
- [x] 6.2 Add hook calls in ForkChild, MergeChild, DiscardChild (outside locks)
- [x] 6.3 Verify existing session tests pass with backward-compatible variadic constructor

## 7. CLI Commands

- [x] 7.1 Create `internal/cli/provenance/provenance.go` — root command + status subcommand
- [x] 7.2 Create `internal/cli/provenance/checkpoint.go` — list, create, show subcommands
- [x] 7.3 Create `internal/cli/provenance/session.go` — tree, list subcommands
- [x] 7.4 Create `internal/cli/provenance/attribution.go` — show, report stubs (Phase 3)
- [x] 7.5 Register provenance command in `cmd/lango/main.go` under "auto" group

## 8. App Wiring

- [x] 8.1 Add `ProvidesProvenance` to `internal/appinit/module.go`
- [x] 8.2 Create `internal/app/modules_provenance.go` — provenance module (depends on RunLedger)
- [x] 8.3 Register provenanceModule in `internal/app/app.go`

## 9. Downstream Updates

- [x] 9.1 Add "Provenance" feature line to `internal/cli/status/status.go`
- [x] 9.2 Fix status test to include Provenance in all-enabled config
- [x] 9.3 Add Provenance to startup summary in `cmd/lango/main.go`

## 10. Verification

- [x] 10.1 `go build ./...` succeeds
- [x] 10.2 `go test ./internal/provenance/...` — all 21 tests pass
- [x] 10.3 `go test ./internal/runledger/...` — regression pass
- [x] 10.4 `go test ./internal/session/...` — regression pass
- [x] 10.5 `go test ./...` — full regression pass (zero failures)
- [x] 10.6 CLI smoke test: `lango provenance --help` shows all subcommands
