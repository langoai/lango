## 1. UX Fix

- [x] 1.1 Distinguish nil-lister unavailable state from empty configured state in TasksPage.
- [x] 1.2 Add regression coverage for the unavailable-state wording.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the unavailable-state behavior.
- [x] 2.2 Update task-surface spec to distinguish unavailable and empty states.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
