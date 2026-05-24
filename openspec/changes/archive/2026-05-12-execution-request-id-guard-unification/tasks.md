## 1. Execution Request-ID Guard Unification

- [x] 1.1 Change escrow/dispute execution entrypoints to reject empty transaction receipt ids with actionable validation errors.
- [x] 1.2 Update regressions for the unified request-id behavior.

## 2. Verification

- [x] 2.1 Run `go test ./internal/escrowrelease ./internal/escrowrefund ./internal/disputehold ./internal/partialsettlementexecution -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for unified execution request-id guards.
- [ ] 3.2 Validate and archive the OpenSpec change.
