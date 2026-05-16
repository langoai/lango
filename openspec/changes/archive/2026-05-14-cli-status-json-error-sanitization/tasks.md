## 1. JSON Error Hardening

- [x] 1.1 Sanitize top-level `statusJSONError.error` text before JSON serialization.
- [x] 1.2 Add regression coverage for malformed JSON error payload text.

## 2. Spec Sync

- [x] 2.1 Record the JSON error payload sanitization contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-json-error-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
