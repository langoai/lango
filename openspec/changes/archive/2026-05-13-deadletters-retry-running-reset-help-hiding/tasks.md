## 1. Help Actionability Fix

- [x] 1.1 Add a Dead Letters regression for retry-running help.
- [x] 1.2 Hide `Ctrl+R` help while retry work is active.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require retry-running reset-help hiding.
- [x] 2.2 Update cockpit feature docs to describe the same running-state help rule.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
