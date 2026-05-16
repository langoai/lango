## 1. Help Copy Sync

- [x] 1.1 Add a regression for the `/help` `Ctrl+C` description.
- [x] 1.2 Update `/help` copy to describe idle and failed double-press quit semantics.

## 2. Docs And Spec Sync

- [x] 2.1 Update the chat-rendering and downstream-docs-sync specs to require the same wording.
- [x] 2.2 Update CLI overview wording to match the runtime help text.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [ ] 3.4 Validate and archive the OpenSpec change.
