## Why

The inline approval strip already accepts `Esc` as a deny key through the shared approval handler, but the compact help text still advertises only `[d]eny`. That leaves a small but real interaction mismatch between the visible affordance and the actual key path.

## What Changes

- Update the inline approval strip to advertise `Esc` alongside `d` for deny.
- Add regression coverage for the compact strip wording.
- Update the chat-rendering spec to require the same inline deny-key discoverability.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Inline approval strip help becomes explicit about the `Esc` deny path.

## Impact

- Affected code: `internal/cli/chat/approval_strip.go`, `internal/cli/chat/approval_strip_test.go`
- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
