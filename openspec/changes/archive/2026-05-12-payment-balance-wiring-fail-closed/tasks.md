## 1. Balance Wiring Fail-Closed

- [x] 1.1 Add failing regressions for unwired `payment.Service.Balance` paths.
- [x] 1.2 Make balance wiring failures return actionable errors instead of panicking.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for payment balance wiring failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
