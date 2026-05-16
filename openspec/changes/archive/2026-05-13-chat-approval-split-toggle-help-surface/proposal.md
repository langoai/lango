## Why

The fullscreen approval dialog already supports `t` to toggle unified versus split diff mode, but the action bar never advertises that key and the public docs do not mention it. That leaves a real operator-facing capability effectively hidden behind trial and error.

## What Changes

- Add `t split` help to the fullscreen approval dialog when diff content exists.
- Add regression coverage for the visible split-toggle help.
- Update the chat-rendering and downstream-docs-sync specs plus cockpit feature docs to describe the same approval-dialog key.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Fullscreen approval dialog help surfaces the `t` split-toggle key.
- `downstream-docs-sync`: Public cockpit docs describe the split-toggle key in Tier 2 approvals.

## Impact

- Affected code: `internal/cli/chat/approval_dialog.go`, `internal/cli/chat/approval_dialog_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/tui-chat-rendering/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
