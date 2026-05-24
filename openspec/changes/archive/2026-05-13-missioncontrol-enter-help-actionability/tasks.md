## 1. UX Fix

- [x] 1.1 Hide `Enter` help when the focused Mission Control lane has no actionable Enter behavior.
- [x] 1.2 Add regression coverage for non-proposal mission rows and actionable proposal/composer states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require `Enter` help only when an Enter action exists.
- [x] 2.2 Update cockpit feature docs to describe the same actionability rule.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
