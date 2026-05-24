## 1. Renderer Hardening

- [x] 1.1 Make fullscreen approval dialog rendering tolerate a nil approval state, including diff preview rendering.
- [x] 1.2 Add a regression covering diff rendering with a nil approval state.

## 2. Spec Sync

- [x] 2.1 Update the `tui-chat-rendering` delta to record nil-state fail-closed approval rendering.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-dialog-nil-state-hardening --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
