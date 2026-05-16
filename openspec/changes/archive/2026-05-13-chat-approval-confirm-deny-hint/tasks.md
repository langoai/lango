## 1. Prompt Surface Fix

- [x] 1.1 Add regressions for confirm-pending deny hints in inline and fullscreen approval surfaces.
- [x] 1.2 Surface `d/Esc` denial in confirm-pending approval prompts.

## 2. Spec Sync

- [x] 2.1 Update the chat-rendering spec to require confirm-pending deny-path discoverability.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [ ] 3.4 Validate and archive the OpenSpec change.
