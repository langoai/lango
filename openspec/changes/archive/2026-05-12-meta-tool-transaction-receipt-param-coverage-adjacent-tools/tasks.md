## 1. Wrapper Param Guards

- [x] 1.1 Add dispute-hold meta tool regression coverage for missing `transaction_receipt_id`.
- [x] 1.2 Add escrow release/refund meta tool regression coverage for missing `transaction_receipt_id`.
- [x] 1.3 Add post-adjudication status/replay meta tool regression coverage for missing `transaction_receipt_id`.

## 2. Verification

- [x] 2.1 Run `go test ./internal/app -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `meta-tools` and `production-readiness` coverage for the broader wrapper-level missing-parameter cluster.
- [ ] 3.2 Validate and archive the OpenSpec change.
