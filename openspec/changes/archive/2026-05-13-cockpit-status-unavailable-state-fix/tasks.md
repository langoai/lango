## 1. UX Fix

- [x] 1.1 Show an explicit unavailable message when the feature-status provider is absent.
- [x] 1.2 Show explicit unavailable messaging in metrics sections when the observability collector is absent.
- [x] 1.3 Add regression coverage for both unavailable states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit status-page spec to require the unavailable-state messaging.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
