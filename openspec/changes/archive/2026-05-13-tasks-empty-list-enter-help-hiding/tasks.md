## 1. Help Actionability Fix

- [x] 1.1 Add a failing Tasks regression for empty-list `Enter` help.
- [x] 1.2 Hide list-mode `Enter` help when no task row exists.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require conditional empty-list `Enter` help.
- [x] 2.2 Update cockpit docs to describe the same empty-state help contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
