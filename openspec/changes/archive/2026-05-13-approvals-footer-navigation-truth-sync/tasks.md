## 1. Footer Help Sync

- [x] 1.1 Hide Approvals footer navigation hints when the active section has zero or one row.
- [x] 1.2 Add a regression covering the reduced footer surface.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so the approvals guidance covers both help surfaces.
- [x] 2.2 Update the `approval-history-view` OpenSpec delta with the footer contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate approvals-footer-navigation-truth-sync --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
