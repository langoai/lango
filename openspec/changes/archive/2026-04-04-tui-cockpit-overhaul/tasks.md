## 1. Wave 1a: ApprovalViewModel + Tier Classification + Request Field Extension

- [x] 1.1 Add `SafetyLevel`, `Category`, `Activity` string fields (omitempty) to `ApprovalRequest` in `internal/approval/approval.go`
- [x] 1.2 Create `internal/approval/viewmodel.go` with `DisplayTier`, `RiskIndicator`, `ApprovalViewModel` types
- [x] 1.3 Implement `ClassifyTier(safetyLevel, category, activity string) DisplayTier` — Fullscreen for dangerous+filesystem/automation/execute/write, Inline otherwise
- [x] 1.4 Implement `ComputeRisk(safetyLevel, category string) RiskIndicator`
- [x] 1.5 Update `internal/toolchain/mw_approval.go` to populate `req.SafetyLevel`, `req.Category`, `req.Activity` from tool metadata
- [x] 1.6 Create `internal/approval/viewmodel_test.go` with table-driven tests for all SafetyLevel × Category × Activity combinations
- [x] 1.7 Verify `go build ./...` and `go test ./...` pass

## 2. Wave 1b: Semantic Palette

- [x] 2.1 Add semantic color aliases (`Danger`, `Info`, `Selection`) to `internal/cli/cockpit/theme/theme.go`
- [x] 2.2 Add `BadgeStyle(color)` and `DividerStyle` helpers to `internal/cli/tui/styles.go`
- [x] 2.3 Reduce border-heavy styles to spacing/badge alternatives in `styles.go`
- [x] 2.4 Verify `go build ./...` passes — no existing consumers broken

## 3. Wave 2: turnrunner.Request Extension

- [x] 3.1 Add `OnToolCall func(callID, toolName string, params map[string]any)` to `turnrunner.Request` in `internal/turnrunner/runner.go`
- [x] 3.2 Add `OnToolResult func(callID, toolName string, success bool, duration time.Duration, preview string)` to `turnrunner.Request`
- [x] 3.3 Add `OnThinking func(agentName string, started bool, summary string)` to `turnrunner.Request`
- [x] 3.4 Add runner-local `map[string]time.Time` for callID→startedAt tracking in `Run()`
- [x] 3.5 In `recordEvent()` (line ~493): add `part.FunctionCall` branch to call `req.OnToolCall` with callID, tool name, params; record startedAt
- [x] 3.6 In `recordEvent()`: add `part.FunctionResponse` branch to call `req.OnToolResult` with callID, tool name, success, computed duration, output preview; clean up startedAt entry
- [x] 3.7 In `recordEvent()`: add `part.Thought == true` branch to call `req.OnThinking` with agent name and thought text
- [x] 3.8 Ensure all new callbacks are nil-safe (skip if nil)
- [x] 3.9 Add tests in `runner_test.go` for new callback invocation timing and nil safety
- [x] 3.10 Verify `go build ./...` and `go test ./...` pass

## 4. Wave 3: Chat Model Integration (Hub File Exclusive)

