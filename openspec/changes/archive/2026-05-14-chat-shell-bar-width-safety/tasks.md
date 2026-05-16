## 1. Shell Bar Hardening

- [x] 1.1 Make the chat header render as a single-line width-safe bar.
- [x] 1.2 Make the turn status strip render as a single-line width-safe bar.
- [x] 1.3 Add regressions for narrow-width shell bar rendering.

## 2. Spec Sync

- [x] 2.1 Record the shell-bar single-line contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-shell-bar-width-safety --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
