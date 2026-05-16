## 1. Mission / Proposal Fail-Closed Coverage

- [x] 1.1 Add regression tests for missing mission store dependencies.
- [x] 1.2 Add regression tests for missing proposal registry and preparer dependencies.

## 2. Verification

- [x] 2.1 Run `go test ./internal/mission ./internal/proposal -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for mission and proposal service dependency guards.
- [ ] 3.2 Validate and archive the OpenSpec change.
