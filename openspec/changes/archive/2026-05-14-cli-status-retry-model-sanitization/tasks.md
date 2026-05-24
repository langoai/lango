## 1. Retry Result Hardening

- [x] 1.1 Sanitize dead-letter retry result message and follow-up error text.
- [x] 1.2 Sanitize dead-letter retry follow-up subtype/family/reason/dispatch/background-task status fields.
- [x] 1.3 Add regression coverage for malformed retry-result JSON fields.

## 2. Spec Sync

- [x] 2.1 Record the retry-result replay-safety contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-retry-model-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
