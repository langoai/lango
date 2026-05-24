## 1. UX Fix

- [x] 1.1 Render explicit unavailable messaging when DeadLettersPage has no list callback.
- [x] 1.2 Add regression coverage for the immediate unavailable view.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit page spec to require the immediate unavailable render path.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
