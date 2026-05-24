## 1. UX Fix

- [x] 1.1 Render `↑/k` and `↓/j` as scroll bindings while task detail mode is open.
- [x] 1.2 Add regression coverage for detail-mode help wording.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require scroll-specific help labels in detail mode.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
