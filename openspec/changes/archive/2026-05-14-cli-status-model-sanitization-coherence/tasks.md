## 1. Status Model Hardening

- [x] 1.1 Sanitize collected provider/model/channel/feature fields in `StatusInfo`.
- [x] 1.2 Sanitize live feature-reason enrichment in `collectStatus()`.
- [x] 1.3 Add regression coverage for malformed status model text.

## 2. Spec Sync

- [x] 2.1 Record the status-model replay-safety contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-model-sanitization-coherence --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
