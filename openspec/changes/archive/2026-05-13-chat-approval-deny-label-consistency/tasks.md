## 1. Wording Fix

- [x] 1.1 Add regressions for approval-surface deny labels.
- [x] 1.2 Standardize visible deny labels on `d/Esc`.

## 2. Spec Sync

- [x] 2.1 Update the chat-rendering spec to require consistent deny-key wording.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
