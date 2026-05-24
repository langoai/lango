## 1. Help Surface Sync

- [x] 1.1 Add a regression for the fullscreen approval dialog split-toggle help.
- [x] 1.2 Surface `t split` in the fullscreen approval action bar when diff content exists.

## 2. Docs And Spec Sync

- [x] 2.1 Update the chat-rendering and downstream-docs-sync specs to require split-toggle discoverability.
- [x] 2.2 Update cockpit feature docs to describe the same Tier 2 split-toggle key.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
