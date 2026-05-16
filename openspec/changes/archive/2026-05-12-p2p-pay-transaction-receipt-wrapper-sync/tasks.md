## 1. Wrapper Guard

- [x] 1.1 Require `transaction_receipt_id` at the `p2p_pay` wrapper layer.
- [x] 1.2 Add regression coverage for the missing `transaction_receipt_id` wrapper error.
- [x] 1.3 Sync docs/specs for the updated required-input contract.

## 2. Verification

- [x] 2.1 Run `go test ./internal/app -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `p2p-payment` and `production-readiness` coverage for the wrapper-level requirement.
- [ ] 3.2 Validate and archive the OpenSpec change.
