## 1. Escrow Adjudication Request-ID Guard Unification

- [x] 1.1 Change escrow adjudication to reject empty transaction receipt ids with an actionable validation error.
- [x] 1.2 Update regressions for the unified request-id behavior.

## 2. Verification

- [x] 2.1 Run `go test ./internal/escrowadjudication -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for the adjudication request-id guard.
- [ ] 3.2 Validate and archive the OpenSpec change.
