## 1. Footer Toggle Label

- [x] 1.1 Update the Approvals footer hint strip to advertise `tab /` instead of slash alone.
- [x] 1.2 Add a regression covering the footer section-toggle label.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so approvals help guidance covers the shared `tab /` footer label.
- [x] 2.2 Update the `approval-history-view` OpenSpec delta with the footer label contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate approvals-footer-toggle-label-sync --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
