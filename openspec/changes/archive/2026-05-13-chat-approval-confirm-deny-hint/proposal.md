## Why

Critical approval confirmation already allows `d` or `Esc` to deny immediately while a confirm-pending prompt is on screen, but the visible prompt only tells the operator to press the confirm key again. That leaves a valid escape path hidden during one of the highest-risk interaction states.

## What Changes

- Add deny-path wording to confirm-pending approval prompts in both the inline strip and fullscreen dialog.
- Add regressions for the prompt text.
- Update the chat-rendering spec to require confirm prompts to preserve deny-path discoverability.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Approval confirm prompts shall mention the live deny path.

## Impact

- Affected code: `internal/cli/chat/approval_strip.go`, `internal/cli/chat/approval_dialog.go`
- Affected tests/specs: `internal/cli/chat/approval_strip_test.go`, `internal/cli/chat/approval_dialog_test.go`, `openspec/specs/tui-chat-rendering/spec.md`
