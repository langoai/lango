## 1. Assistant Raw Hardening

- [x] 1.1 Strip control sequences from stored assistant raw content while preserving markdown/newlines.
- [x] 1.2 Use sanitized assistant text for duplicate failure-status suppression.
- [x] 1.3 Add regression coverage for stored sanitized assistant raw content and duplicate suppression.

## 2. Spec Sync

- [x] 2.1 Record the assistant-raw coherence contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-assistant-raw-sanitization-coherence --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
