## 1. Request ID Compaction

- [x] 1.1 Compact long request IDs before appending approval transcript event annotations.
- [x] 1.2 Keep short request IDs unchanged.
- [x] 1.3 Add regressions for compact and unchanged cases.

## 2. Spec Sync

- [x] 2.1 Record the compact request-id formatting contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-requestid-compaction --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
