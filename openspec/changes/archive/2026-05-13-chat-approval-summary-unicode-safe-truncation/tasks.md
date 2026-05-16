## 1. Unicode-Safe Summary Truncation

- [x] 1.1 Replace byte slicing in approval transcript summary truncation with Unicode-safe truncation.
- [x] 1.2 Add a regression covering non-ASCII summary previews.

## 2. Spec Sync

- [x] 2.1 Record the Unicode-safe truncation contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-summary-unicode-safe-truncation --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
