## 1. Payment Gate Fail-Closed

- [x] 1.1 Add a failing regression for missing receipt-store wiring in `paymentgate.Service`.
- [x] 1.2 Make direct-payment gate evaluation return an actionable unavailable-store error instead of panicking.

## 2. Verification

- [x] 2.1 Run `go test ./internal/paymentgate -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for payment-gate store failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
