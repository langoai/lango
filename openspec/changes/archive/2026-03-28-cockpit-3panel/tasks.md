## 1. Chat RenderParts + Cursor Blink

- [x] 1.1 Add `ChatParts` struct and `RenderParts()` method to ChatModel (`chat/chat.go`)
- [x] 1.2 Refactor `View()` to call `RenderParts()` and join sections — no behavior change
- [x] 1.3 Add `CursorTickMsg` type to `chat/messages.go`
- [x] 1.4 Add `showCursor`, `cursorTickActive` fields to `chatViewModel`, add `stopCursorBlink()` helper (`chat/chatview.go`)
- [x] 1.5 Start cursor tick on first `ChunkMsg` during streaming with dedup guard
- [x] 1.6 Handle `CursorTickMsg`: toggle cursor during streaming, stop when not streaming
- [x] 1.7 Call `stopCursorBlink()` on `DoneMsg` and `ErrorMsg`
- [x] 1.8 Render "▌" block cursor after stream content when `showCursor=true`
- [x] 1.9 Tests for RenderParts and cursor blink lifecycle

## 2. Context Panel + 3-Panel Layout

- [x] 2.1 Create `contextpanel.go` with `ContextPanel` tea.Model: NewContextPanel, Start/Stop, SetHeight/SetVisible, 5s tick refresh
- [x] 2.2 Render token usage (input/output/total/cache), top-5 tool stats, uptime in ContextPanel.View()
- [x] 2.3 Add `ContextPanelWidth = 28` to `theme/theme.go`
- [x] 2.4 Add `ToggleContext` (ctrl+p) and `CopyClipboard` (ctrl+y) to `keymap.go`
- [x] 2.5 Add `contextPanel` and `contextVisible` fields to cockpit Model, create panel in `New()`
- [x] 2.6 Implement Ctrl+P toggle: Start/Stop + synthetic WindowSizeMsg to child, all pages, contextPanel
- [x] 2.7 Update `WindowSizeMsg` handler to propagate to contextPanel and account for `contextPanelWidth()`
- [x] 2.8 Update `View()` to 3-panel: `JoinHorizontal(sidebar, main, contextPanel)` when visible
- [x] 2.9 Tests for context toggle, 3-panel resize propagation, panel visibility

## 3. Mouse Sidebar + Clipboard

- [x] 3.1 Add `tea.MouseMsg` handler in sidebar.Update() BEFORE focused guard (coordinate hit-test: Y→item index, skip disabled)
- [x] 3.2 Add `tea.MouseMsg` case in cockpit.Update() BEFORE KeyMsg: forward to sidebar if X < sidebarWidth, else to active page/child
- [x] 3.3 Implement Ctrl+Y clipboard handler: copy child.View() or pages[activePage].View()
- [x] 3.4 Tests for mouse click navigation, disabled item click, out-of-bounds click, clipboard

## 4. Session ListSessions + Sessions Page

- [x] 4.1 Add `SessionSummary` struct to `session/store.go`
- [x] 4.2 Add `ListSessions(ctx) ([]SessionSummary, error)` to `session.Store` interface
- [x] 4.3 Implement `ListSessions` in `session/ent_store.go` (order by UpdatedAt desc)
- [x] 4.4 Add `ListSessions` stub to 5 mock/stub files: testutil/mock_session_store.go, session/child_test.go, gateway/middleware_test.go, app/wiring_automation_test.go, turnrunner/runner_test.go
- [x] 4.5 Create `pages/sessions.go`: SessionsPage with listFn callback, cursor navigation, relativeTime
- [x] 4.6 Add `PageSessions` to router.go enum, String(), PageIDFromString()
- [x] 4.7 Enable sessions sidebar item (Disabled: false in sidebar.go)
- [x] 4.8 Register SessionsPage in main.go runCockpit()
- [x] 4.9 Tests for SessionsPage (activate, load, cursor, view, relativeTime)

## 5. Default Switch + Docs

- [x] 5.1 Change root RunE from `runChat()` to `runCockpit()` in main.go
- [x] 5.2 Add `chatCmd()` function and register in "start" group
- [x] 5.3 Update `cockpitCmd()` short description
- [x] 5.4 Update `tui-cockpit-layout/spec.md` with MODIFIED default entry point

## 6. OpenSpec Specs

- [x] 6.1 Create `cockpit-context-panel/spec.md` (NEW)
- [x] 6.2 Create `cockpit-sessions-page/spec.md` (NEW)
- [x] 6.3 Create `chat-render-parts/spec.md` (NEW)
- [x] 6.4 Create `chat-cursor-blink/spec.md` (NEW)
- [x] 6.5 Create `cockpit-shell/spec.md` delta (MODIFIED: 3-panel, mouse, clipboard, context toggle)
- [x] 6.6 Create `cockpit-sidebar/spec.md` delta (MODIFIED: mouse handler, sessions enabled)
- [x] 6.7 Create `tui-cockpit-layout/spec.md` delta (MODIFIED: default entry point)

## 7. Build Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./internal/cli/chat/... ./internal/cli/cockpit/... ./internal/session/... ./cmd/lango/...` passes
- [x] 7.3 `go vet ./internal/cli/chat/... ./internal/cli/cockpit/... ./internal/session/... ./cmd/lango/...` passes
