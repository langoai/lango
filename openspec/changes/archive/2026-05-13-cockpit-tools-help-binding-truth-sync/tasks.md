## 1. UX Fix

- [x] 1.1 Remove the stale `Esc back` binding from the Tools page help.
- [x] 1.2 Update tests to match the actual binding set.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit tools-page spec to match the actual help/navigation contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
