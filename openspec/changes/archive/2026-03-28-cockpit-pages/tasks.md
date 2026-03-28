## 1. Settings Embedded Mode

- [x] 1.1 Add `OnSave func(cfg *config.Config, explicitKeys map[string]bool) error` and `saveSuccess bool` fields to `Editor` struct (`internal/cli/settings/editor.go`)
- [x] 1.2 Add `NewEditorForEmbedding(cfg *config.Config, onSave OnSaveFunc) *Editor` constructor — starts at StepMenu, sets OnSave
- [x] 1.3 Modify `handleMenuSelection("save")` to branch: OnSave non-nil → call callback + set saveSuccess, return nil; OnSave nil → existing tea.Quit path
- [x] 1.4 Add inline save banner to `View()` at StepMenu: error banner (red) if `e.err != nil`, success banner (green) if `e.saveSuccess`. Clear both on next key input.
- [x] 1.5 Tests for embedded save flow and standalone save unchanged

## 2. Cockpit Page Router

- [x] 2.1 Create `internal/cli/cockpit/router.go` — `PageID` enum (PageChat/PageSettings/PageTools/PageStatus), `Page` interface (tea.Model + Title + ShortHelp + Activate + Deactivate)

## 3. Cockpit Core Extension

- [x] 3.1 Extend `Deps` with ToolCatalog, MetricsCollector, FeatureStatuses, ConfigStore, ProfileName (`deps.go`)
- [x] 3.2 Extend `keyMap` with Ctrl+1-4 (page switch), Tab (focus toggle) (`keymap.go`)
- [x] 3.3 Add `pages map[PageID]Page`, `activePage PageID`, `sidebarFocused bool` to Model. Extend `New()` to create pages. Extend `Update()` with page routing + focus ring + PageSelectedMsg handling + Activate/Deactivate lifecycle. Extend `View()` to dispatch to active page (`cockpit.go`)
- [x] 3.4 Tests for page routing, focus toggle, Activate/Deactivate calls

## 4. Sidebar Interactive

- [x] 4.1 Add `cursor int`, `focused bool`, `Disabled bool` to MenuItem. Add `SetFocused(bool)`. Mark sessions as `Disabled: true` (`sidebar.go`)
- [x] 4.2 Implement focused `Update()`: up/down move cursor (skip disabled), Enter emits `PageSelectedMsg`. Unfocused Update returns unchanged.
- [x] 4.3 Update `View()`: show cursor indicator when focused, dim disabled items
- [x] 4.4 Tests for cursor navigation, disabled skip, PageSelectedMsg, focus states

## 5. ToolsPage

- [x] 5.1 Create `internal/cli/cockpit/pages/tools.go` — category list with cursor, tool detail panel, Page interface (Activate/Deactivate no-op)
- [x] 5.2 Tests for category listing, tool display

## 6. StatusPage

- [x] 6.1 Create `internal/cli/cockpit/pages/status.go` — feature flags, token usage, tool stats, uptime, provider/model info
- [x] 6.2 Implement Activate() → first Snapshot + tea.Tick start. Deactivate() → tickActive=false
- [x] 6.3 Tests for Snapshot rendering, Activate/Deactivate tick lifecycle

## 7. SettingsPage

- [x] 7.1 Create `internal/cli/cockpit/pages/settings.go` — embeds Editor via NewEditorForEmbedding, OnSave callback wires ConfigStore.Save
- [x] 7.2 Page interface: Activate/Deactivate no-op, Title="Settings"
- [x] 7.3 Tests for embedded save callback, Page interface compliance

## 8. main.go Wiring

- [x] 8.1 Update `runCockpit()` in `cmd/lango/main.go` — pass ToolCatalog, MetricsCollector, FeatureStatuses, ConfigStore, ProfileName to cockpit.Deps
- [x] 8.2 Verify `go build ./...` && `go test ./...` && `go vet ./...`
