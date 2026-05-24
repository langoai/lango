## 1. Regression Coverage

- [x] 1.1 Add an exact missing-parameter regression for `sentinel_acknowledge`.

## 2. Docs And Spec Sync

- [x] 2.1 Update prompt/feature docs for the `sentinel_acknowledge` required-input contract.
- [x] 2.2 Sync the sentinel and production-readiness specs to the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/economy/escrow/sentinel -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
