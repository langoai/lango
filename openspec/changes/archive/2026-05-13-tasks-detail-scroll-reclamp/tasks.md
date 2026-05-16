## 1. UX Fix

- [x] 1.1 Re-clamp task detail scroll when the selected detail content shrinks.
- [x] 1.2 Add regression coverage for refresh-time scroll re-clamping.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require effective viewport-range clamping after refresh.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
