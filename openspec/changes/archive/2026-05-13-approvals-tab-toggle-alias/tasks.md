## 1. UX Fix

- [x] 1.1 Add `Tab` as a section-toggle alias for the Approvals page while preserving `/`.
- [x] 1.2 Update the page help to advertise the actual section-toggle keys.
- [x] 1.3 Add regression coverage for both `Tab` and `/` toggles.

## 2. Docs And Spec Sync

- [x] 2.1 Update the approvals cockpit spec to require `Tab` support and `/` compatibility.
- [x] 2.2 Update public cockpit docs to describe the new toggle key contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
