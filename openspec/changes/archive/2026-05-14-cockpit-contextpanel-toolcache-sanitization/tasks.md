## 1. Tool Cache Hardening

- [x] 1.1 Sanitize tool names when building cached `sortedTools`.
- [x] 1.2 Add regression coverage for malformed cached tool names.

## 2. Spec Sync

- [x] 2.1 Record the cached tool-label replay-safety contract in `cockpit-context-panel`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-contextpanel-toolcache-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
