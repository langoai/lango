## 1. UX Fix

- [x] 1.1 Hide Approvals navigation help when the active section has no rows.
- [x] 1.2 Add regression coverage for empty/unavailable section help.

## 2. Docs And Spec Sync

- [x] 2.1 Update the approval-history-view spec to require hiding inert navigation help.
- [x] 2.2 Update cockpit docs to describe that section navigation hints appear only when rows exist.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
