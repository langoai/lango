## 1. Regression Coverage

- [x] 1.1 Add exact missing-parameter regressions for `contract_read`, `contract_call`, and `contract_abi_load`.

## 2. Docs And Spec Sync

- [x] 2.1 Update prompt/public docs for the contract required-input wrapper contract.
- [x] 2.2 Sync the contract interaction and production-readiness specs to the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/app -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
