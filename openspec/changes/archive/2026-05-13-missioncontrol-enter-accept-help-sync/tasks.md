## 1. UX Fix

- [x] 1.1 Render `Enter` as `accept` for focused proposed mission rows.
- [x] 1.2 Add regression coverage for proposal-row vs non-proposal help labeling.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the proposal-row `Enter` help label.
- [x] 2.2 Update cockpit feature docs to describe the same context-sensitive `Enter` behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
