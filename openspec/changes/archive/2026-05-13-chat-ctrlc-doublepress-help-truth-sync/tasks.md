## 1. Help Truth Sync

- [x] 1.1 Add a chat regression for idle/failed `Ctrl+C` help wording.
- [x] 1.2 Update chat help surfaces to describe double-press quit semantics.

## 2. Docs And Spec Sync

- [x] 2.1 Update the chat-rendering and downstream-docs-sync specs to require the same `Ctrl+C` contract.
- [x] 2.2 Update cockpit feature docs and CLI overview wording to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
