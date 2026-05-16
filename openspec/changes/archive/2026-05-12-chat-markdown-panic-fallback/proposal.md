## Why

A full `go test ./...` run exposed a real transcript stability bug: Glamour rendering can panic while `Mission Control` replays immediate assistant output, which crashes the whole test process and would crash the TUI in production.

## What Changes

- Make `renderMarkdown()` fail closed to plain text when the markdown renderer returns an error or panics.
- Add direct regressions for both the error and panic fallback paths.
- Sync the TUI chat rendering spec so the plain-text fallback becomes part of the rendering contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `tui-chat-rendering`: markdown rendering now recovers to plain text when Glamour errors or panics.

## Impact

- Affected code: `internal/cli/chat/markdown.go`
- Affected tests: `internal/cli/chat/markdown_test.go`
- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
