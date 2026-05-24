## 1. Prompt Truth Fix

- [x] 1.1 Add regressions for `s`-confirm prompts in the inline strip and fullscreen dialog.
- [x] 1.2 Render the actual pending confirm key in both approval surfaces.

## 2. Spec Sync

- [x] 2.1 Update the chat-rendering spec to require key-accurate confirm prompts.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
