## 1. Retry Running Hint Sync

- [x] 1.1 Update the Dead Letters running-state filter hint so it does not advertise `Ctrl+R`.
- [x] 1.2 Add a regression for the running-state filter hint text.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so the Dead Letters running-state guidance covers the filter hint line as well as the help bar.
- [x] 2.2 Update the `cockpit-pages` OpenSpec delta with the running-state hint contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate deadletters-running-filter-hint-sync --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
