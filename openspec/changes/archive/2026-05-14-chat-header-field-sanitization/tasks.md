## 1. Header Field Hardening

- [x] 1.1 Sanitize provider, model, and session-key display values before rendering.
- [x] 1.2 Add regression coverage for multiline and escaped header field values.

## 2. Spec Sync

- [x] 2.1 Record the header-field sanitization contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-header-field-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