- [x] 4.1 Add `ToolStartedMsg`, `ToolFinishedMsg`, `ThinkingStartedMsg`, `ThinkingFinishedMsg`, `TaskStripTickMsg`, `PendingIndicatorTickMsg` to `internal/cli/chat/messages.go`
- [x] 4.2 Add `ViewModel approval.ApprovalViewModel` field to `ApprovalRequestMsg` in `messages.go`
- [x] 4.3 Create `internal/cli/chat/bridge.go` with `enrichRequest(program *tea.Program, req *turnrunner.Request)` — wires OnToolCall/OnToolResult/OnThinking to tea messages without overwriting existing callbacks
- [x] 4.4 Add `itemTool` and `itemThinking` to `transcriptItemKind` in `internal/cli/chat/chatview.go`
- [x] 4.5 Implement `appendToolStart(callID, toolName, params)` and `finalizeToolResult(callID, success, duration, output)` in `chatview.go`
- [x] 4.6 Implement `appendThinking()` and `finalizeThinking(summary, duration)` in `chatview.go`
- [x] 4.7 Add render dispatch for `itemTool` and `itemThinking` in `chatview.render()`
- [x] 4.8 Create `internal/cli/chat/render_tool.go` with `ToolItemState` enum and `renderToolBlock()` — state icons: ⚙ running, ✓ success, ✗ error, ⊘ canceled, 🔒 awaiting
- [x] 4.9 Create `internal/cli/chat/render_thinking.go` with `renderThinkingBlock()` (collapsible) and `renderPendingIndicator()`
- [x] 4.10 Add `BackgroundManager *background.Manager` to chat `Deps` struct in `chat.go`
- [x] 4.11 Add Update() cases for `ToolStartedMsg`, `ToolFinishedMsg`, `ThinkingStartedMsg`, `ThinkingFinishedMsg`, `TaskStripTickMsg`, `PendingIndicatorTickMsg` in `chat.go`
- [x] 4.12 Wire `enrichRequest(m.program, &req)` in `submitCmd()` and start pending indicator tick on submit
- [x] 4.13 Create `internal/cli/chat/taskstrip.go` — `taskStripModel` with BackgroundManager ref, 2s tick, nil-safe View()
- [x] 4.14 Include task strip in `ChatModel.View()` / `RenderParts()` between main and footer
- [x] 4.15 Transform `renderFooter()` in `statusbar.go` into operational HUD: mode, permission, model, context budget, pending tasks; condensed for width < 80
- [x] 4.16 Add approval tier dispatch in `approval.go`: `TUIApprovalProvider.RequestApproval()` creates `ApprovalViewModel`, `renderApproval()` switches on tier
- [x] 4.17 Create `internal/cli/chat/approval_strip.go` as stub — `renderApprovalStrip()` delegates to `renderApprovalBanner()`
- [x] 4.18 Create `internal/cli/chat/approval_dialog.go` as stub — `renderApprovalDialog()` delegates to `renderApprovalBanner()`, `handleApprovalDialogKey()` returns nil, `scrollApprovalDialog()` is no-op
- [x] 4.19 Pre-plant tier 2 key dispatch in `handleApprovingKey()` calling stub `handleApprovalDialogKey()` and `scrollApprovalDialog()`
- [x] 4.20 Add headless tests: tool item append/state/render, thinking item, pending indicator, footer HUD, task strip nil-manager, approval tier dispatch, bridge msg routing
- [x] 4.21 Verify `go build ./...` and `go test ./...` pass

## 5. Wave 4: Approval Surfaces + Cockpit Wiring + Documentation

- [x] 5.1 Replace `internal/cli/chat/approval_strip.go` stub with real Tier 1 renderer — single-line compact strip with tool name, summary, action keys
- [x] 5.2 Replace `internal/cli/chat/approval_dialog.go` stub with real Tier 2 renderer — fullscreen overlay with risk badge, params, diff viewport, action bar, `t` toggle, `↑↓` scroll, 500-line diff cap
- [x] 5.3 Create `internal/cli/cockpit/pages/tasks.go` — `TasksPage` implementing `Page`, table view, 2s tick, Activate/Deactivate lifecycle, nil-manager fallback
- [x] 5.4 Add `PageTasks PageID = 5` to `internal/cli/cockpit/router.go`
- [x] 5.5 Add `Page5 key.Binding` (Ctrl+5) to `internal/cli/cockpit/keymap.go`
- [x] 5.6 Add "Tasks" menu entry to `internal/cli/cockpit/sidebar/sidebar.go`
- [x] 5.7 Add `BackgroundManager *background.Manager` to `internal/cli/cockpit/deps.go`
- [x] 5.8 Add Page5 key handling and Tasks page registration with BackgroundManager in `internal/cli/cockpit/cockpit.go`
- [x] 5.9 Update `cmd/lango/main.go` runCockpit(): add `BackgroundManager: application.BackgroundManager` to Deps, register Tasks page
- [x] 5.10 Add headless tests: approval strip/dialog render, tier dispatch end-to-end, tasks page render, cockpit routing, sidebar item count
- [x] 5.11 Update `README.md` with new TUI capabilities (tool indicator, approval tiers, task surface, footer HUD) — verify against actual code
- [x] 5.12 Update `docs/` with cockpit navigation, approval tiers, background tasks documentation
- [x] 5.13 Sync related `skills/` and prompts if applicable
- [x] 5.14 Verify `go build ./...` and `go test ./...` pass
- [x] 5.15 Run OpenSpec workflow: `opsx:verify` → `opsx:sync` → `opsx:archive`
