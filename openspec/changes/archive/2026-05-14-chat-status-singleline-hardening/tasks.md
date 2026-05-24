## 1. Compact Row Hardening

- [x] 1.1 Normalize compact status row content to single-line-safe text.
- [x] 1.2 Ensure approval event rows inherit the same single-line-safe normalization.
- [x] 1.3 Add regressions for multiline compact-row content.

## 2. Spec Sync

- [x] 2.1 Record the single-line compact-row contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-status-singleline-hardening --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
