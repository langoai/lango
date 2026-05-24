## 1. UX Fix

- [x] 1.1 Render `No approval history yet.` for configured-empty history sections.
- [x] 1.2 Add regression coverage for empty history with configured grants.

## 2. Docs And Spec Sync

- [x] 2.1 Update the approval-history-view spec to require the unified history-empty wording.
- [x] 2.2 Update cockpit docs to describe the same history-empty wording.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
