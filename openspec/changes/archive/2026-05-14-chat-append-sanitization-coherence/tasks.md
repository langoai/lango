## 1. Transcript Append Hardening

- [x] 1.1 Sanitize appended `system` and `status` content before storing it.
- [x] 1.2 Sanitize display-only channel/delegation metadata before storing it.
- [x] 1.3 Add regression coverage for stored display-safe transcript data.

## 2. Spec Sync

- [x] 2.1 Record the append-time sanitization coherence contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-append-sanitization-coherence --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
