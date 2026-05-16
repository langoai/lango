## 1. UX Fix

- [x] 1.1 Hide `↑/↓` help in Approvals when the active section has fewer than two rows.
- [x] 1.2 Add regression coverage for single-row history and grants states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the approval-history-view spec to require navigation help only when another row exists.
- [x] 2.2 Update cockpit docs to describe the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
