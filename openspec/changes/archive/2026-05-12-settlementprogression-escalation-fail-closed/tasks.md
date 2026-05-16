## 1. Settlement Progression Escalation Fail-Closed

- [x] 1.1 Add a failing regression for unknown current settlement progression status during escalation.
- [x] 1.2 Replace the panic path with an actionable error return.

## 2. Verification

- [x] 2.1 Run `go test ./internal/settlementprogression -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for settlement progression escalation failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
