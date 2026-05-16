## 1. UX Fix

- [x] 1.1 Route `Enter` to retry submission while Dead Letters retry confirmation is pending.
- [x] 1.2 Add regression coverage for the runtime `Enter` confirm path.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec so the confirm behavior is pinned as runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
