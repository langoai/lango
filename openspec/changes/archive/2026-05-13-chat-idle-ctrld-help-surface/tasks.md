## 1. Help Surface Sync

- [x] 1.1 Add a chat regression for idle/failed `Ctrl+D` help visibility.
- [x] 1.2 Show `Ctrl+D` in the idle/failed chat help bar.

## 2. Spec Sync

- [x] 2.1 Update the chat-rendering spec to require idle/failed `Ctrl+D` help discoverability.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
