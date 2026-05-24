## 1. Turn Strip Copy Sync

- [x] 1.1 Update the approval-state turn strip hint to use `d/Esc`.
- [x] 1.2 Add regression coverage for the approving-state copy.

## 2. Spec Sync

- [x] 2.1 Record the approval-state turn strip deny-label contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-turnstrip-deny-label-sync --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
