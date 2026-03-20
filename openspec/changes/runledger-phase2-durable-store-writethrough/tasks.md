## 1. Persistent Store

- [x] 1.1 Implement Ent-backed `RunLedgerStore`
- [x] 1.2 Persist journal events, cached snapshots, and run-step projections
- [x] 1.3 Preserve per-run monotonic sequence assignment under concurrent appends

## 2. Canonical Write Path

- [ ] 2.1 Add write-through adapters for workflow and background execution
- [ ] 2.2 Ensure projections reuse the canonical RunLedger `run_id`
- [ ] 2.3 Add projection sync markers and degraded-projection handling

## 3. Replay and Repair

- [ ] 3.1 Add projection drift detection
- [ ] 3.2 Add projection rebuild/replay API
- [ ] 3.3 Add tests for projection failure then replay recovery

## 4. Downstream

- [x] 4.1 Update `lango run journal` to read persistent journal entries
- [x] 4.2 Update docs and README to reflect Phase 2 durability
- [x] 4.3 Update RunLedger spec with persistence and write-through behavior

## 5. Verification

- [x] 5.1 Run `go build ./...`
- [x] 5.2 Run RunLedger, workflow, and background tests
- [x] 5.3 Run `go test ./...`
