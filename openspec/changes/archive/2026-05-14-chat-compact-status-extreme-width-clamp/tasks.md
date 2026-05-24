## 1. Extreme Width Clamp

- [x] 1.1 Clamp the final rendered compact status/approval row to the requested width.
- [x] 1.2 Add regressions for very narrow width values.

## 2. Spec Sync

- [x] 2.1 Record the extreme-width clamp contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-compact-status-extreme-width-clamp --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
