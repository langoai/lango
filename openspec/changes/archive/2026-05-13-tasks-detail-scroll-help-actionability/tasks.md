## 1. Help Actionability Fix

- [x] 1.1 Add a failing Tasks detail-mode regression for non-scrollable detail help.
- [x] 1.2 Hide detail scroll help when the selected task has no scrollable overflow.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require conditional detail scroll help.
- [x] 2.2 Update cockpit docs to describe the same conditional scroll-help contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
