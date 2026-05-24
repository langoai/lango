## 1. Channel Badge Color Hardening

- [x] 1.1 Use sanitized channel text when selecting the badge color.
- [x] 1.2 Add regression coverage for escaped known channel names.

## 2. Spec Sync

- [x] 2.1 Record the sanitized badge-color key contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-channel-badge-color-key-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
