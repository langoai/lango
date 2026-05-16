## 1. Label Sync

- [x] 1.1 Update the Approvals help-bar label for `R` to `revoke session`.
- [x] 1.2 Add a regression covering the unified label.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so the `R` action uses the same session-scoped wording.
- [x] 2.2 Update the `approval-history-view` OpenSpec delta with the unified label.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate approvals-revoke-session-label-sync --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
