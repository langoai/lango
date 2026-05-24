## Why

The chat runtime always supports `Ctrl+D` as the immediate quit path, but the idle/failed in-product help bar does not currently advertise it. That leaves the most direct quit key discoverable only through `/help` or the first `Ctrl+C` warning instead of the primary help surface.

## What Changes

- Add `Ctrl+D` to the idle/failed chat help bar.
- Add regression coverage for the idle help surface.
- Update the chat-rendering spec to require the same immediate-quit discoverability.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Idle/failed chat help shall advertise the immediate `Ctrl+D` quit path.

## Impact

- Affected code: `internal/cli/chat/statusbar.go`, `internal/cli/chat/statusbar_test.go`
- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
