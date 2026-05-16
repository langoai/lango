## 1. UX Fix

- [x] 1.1 Render `Enter` as `details` in list mode and `close detail` in detail mode.
- [x] 1.2 Add regression coverage for the context-sensitive `Enter` help label.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require the detail-mode `Enter` help label to match the close action.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
