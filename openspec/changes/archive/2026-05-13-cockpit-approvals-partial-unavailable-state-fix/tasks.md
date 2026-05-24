## 1. UX Fix

- [x] 1.1 Distinguish missing history-store state from empty history.
- [x] 1.2 Distinguish missing grant-store state from empty grants.
- [x] 1.3 Add regression coverage for the partial unavailable states.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the section-level unavailable-state behavior.
- [x] 2.2 Update the approvals page spec to distinguish partial unavailable and empty states.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
