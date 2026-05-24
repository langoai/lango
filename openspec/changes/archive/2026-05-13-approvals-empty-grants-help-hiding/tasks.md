## 1. UX Fix

- [x] 1.1 Hide `r` and `R` help when the grants section has no rows.
- [x] 1.2 Add regression coverage for empty and populated grants help.

## 2. Docs And Spec Sync

- [x] 2.1 Update the approval-history-view spec to require revoke help only when grant rows exist.
- [x] 2.2 Update cockpit docs to describe the same grants-section help contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
