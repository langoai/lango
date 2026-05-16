## 1. UX Fix

- [x] 1.1 Clamp task detail scrolling to the last meaningful content offset.
- [x] 1.2 Add regression coverage for maximum detail scroll.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require bounded detail scrolling.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
