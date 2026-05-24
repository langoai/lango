## 1. Status CLI Hardening

- [x] 1.1 Sanitize rendered provider/model, feature, and channel labels in the status dashboard.
- [x] 1.2 Sanitize rendered dead-letter backlog/detail/summary text.
- [x] 1.3 Add regression coverage for malformed status CLI labels.

## 2. Spec Sync

- [x] 2.1 Record the plain-text status CLI contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-text-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
