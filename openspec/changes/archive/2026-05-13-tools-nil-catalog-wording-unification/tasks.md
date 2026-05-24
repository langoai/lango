## 1. UX Fix

- [x] 1.1 Unify the nil-catalog unavailable wording across both Tools page panes.
- [x] 1.2 Update regressions for the shared wording.

## 2. Docs And Spec Sync

- [x] 2.1 Update the Tools page spec to reflect the unified unavailable wording.
- [x] 2.2 Update cockpit feature docs to describe the same unavailable wording.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
