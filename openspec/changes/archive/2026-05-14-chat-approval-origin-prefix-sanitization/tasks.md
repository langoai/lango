## 1. Origin Prefix Hardening

- [x] 1.1 Sanitize the session-key channel prefix before origin/badge matching.
- [x] 1.2 Add regression coverage for escaped known session-key prefixes.

## 2. Spec Sync

- [x] 2.1 Record the sanitized origin-prefix matching contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-origin-prefix-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
