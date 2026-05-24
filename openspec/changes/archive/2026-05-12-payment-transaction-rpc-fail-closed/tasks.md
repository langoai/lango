## 1. Transaction RPC Fail-Closed

- [x] 1.1 Add failing regressions for missing transaction RPC in submit and confirmation helpers.
- [x] 1.2 Make submit and confirmation helpers return actionable errors instead of panicking.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for payment transaction RPC helper failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
