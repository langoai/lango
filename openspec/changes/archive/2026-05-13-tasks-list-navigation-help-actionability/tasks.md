## 1. Help Actionability Fix

- [x] 1.1 Add failing Tasks regressions for zero-task and single-task list help.
- [x] 1.2 Hide list-mode navigation help when another task row does not exist.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require conditional list-mode navigation help.
- [x] 2.2 Update cockpit docs to describe the same list-help contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
