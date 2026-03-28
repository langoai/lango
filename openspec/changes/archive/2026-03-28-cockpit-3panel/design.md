## Context

Change-1+2 delivered a 2-panel cockpit (`lango cockpit`) with sidebar, 4 pages (Chat/Settings/Tools/Status), and keyboard navigation. Change-3 completes the TUI with a 3-panel layout, mouse support, session browsing, streaming UX, and makes cockpit the default entry point.

Key constraints from prior changes:
- `childModel` interface wraps ChatModel. Cockpit owns layout, child owns chat rendering.
- Page interface uses Activate/Deactivate lifecycle for resource management (e.g., StatusPage tick).
- Sidebar items hardcoded in `New()`. Focus guard (`!focused → return`) controls keyboard routing.
- `WithMouseCellMotion()` already enabled on tea.Program but zero MouseMsg handlers exist.
- turnrunner.Request has no OnToolStart/OnToolEnd callbacks — tool lifecycle events not available to UI.
- MetricsCollector.Snapshot() provides token/tool/uptime data. ContextBudgetManager not exposed on App.

## Goals / Non-Goals

**Goals:**
- 3-panel layout (sidebar + main + context) with Ctrl+P toggle and correct width propagation
- Mouse click sidebar navigation (coordinate-based, works regardless of focus state)
- Ctrl+Y clipboard copy of active main panel
- ContextPanel with live token/tool/uptime metrics (5s auto-refresh)
- ChatModel.RenderParts() for composable rendering + streaming cursor blink
- SessionsPage with session.Store.ListSessions() and SessionSummary
- `lango` default → cockpit, `lango chat` for legacy

**Non-Goals:**
- Tool lifecycle messages (ToolStartMsg/ToolEndMsg) — no producer path in turnrunner
- Toast/notification system — no triggering events without tool lifecycle
- Budget allocation display — ContextBudgetManager not on App struct
- Session switching (clicking a session to load it) — deferred
- Width-aware Editor rendering — deferred
- Mouse hover/drag interactions — only click-to-navigate

## Decisions

### D1: ContextPanel uses Start()/Stop(), not Activate()/Deactivate()
ContextPanel is NOT a Page. It's a persistent side panel managed by cockpit's Ctrl+P toggle. Using different method names avoids confusion with the Page interface lifecycle. Cockpit.go calls Start() when visible, Stop() when hidden.

Alternative: Make it a Page → rejected because it's always-visible when toggled, not switched via sidebar.

### D2: Mouse routing via coordinate hit-test, not bubblezone
Sidebar items each occupy 1 row. On MouseActionRelease, cockpit checks `msg.X < sidebarWidth()` and forwards to sidebar. Sidebar checks `msg.Y` to find the item index. Simple and dependency-free.

Alternative: Add bubblezone dependency → rejected to avoid new dependency for simple case.

### D3: Mouse handler BEFORE focused guard in sidebar
Current sidebar.Update() returns early when `!focused` for keyboard events. MouseMsg case is placed before the focused guard so clicks work regardless of focus state. Keyboard navigation still requires focus.

### D4: SessionSummary in session package, imported directly by pages
No import cycle exists between `internal/cli/cockpit/pages` → `internal/session`. SessionSummary struct lives in session/store.go. SessionsPage takes a callback `func(ctx) ([]session.SessionSummary, error)` following the StatusPage provider pattern.

### D5: Ctrl+P toggle sends synthetic resize to ALL components
Same pattern as existing Ctrl+B sidebar toggle. When context panel visibility changes, cockpit sends `tea.WindowSizeMsg{Width: total - sidebarWidth - contextPanelWidth, Height: h}` to child, all registered pages, and contextPanel itself.

### D6: Cursor blink with tick dedup guard
`chatView.cursorTickActive bool` prevents duplicate tick creation on rapid ChunkMsg. Tick starts on first ChunkMsg during streaming, toggles showCursor every 400ms, stops naturally when state leaves stateStreaming.

### D7: Default switch with OpenSpec update
Root RunE changes from `runChat()` to `runCockpit()`. New `lango chat` subcommand preserves legacy access. `tui-cockpit-layout/spec.md` gets MODIFIED section documenting the behavioral change.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Context panel + sidebar = narrow main area on small terminals | ContextPanel off by default. User toggles Ctrl+P only when wanted. Min main width ~60 chars. |
| Mouse Y-offset may drift if sidebar adds header/padding | Sidebar items start at Y=0 with current rendering. If changed, adjust offset constant. |
| ListSessions interface change ripples to 5 mock/stub files | All stubs return `nil, nil`. Minimal blast radius. |
| Clipboard write may fail silently (no clipboard on headless) | `_ = clipboard.WriteAll(...)` — best effort, no error surfaced. Acceptable for TUI. |
| Default switch breaks users expecting plain chat | `lango chat` preserves exact prior behavior. Cockpit includes chat as default page. |
| Cursor tick leaks if state transitions are missed | Tick checks `state == stateStreaming` on every fire. Self-stops if state changed. DoneMsg/ErrorMsg also call stopCursorBlink(). |
