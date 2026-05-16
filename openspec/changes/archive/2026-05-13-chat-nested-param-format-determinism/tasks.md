## 1. Deterministic Nested Param Formatting

- [x] 1.1 Format nested approval param values deterministically across banner and fullscreen dialog.
- [x] 1.2 Reuse the same deterministic formatter for tool transcript previews.
- [x] 1.3 Add regressions covering nested structured payloads.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so deterministic param rendering explicitly covers structured payloads.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the nested-format contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-nested-param-format-determinism --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
