## 1. Validation Error Hardening

- [x] 1.1 Sanitize invalid dead-letter subtype and family flag values in CLI validation errors.
- [x] 1.2 Sanitize invalid dead-letter timestamp flag values in CLI validation errors.
- [x] 1.3 Add regression coverage for malformed invalid-flag text.

## 2. Spec Sync

- [x] 2.1 Record the dead-letter validation-error sanitization contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-invalid-flag-error-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
