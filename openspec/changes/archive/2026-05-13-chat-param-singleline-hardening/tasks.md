## 1. Single-Line Param Hardening

- [x] 1.1 Normalize approval and tool-preview param values into single-line-safe display text.
- [x] 1.2 Add regressions covering multiline string params.

## 2. Spec Sync

- [x] 2.1 Update the `tui-chat-rendering` delta with the single-line param rendering contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-param-singleline-hardening --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
