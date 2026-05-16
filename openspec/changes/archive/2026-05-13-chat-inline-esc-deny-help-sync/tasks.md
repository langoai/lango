## 1. Help Surface Sync

- [x] 1.1 Add a regression for the inline strip deny label.
- [x] 1.2 Update the inline strip to advertise `Esc` alongside `d`.

## 2. Spec Sync

- [x] 2.1 Update the chat-rendering spec to require inline `Esc` deny discoverability.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
