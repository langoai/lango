## 1. Checkpoint API Extension

- [x] 1.1 Add `metadata map[string]string` parameter to internal `create()` method in `checkpoint.go`
- [x] 1.2 Update `CreateManual` caller to pass `nil` metadata to `create()`
- [x] 1.3 Update `OnJournalEvent` callers to pass `nil` metadata to `create()`
- [x] 1.4 Set `cp.Metadata = metadata` in `create()` before save
- [x] 1.5 Add `CreateManualWithMetadata` public method (runID optional, label required)
- [x] 1.6 Add tests for `CreateManualWithMetadata` (success, empty label, with metadata, no runID)

## 2. Config Fingerprint and Hook Snapshot

- [x] 2.1 Add config fingerprint computation function in `modules_provenance.go`
- [x] 2.2 Add hook registry snapshot function in `modules_provenance.go`
- [x] 2.3 Cache fingerprint and metadata in `provenanceValues` struct

## 3. Session Observer Extension

- [x] 3.1 Extend `buildProvenanceAgentOptions` to accept cached metadata and checkpoint service
- [x] 3.2 Add `sync.Map` idempotency guard for config checkpoint creation
- [x] 3.3 Create `session_config_snapshot` checkpoint in rootSessionObserver callback

## 4. Verification

- [x] 4.1 Run `go test ./internal/provenance/...` and verify all pass
- [x] 4.2 Run `go test ./internal/app/... -run TestProvenance` and verify all pass
- [x] 4.3 Run `go build ./...` and verify no compilation errors
