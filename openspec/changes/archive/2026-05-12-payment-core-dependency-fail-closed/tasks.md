## 1. Payment Core Dependency Guards

- [x] 1.1 Add failing regressions for missing `wallet`, `limiter`, and `store` dependencies in payment service paths.
- [x] 1.2 Make user-facing payment service methods fail closed with actionable dependency errors.

## 2. Verification

- [x] 2.1 Run `go test ./internal/payment -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for payment core dependency wiring failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
