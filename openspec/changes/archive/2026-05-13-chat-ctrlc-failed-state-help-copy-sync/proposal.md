## Why

The chat runtime uses the same double-press `Ctrl+C` quit path in both idle and failed states, but the `/help` key-binding text still describes that behavior only for idle. That leaves a small but real mismatch between the visible help text and the actual state machine.

## What Changes

- Update `/help` key-binding text to describe the double-press quit path for idle and failed states.
- Add regression coverage for the `/help` output.
- Align the CLI overview wording with the same idle/failed semantics.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: `/help` key-binding text for `Ctrl+C` becomes failed-state-aware.
- `downstream-docs-sync`: CLI overview wording for Chat quit semantics stays aligned with runtime help text.

## Impact

- Affected code: `internal/cli/chat/commands.go`, `internal/cli/chat/chat_test.go`
- Affected docs/specs: `docs/cli/core.md`, `openspec/specs/tui-chat-rendering/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
