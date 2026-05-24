## 1. Tasks Detail Reset

- [x] 1.1 Close Tasks detail mode when refresh leaves the task list empty.
- [x] 1.2 Add a regression covering refresh-to-empty from detail mode.

## 2. Downstream Sync

- [x] 2.1 Update the public cockpit docs for the empty-after-refresh detail reset.
- [x] 2.2 Update the `tui-task-surface` spec delta with the reset contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate tasks-empty-refresh-closes-detail --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
