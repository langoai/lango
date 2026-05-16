## 1. Channel Sanitization Hardening

- [x] 1.1 Strip ANSI/OSC escape sequences from remote channel sender names before rendering.
- [x] 1.2 Add regression coverage for sender-name sanitization.

## 2. Spec Sync

- [x] 2.1 Record the remote sender/message sanitization contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-channel-sender-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
