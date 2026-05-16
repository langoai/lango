## 1. Wrapper Guard

- [x] 1.1 Tighten `payment_x402_fetch` to enforce the required `url` parameter at the wrapper layer.
- [x] 1.2 Add regression coverage for the missing-URL wrapper path.

## 2. Verification

- [x] 2.1 Run `go test ./internal/tools/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `payment-tools` and `production-readiness` coverage for the X402 fetch wrapper guard.
- [x] 3.2 Validate and archive the OpenSpec change.
