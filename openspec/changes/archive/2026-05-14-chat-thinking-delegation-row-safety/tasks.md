## 1. Transcript Row Hardening

- [x] 1.1 Make `thinking` rows clamp to the available width and normalize multiline preview content to one line.
- [x] 1.2 Make `delegation` rows clamp to the available width and normalize multiline actor/reason text to one line.
- [x] 1.3 Add regressions for narrow-width and multiline rendering.

## 2. Spec Sync

- [x] 2.1 Record the new `thinking`/`delegation` row rendering contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-thinking-delegation-row-safety --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
