## 1. Runtime Fix

- [x] 1.1 Add a failing Mission Control regression for composer/activity-lane `↑/↓` navigation.
- [x] 1.2 Route composer/activity-lane `↑/↓` keys to the activity cursor so the regression passes.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require focused activity-lane navigation behavior.
- [x] 2.2 Update cockpit feature docs to describe the same focused-lane navigation contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
