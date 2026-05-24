## Why

The chat approval surfaces now mostly use `allow session` wording consistently, but the inline approval strip still abbreviates the `s` action as `[s]ession`. That leaves the same approval action labeled differently depending on which approval surface the operator sees.

## What Changes

- Change the inline approval strip label from `[s]ession` to `[s] allow session`.
- Add regression coverage for the inline strip wording.
- Update the chat-rendering spec to require the unified session-approval wording across approval surfaces.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Inline approval strip wording becomes consistent with the rest of the chat approval surface.

## Impact

- Affected code: `internal/cli/chat/approval_strip.go`, `internal/cli/chat/approval_strip_test.go`
- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
