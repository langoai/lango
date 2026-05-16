## 1. UX Fix

- [x] 1.1 Sort loaded session summaries newest-first before rendering.
- [x] 1.2 Add regression coverage for unsorted session inputs.

## 2. Docs And Spec Sync

- [x] 2.1 Confirm the cockpit sessions-page spec reflects enforced newest-first ordering.
- [x] 2.2 Confirm cockpit feature docs reflect the same ordering contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
