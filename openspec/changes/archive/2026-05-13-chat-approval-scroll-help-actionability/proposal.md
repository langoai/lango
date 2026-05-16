## Why

The fullscreen approval dialog currently advertises `↑/↓ scroll` whenever diff content exists, even if the diff preview fully fits within the visible dialog body. That leaves the action bar exposing inert scroll keys in short-diff cases.

## What Changes

- Show approval-dialog `↑/↓ scroll` help only when the diff preview actually overflows the visible body.
- Add regression coverage for short-diff and long-diff dialog help.
- Update the chat-rendering spec to require conditional scroll-help actionability.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Fullscreen approval-dialog scroll help becomes conditional on actual diff overflow.

## Impact

- Affected code: `internal/cli/chat/approval_dialog.go`, `internal/cli/chat/approval_dialog_test.go`
- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
