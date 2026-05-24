## 1. Post-Adjudication Status Request Guard Coverage

- [x] 1.1 Add a regression for missing transaction receipt id in status lookup.
- [x] 1.2 Make status lookup return an actionable validation error for an empty transaction receipt id.

## 2. Verification

- [x] 2.1 Run `go test ./internal/postadjudicationstatus -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for the status request-id guard.
- [ ] 3.2 Validate and archive the OpenSpec change.
