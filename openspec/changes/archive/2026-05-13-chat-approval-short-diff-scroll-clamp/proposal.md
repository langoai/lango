## Why

The fullscreen approval dialog now hides scroll help when a diff preview fits, but its internal scroll offset can still drift positive if the operator presses `↓` anyway. That can cause the rendered short diff to jump to a partial tail instead of staying pinned to the full visible content.

## What Changes

- Clamp fullscreen approval-dialog scroll offset to the last meaningful diff start position before rendering.
- Add regressions showing that non-overflow diffs stay pinned at offset zero and continue to render completely.
- Update the chat-rendering spec to require this no-overflow clamp behavior.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Fullscreen approval diff scrolling must clamp to the last meaningful visible start.

## Impact

- Affected code: `internal/cli/chat/approval_dialog.go`, `internal/cli/chat/approval_dialog_test.go`
- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
