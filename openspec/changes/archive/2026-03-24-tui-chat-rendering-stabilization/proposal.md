## Why

TUI chat has three rendering bugs: input box border triple-duplication, AI response not displayed after completion, and excessive blank lines between messages. These make the chat interface unusable for production. The fix applies crush-aligned structural patterns (parts-based layout, block-join rendering) scoped strictly to rendering stabilization.

## What Changes

- **View/Layout unification**: `View()` and `recalcLayout()` share the same parts structure so measured heights always match rendered output.
- **Input width safety margin**: `SetWidth()` subtracts 2 for border padding with minimum clamp, preventing border wrap/triplication.
- **Block-join chat rendering**: `render()` uses `strings.Join(blocks, "\n\n")` instead of per-entry `\n` prefix, eliminating leading blank line accumulation.
- **Assistant rawContent preservation**: `chatEntry.rawContent` stores original markdown for resize reflow; `appendAssistant()` is the single entry point for all assistant messages.
- **DoneMsg 3-rule processing**: (1) finalize stream if present, (2) fall back to ResponseText for non-streaming, (3) add system error with deduplication.
- **ErrorMsg partial-first**: preserve in-flight stream before adding error system message.
- **Approval banner layout**: `recalcLayout()` called on approval state change; banner width clamped.

## Capabilities

### New Capabilities
- `tui-chat-rendering`: Stable full-screen TUI chat layout with parts-based View/Layout agreement, block-join message rendering, and resize-aware markdown reflow.

### Modified Capabilities

## Impact

- `internal/cli/chat/chat.go` — View(), recalcLayout(), DoneMsg/ErrorMsg/ApprovalRequestMsg handlers
- `internal/cli/chat/input.go` — SetWidth() margin and clamp
- `internal/cli/chat/chatview.go` — chatEntry struct, appendAssistant(), finalizeStream(), contentWidth(), render()
- `internal/cli/chat/approval.go` — renderApprovalBanner() width clamp
- No changes to runtime packages (internal/app/, internal/graph/, etc.)
