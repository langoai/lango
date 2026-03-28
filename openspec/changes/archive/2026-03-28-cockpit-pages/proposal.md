## Why

The cockpit TUI (Change-1) currently has only a single Chat page with a non-interactive sidebar. Users cannot access settings, tool catalog, or system status from within the cockpit. Adding pages makes the cockpit a practical multi-function terminal — not just a chat wrapper with a sidebar.

## What Changes

- Add Page interface with Activate/Deactivate lifecycle for page-level resource management
- Add page router with Ctrl+1-4 direct switching and sidebar-driven navigation
- Make sidebar interactive: focus ring (Tab toggle), up/down/enter navigation, disabled items
- Add SettingsPage embedding `settings.Editor` with OnSave callback (no tea.Quit in embedded mode)
- Add StatusPage showing feature flags, token metrics, tool execution stats (auto-refreshing)
- Add ToolsPage browsing tool catalog by category
- Extend cockpit Deps with ToolCatalog, MetricsCollector, FeatureStatuses, ConfigStore
- Add save result inline banner to Editor (success/error feedback in embedded mode)
- Mark sessions sidebar item as disabled (SessionsPage deferred to Change-3)

## Capabilities

### New Capabilities
- `cockpit-pages`: Page interface (Activate/Deactivate lifecycle), PageID routing, cockpit core multi-page orchestration, Deps extension, focus ring between sidebar and content
- `cockpit-tools-page`: Tool catalog browser with category navigation and tool detail view
- `cockpit-status-page`: Feature status dashboard with auto-refreshing metrics via tea.Tick
- `cockpit-settings-page`: Embedded settings editor with OnSave callback, inline save banner

### Modified Capabilities
- `cockpit-sidebar`: Interactive navigation (cursor, focused state, PageSelectedMsg, disabled items)
- `cockpit-shell`: Extended Update() with page routing (Ctrl+1-4, Tab focus), View() dispatches to active page

## Impact

- **Modified**: `internal/cli/settings/editor.go` — OnSave field, saveSuccess field, inline banner, NewEditorForEmbedding constructor
- **Modified**: `internal/cli/cockpit/cockpit.go` — pages map, activePage, sidebarFocused, page routing
- **Modified**: `internal/cli/cockpit/deps.go` — 5 new fields
- **Modified**: `internal/cli/cockpit/keymap.go` — Ctrl+1-4, Tab bindings
- **Modified**: `internal/cli/cockpit/sidebar/sidebar.go` — cursor, focused, Disabled, PageSelectedMsg
- **New**: `internal/cli/cockpit/router.go`, `internal/cli/cockpit/pages/tools.go`, `pages/status.go`, `pages/settings.go`
- **Modified**: `cmd/lango/main.go` — runCockpit() Deps wiring expansion
