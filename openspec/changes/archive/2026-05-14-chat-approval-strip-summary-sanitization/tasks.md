## 1. Approval Strip Hardening

- [x] 1.1 Sanitize inline approval-strip summary text before truncation.
- [x] 1.2 Add regression coverage for multiline and escaped summaries.

## 2. Spec Sync

- [x] 2.1 Record the inline approval-strip summary contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-strip-summary-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
