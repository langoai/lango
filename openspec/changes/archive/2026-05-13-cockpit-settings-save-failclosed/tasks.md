## 1. UX Fix

- [x] 1.1 Make nil-store embedded save fail closed with an actionable error.
- [x] 1.2 Render explicit save-unavailable messaging in the Settings page when persistence is unavailable.
- [x] 1.3 Add regression coverage for the unavailable save path.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the degraded save behavior.
- [x] 2.2 Update cockpit settings-page spec to require fail-closed save behavior in degraded mode.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
