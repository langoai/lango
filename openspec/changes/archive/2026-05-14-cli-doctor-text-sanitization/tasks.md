## 1. Renderer Hardening

- [x] 1.1 Sanitize doctor TUI renderer text fields before display.
- [x] 1.2 Sanitize doctor JSON renderer text fields before serialization.
- [x] 1.3 Add regression coverage for malformed doctor result text.

## 2. Spec Sync

- [x] 2.1 Record the doctor text sanitization contract in `cli-doctor`.
- [x] 2.2 Update downstream `docs/cli/core.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/doctor/... -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-doctor-text-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
