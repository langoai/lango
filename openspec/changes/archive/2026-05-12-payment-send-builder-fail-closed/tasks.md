## 1. Payment Builder Fail-Closed

- [x] 1.1 Add failing tests for unwired transaction builder paths in `tx_builder` and `payment.Service.Send`.
- [x] 1.2 Make unwired builder paths return actionable errors instead of panicking.
- [x] 1.3 Keep payment records and optional field assertions aligned with the new fail-closed behavior.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` spec coverage for payment builder wiring failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
