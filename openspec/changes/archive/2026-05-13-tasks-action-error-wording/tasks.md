## 1. UX Fix

- [x] 1.1 Render cancel/retry failures as explicit task-action failure messages.
- [x] 1.2 Update regressions for transient error status messaging.

## 2. Docs And Spec Sync

- [x] 2.1 Update the task-surface spec to require explicit task-action failure wording.
- [x] 2.2 Update cockpit docs to describe the same failure message contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
