## 1. UX Fix

- [x] 1.1 Render configured session-list failures with explicit session-list failure wording.
- [x] 1.2 Add regression coverage for the load-error view.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit sessions-page spec to describe the configured-source failure state.
- [x] 2.2 Update cockpit feature docs to describe the same failure contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
