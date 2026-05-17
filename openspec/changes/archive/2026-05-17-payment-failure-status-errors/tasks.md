## 1. Tests First

- [x] 1.1 Add a failing focused test for failed-status update persistence errors.

## 2. Implementation

- [x] 2.1 Make `failTx` return failed-status update errors without losing the original payment failure.
- [x] 2.2 Update all `Send` failure paths that call `failTx` to include the returned failure information.
- [x] 2.3 Preserve existing successful payment and normal failed-status behavior.

## 3. Verification

- [x] 3.1 Run focused payment service tests.
- [x] 3.2 Run `openspec validate payment-failure-status-errors --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
