## Why

The current TUI renders all agent activity (text streaming, tool execution, thinking, approvals) as undifferentiated text chunks in a single transcript. Users cannot distinguish whether the agent is thinking, running a tool, or waiting for approval. All approval requests use the same yellow banner regardless of risk level. Background tasks and multi-agent activity are invisible. This makes the cockpit feel like a chat window rather than an operational control surface.

## What Changes

- Add tool lifecycle visibility to the transcript: each tool call shows as a distinct item with running/success/error state and duration
- Add thinking/reasoning visibility: thinking blocks render as collapsible items with duration, plus a pending indicator for the submit-to-first-event gap
- Introduce 2-tier approval UX: inline strip for low-risk tools, fullscreen dialog with diff preview for dangerous operations (exec, file_edit, file_write)
- Add background task surface: a task strip in the chat view showing active tasks, plus a dedicated Tasks cockpit page with full status table
- Transform the footer from a simple help bar into an operational HUD showing mode, permission, model, context budget, and pending task count
- Extend `turnrunner.Request` with `OnToolCall`, `OnToolResult`, and `OnThinking` callbacks to surface runtime events to the TUI
- Refresh the color palette from border-heavy purple to semantic accent tokens (success/warning/danger/info)

## Capabilities

### New Capabilities
- `tui-tool-lifecycle`: Tool call transcript items with state machine (running/success/error/canceled/awaiting_approval), duration tracking, and per-tool renderers
- `tui-thinking-indicator`: Thinking/reasoning transcript items using `genai.Part.Thought`, collapsible rendering, and submit-to-first-event pending indicator
- `tui-approval-tiers`: 2-tier approval surface — inline strip (Tier 1) for safe/moderate tools, fullscreen dialog with diff preview (Tier 2) for dangerous filesystem/exec tools
- `tui-task-surface`: Background task strip in chat view and dedicated cockpit Tasks page with status table, keyboard navigation, and BackgroundManager wiring
- `tui-runtime-bridge`: Chat-internal bridge translating `turnrunner.Request` callbacks into typed Bubble Tea messages for tool/thinking/approval events

### Modified Capabilities
- `interactive-tui-chat`: Add new transcript item kinds (tool, thinking), new message types, pending indicator, task strip in View(), approval tier dispatch in stateApproving handler, and footer operational HUD
- `cockpit-shell`: Add Tasks page registration (PageTasks), Ctrl+5 keybinding, sidebar menu entry, BackgroundManager in Deps
- `channel-approval`: Extend `ApprovalRequest` with `SafetyLevel`, `Category`, `Activity` fields for tier classification
- `tool-middleware`: Populate new `ApprovalRequest` fields from tool metadata in approval middleware
- `cockpit-theme`: Add semantic color aliases (Danger, Info), reduce border-heavy styles to spacing/badge patterns

## Impact

- **Core runtime** (`internal/turnrunner/runner.go`): New callbacks on `Request` struct, `part.Thought` detection in event loop, `callID→startedAt` map for duration
- **Chat model** (`internal/cli/chat/`): All hub files modified — messages.go, chatview.go, chat.go, statusbar.go, approval.go. New files: bridge.go, render_tool.go, render_thinking.go, taskstrip.go, approval_strip.go (stub→real), approval_dialog.go (stub→real)
- **Approval system** (`internal/approval/`): New `viewmodel.go` with tier classification. `ApprovalRequest` gains 3 optional fields
- **Toolchain** (`internal/toolchain/mw_approval.go`): Populates new approval request fields
- **Cockpit** (`internal/cli/cockpit/`): New Tasks page, router/keymap/sidebar/deps updates, theme palette refresh
- **Entry point** (`cmd/lango/main.go`): Wiring BackgroundManager into cockpit Deps, Tasks page registration
- **Documentation**: README.md, docs/, skills/ updated to reflect new TUI capabilities
