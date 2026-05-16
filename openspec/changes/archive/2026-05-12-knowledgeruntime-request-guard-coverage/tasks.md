## 1. Knowledge Runtime Request Guard Coverage

- [x] 1.1 Add a regression for missing transaction receipt id in `SelectExecutionPath`.
- [x] 1.2 Make `SelectExecutionPath` return an actionable validation error for empty transaction receipt ids.

## 2. Verification

- [x] 2.1 Run `go test ./internal/knowledgeruntime -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `production-readiness` coverage for the knowledge-runtime request guard.
- [ ] 3.2 Validate and archive the OpenSpec change.
