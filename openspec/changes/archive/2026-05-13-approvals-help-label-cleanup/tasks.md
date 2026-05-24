## 1. UX Fix

- [x] 1.1 Keep `Tab` and `/` section toggles but render a cleaner help label.
- [x] 1.2 Update approvals help regressions for the new label.

## 2. Docs And Spec Sync

- [x] 2.1 Update the approvals spec to require a readable dual-key help hint.
- [x] 2.2 Confirm cockpit docs remain aligned with the same key wording.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
