## 1. Tasks Page Hardening

- [x] 1.1 Sanitize rendered task table text.
- [x] 1.2 Sanitize rendered task detail text.
- [x] 1.3 Sanitize transient task action status messages.
- [x] 1.4 Add regression coverage for malformed task metadata.

## 2. Spec Sync

- [x] 2.1 Record the Tasks page text-sanitization contract in `tui-task-surface`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-tasks-text-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
