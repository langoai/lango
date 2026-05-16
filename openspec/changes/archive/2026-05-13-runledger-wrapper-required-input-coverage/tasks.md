## 1. Guard Coverage

- [x] 1.1 Require explicit wrapper inputs across the `run_*` tool handlers.
- [x] 1.2 Add tool-entrypoint regression coverage for missing required run-ledger inputs.

## 2. Docs And Spec Sync

- [x] 2.1 Update run-ledger feature docs for the required wrapper contract.
- [x] 2.2 Update run-ledger and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/runledger -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
