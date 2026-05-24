## 1. Live Feature Model Hardening

- [x] 1.1 Sanitize live health feature strings before storing them in `StatusInfo.ServerInfo`.
- [x] 1.2 Add regression coverage for malformed live feature JSON fields.

## 2. Spec Sync

- [x] 2.1 Record the live feature replay-safety contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-live-feature-model-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
