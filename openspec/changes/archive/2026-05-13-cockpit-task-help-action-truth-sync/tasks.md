## 1. UX Fix

- [x] 1.1 Make Tasks page help expose only the currently valid action key.
- [x] 1.2 Add regression coverage for running, failed, and no-action task states.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to say the help bar is selected-task aware.
- [x] 2.2 Update the task-surface spec to require state-accurate action hints.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
