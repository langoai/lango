## Context

Change-1 delivered a 2-panel cockpit (`lango cockpit`) with a non-interactive sidebar and ChatModel wrapper. The cockpit uses a `childModel` interface with consume-or-forward delegation. Now we add 3 pages (Settings, Status, Tools) and make the sidebar interactive.

Key constraints from Change-1 review rounds:
- Editor save triggers via `handleMenuSelection("save")` → `Completed=true` + `tea.Quit`. Embedded mode must intercept this to call OnSave without quitting.
- Editor width/height are stored but not used in rendering. Width-aware layout deferred to Change-3.
- ToolCatalog, MetricsCollector, FeatureStatuses are already on App with complete APIs.
- RetrievalCoordinator/Adjuster stats and BudgetManager access deferred to follow-up PRs.

## Goals / Non-Goals

**Goals:**
- 3 pages (Settings, Status, Tools) navigable via sidebar and Ctrl+1-4
- Interactive sidebar with focus ring, cursor navigation, disabled items
- Settings save without program exit (OnSave callback)
- Save result feedback (inline banner)
- StatusPage auto-refresh via tea.Tick with Activate/Deactivate lifecycle

**Non-Goals:**
- SessionsPage (Change-3 — needs session.Store.ListKeys)
- Width-aware Editor rendering (Change-3)
- Retrieval/budget stats in StatusPage (follow-up PR)
- Mouse zone interaction (Change-3)

## Decisions

### D1: Page interface with Activate/Deactivate lifecycle

Pages implement `Activate() tea.Cmd` and `Deactivate()` beyond standard `tea.Model`. Cockpit root calls these on page switch. This solves StatusPage's tea.Tick start/stop without global tick management.

ChatModel (childModel) does NOT implement Page — it uses Init() from program start and runs continuously.

### D2: Focus ring via sidebarFocused flag

Tab toggles focus between sidebar and main content. When sidebar is focused, up/down/enter go to sidebar. When main is focused, all keys go to active page via consume-or-forward. This avoids Enter key conflicts between sidebar selection and chat submit.

### D3: OnSave intercepts "save" menu action, not StepComplete

The save trigger is `handleMenuSelection("save")` at editor.go:460. When `OnSave != nil`, call the callback and return to menu without `tea.Quit`. Standalone mode (`OnSave == nil`) keeps existing behavior.

### D4: Disabled sidebar items for unimplemented pages

MenuItem.Disabled=true renders items dim and skips them during cursor navigation. Sessions item is disabled in Change-2.

### D5: Inline save banner (not toast)

Cockpit has no toast system yet. Editor.View() renders success/error banner at menu top. Cleared on next key input. Simple and self-contained.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Editor save → tea.Quit kills cockpit | OnSave branch intercepts before tea.Quit. Standalone path unchanged. |
| Editor renders wider than cockpit panel | Allowed in Change-2 (clips). Width-aware rendering in Change-3. |
| StatusPage tick continues when page inactive | Deactivate() sets tickActive=false. Tick callback checks flag before scheduling next tick. |
| sessions sidebar Enter → route hole | Disabled items skip on Enter. Cursor navigation skips disabled items. |
| Save failure silent | Inline error banner in Editor.View(). Cleared on next input. |
