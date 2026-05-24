## 1. Help Surface Fix

- [x] 1.1 Add `Ctrl+D quit` to the approving-state help bar.
- [x] 1.2 Add a regression covering the approving-state help surface.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so chat key guidance explicitly covers `Ctrl+D` during approval.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the approving-state quit contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approving-ctrld-help-surface --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
