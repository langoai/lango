## 1. Cron Loop Source Hardening

- [x] 1.1 Sanitize cron job name and last-run status before loop projection.
- [x] 1.2 Add regression coverage for malformed scheduled-loop source text.

## 2. Spec Sync

- [x] 2.1 Record the scheduled-loop source replay-safety contract in `mission-control-tui`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-cron-loop-source-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
