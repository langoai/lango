## 1. Error Boundary Hardening

- [x] 1.1 Sanitize non-JSON status command errors at the shared command boundary.
- [x] 1.2 Preserve `errors.Is` / `Unwrap` behavior for sanitized CLI errors.
- [x] 1.3 Add regression coverage for malformed downstream status errors.

## 2. Spec Sync

- [x] 2.1 Record the shared non-JSON error sanitization contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-nonjson-error-boundary-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
