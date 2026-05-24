## 1. Turn Strip Sync

- [x] 1.1 Update the approving-state turn strip hint to mention `d/Esc` deny and `Ctrl+D` quit.
- [x] 1.2 Add a regression covering the approving-state hint copy.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so approval-state chat guidance covers the same deny/quit hint wording.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the turn-strip contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-turnstrip-hint-sync --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
