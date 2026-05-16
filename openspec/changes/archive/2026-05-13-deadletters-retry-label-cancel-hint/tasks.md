## 1. UX Fix

- [x] 1.1 Update the Dead Letters retry-confirm label to mention the cancel key.
- [x] 1.2 Update regressions for the retry-confirm detail text.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the confirm-state label to mention confirm and cancel paths.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
