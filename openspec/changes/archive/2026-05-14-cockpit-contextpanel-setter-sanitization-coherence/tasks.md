## 1. Context Panel Setter Hardening

- [x] 1.1 Sanitize channel names inside `SetChannelStatuses()`.
- [x] 1.2 Sanitize runtime active-agent labels inside `SetRuntimeStatus()`.
- [x] 1.3 Add regression coverage for malformed setter input values.

## 2. Spec Sync

- [x] 2.1 Record the setter-boundary replay-safety contract in `cockpit-context-panel`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-contextpanel-setter-sanitization-coherence --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
