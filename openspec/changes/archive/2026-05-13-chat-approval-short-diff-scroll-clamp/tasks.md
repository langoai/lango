## 1. Scroll Clamp Fix

- [x] 1.1 Add a failing approval-dialog regression for non-overflow diff scrolling.
- [x] 1.2 Clamp approval-dialog scroll offset to the last meaningful visible start.

## 2. Spec Sync

- [x] 2.1 Update the chat-rendering spec to require non-overflow scroll clamping.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [ ] 3.4 Validate and archive the OpenSpec change.
