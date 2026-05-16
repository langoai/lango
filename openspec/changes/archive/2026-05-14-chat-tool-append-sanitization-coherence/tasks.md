## 1. Tool Entry Hardening

- [x] 1.1 Sanitize stored tool-entry names at append time.
- [x] 1.2 Add regression coverage for stored sanitized tool names.

## 2. Spec Sync

- [x] 2.1 Record the tool-entry append-time coherence contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-tool-append-sanitization-coherence --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
