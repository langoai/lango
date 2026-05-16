## 1. Compact Row Width Safety

- [x] 1.1 Make compact status transcript rows width-aware.
- [x] 1.2 Make compact approval transcript rows width-aware.
- [x] 1.3 Add regressions for narrow-width status and approval rows.

## 2. Spec Sync

- [x] 2.1 Record the width-safety contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-compact-status-width-safety --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
