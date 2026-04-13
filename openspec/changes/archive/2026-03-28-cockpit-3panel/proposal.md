## Why

The cockpit TUI (Change-1+2) has a 2-panel layout (sidebar + main), 4 pages, and keyboard-only sidebar navigation. Users cannot see live metrics without switching to StatusPage, cannot use mouse for navigation, cannot browse sessions, and must explicitly launch `lango cockpit`. Adding a context panel, mouse support, session browsing, chat UX polish, and making cockpit the default completes the TUI as a production-ready experience.

## What Changes

- Extract ChatModel.View() into public RenderParts() for composable rendering
- Add streaming cursor blink animation (400ms tea.Tick) during agent responses
- Add toggleable right context panel (Ctrl+P) showing live token usage, tool stats, and uptime
- Extend cockpit from 2-panel to 3-panel layout (sidebar + main + context)
- Add mouse click-to-navigate for sidebar (coordinate-based hit testing)
- Add Ctrl+Y clipboard copy for active main panel content
- Add session.Store.ListSessions() with SessionSummary return type
- Create SessionsPage with cursor-navigable session list and relative timestamps
- Enable sessions sidebar item
- Switch `lango` default from plain chat to cockpit, add `lango chat` subcommand for legacy access
- Synthetic WindowSizeMsg propagation on Ctrl+P toggle (same pattern as Ctrl+B sidebar toggle)

## Capabilities

### New Capabilities
- `cockpit-context-panel`: Toggleable right-side panel with live token/tool/uptime metrics, Start/Stop lifecycle, 5s auto-refresh
- `cockpit-sessions-page`: SessionsPage component with session listing, cursor navigation, relative timestamps
- `chat-render-parts`: ChatParts struct and RenderParts() method for composable chat view sections
- `chat-cursor-blink`: Streaming cursor blink animation with CursorTickMsg and tick dedup guard

### Modified Capabilities
- `cockpit-shell`: 3-panel layout, mouse routing (tea.MouseMsg → sidebar hit-test), Ctrl+P context toggle with synthetic resize, Ctrl+Y clipboard
- `cockpit-sidebar`: Mouse click handler (before focused guard), sessions item enabled
- `tui-cockpit-layout`: Default entry point changed from single-column chat to multi-panel cockpit, `lango chat` for legacy

## Impact

- **Modified**: `internal/cli/chat/chat.go`, `chatview.go`, `messages.go` — RenderParts + cursor blink
- **Modified**: `internal/cli/cockpit/cockpit.go` — 3-panel layout, context panel, mouse routing, clipboard
- **Modified**: `internal/cli/cockpit/keymap.go` — Ctrl+P, Ctrl+Y bindings
- **Modified**: `internal/cli/cockpit/sidebar/sidebar.go` — mouse handler, sessions enabled
- **Modified**: `internal/cli/cockpit/theme/theme.go` — ContextPanelWidth constant
- **Modified**: `internal/cli/cockpit/router.go` — PageSessions enum
- **Modified**: `internal/session/store.go`, `ent_store.go` — SessionSummary + ListSessions
- **Modified**: 5 mock/stub test files — ListSessions interface compliance
- **Modified**: `cmd/lango/main.go` — default switch, chatCmd, SessionsPage registration
- **New**: `internal/cli/cockpit/contextpanel.go` — standalone ContextPanel tea.Model
- **New**: `internal/cli/cockpit/pages/sessions.go` — SessionsPage component
