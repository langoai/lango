## 1. Knowledge Runtime Fail-Closed

- [x] 1.1 Add failing regressions for missing receipt-store wiring in `knowledgeruntime.Service`.
- [x] 1.2 Make knowledge-runtime entrypoints return actionable unavailable-store errors instead of panicking.

## 2. Verification

- [x] 2.1 Run `go test ./internal/knowledgeruntime -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for knowledgeruntime store failures.
- [ ] 3.2 Validate and archive the OpenSpec change.
