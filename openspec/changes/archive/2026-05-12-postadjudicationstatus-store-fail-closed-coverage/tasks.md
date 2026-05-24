## 1. Post-Adjudication Status Fail-Closed Coverage

- [x] 1.1 Add regression tests for missing receipt-store wiring in post-adjudication status entrypoints.

## 2. Verification

- [x] 2.1 Run `go test ./internal/postadjudicationstatus -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for post-adjudication status store failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
