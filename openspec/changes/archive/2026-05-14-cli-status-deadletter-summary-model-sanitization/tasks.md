## 1. Summary Model Hardening

- [x] 1.1 Sanitize dead-letter summary bucket and top-item labels before they enter the aggregated model.
- [x] 1.2 Add regression coverage for malformed dead-letter summary JSON labels.

## 2. Spec Sync

- [x] 2.1 Record the dead-letter summary replay-safety contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-deadletter-summary-model-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
