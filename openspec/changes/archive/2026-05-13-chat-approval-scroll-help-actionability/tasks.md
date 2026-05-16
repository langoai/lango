## 1. Help Actionability Fix

- [x] 1.1 Add approval-dialog regressions for short-diff and long-diff action bars.
- [x] 1.2 Hide approval-dialog scroll help when the diff preview has no overflow.

## 2. Spec Sync

- [x] 2.1 Update the chat-rendering spec to require conditional approval-dialog scroll help.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
