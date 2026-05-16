## 1. Channel Row Hardening

- [x] 1.1 Normalize channel sender/text payloads to a single line.
- [x] 1.2 Clamp the final rendered channel transcript row to the available width.
- [x] 1.3 Add regressions for narrow-width and multiline channel rendering.

## 2. Spec Sync

- [x] 2.1 Record the channel-row width/single-line contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-channel-row-width-hardening --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
