## Why

The chat runtime already requires double-press `Ctrl+C` to quit from idle or failed states, but the visible help surfaces still imply that a single `Ctrl+C` quits immediately. That leaves the in-product help and public docs overstating the actual quit behavior.

## What Changes

- Update idle/failed chat help copy to describe `Ctrl+C` as a double-press quit path.
- Update `/help` key-binding text to describe the same `Ctrl+C` behavior.
- Sync the cockpit feature docs and CLI overview with the same quit semantics.
- Add or extend specs so the runtime and docs share the same contract.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Chat help copy for idle/failed `Ctrl+C` becomes double-press-aware.
- `downstream-docs-sync`: Public docs for the cockpit Chat surface and CLI overview need the same `Ctrl+C` quit semantics.

## Impact

- Affected code: `internal/cli/chat/statusbar.go`, `internal/cli/chat/commands.go`, `internal/cli/chat/statusbar_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `docs/cli/core.md`, `openspec/specs/tui-chat-rendering/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
