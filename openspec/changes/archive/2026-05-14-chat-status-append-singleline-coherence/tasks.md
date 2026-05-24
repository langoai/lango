## 1. Status Entry Hardening

- [x] 1.1 Normalize appended status content to display-safe single-line text.
- [x] 1.2 Add regression coverage for stored multiline status content.

## 2. Spec Sync

- [x] 2.1 Record the status-entry append-time coherence contract in the `tui-chat-rendering` delta.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-status-append-singleline-coherence --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
