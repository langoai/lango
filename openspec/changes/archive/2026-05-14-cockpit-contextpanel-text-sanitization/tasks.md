## 1. Context Panel Hardening

- [x] 1.1 Sanitize rendered top-tool names.
- [x] 1.2 Sanitize rendered runtime active-agent and channel names.
- [x] 1.3 Add regression coverage for malformed context-panel labels.

## 2. Spec Sync

- [x] 2.1 Record the context-panel text-sanitization contract in `cockpit-context-panel`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-contextpanel-text-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
