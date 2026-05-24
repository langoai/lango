## 1. UX Fix

- [x] 1.1 Render the Approvals section-toggle help key as `tab /`.
- [x] 1.2 Update regressions for the cleaned-up key label.

## 2. Docs And Spec Sync

- [x] 2.1 Update the approval-history-view spec to require the `tab /` label.
- [x] 2.2 Update cockpit docs to use the same visible key wording.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
