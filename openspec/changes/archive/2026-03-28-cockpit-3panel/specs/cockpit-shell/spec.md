## MODIFIED Requirements

### Requirement: Consume-or-forward message delegation
The cockpit `Update()` SHALL route messages based on `sidebarFocused` and `activePage`. When sidebar is focused, key events SHALL go to sidebar. When content is focused and activePage is PageChat, keys SHALL go to child via existing consume-or-forward. For non-chat pages, keys SHALL go to `pages[activePage].Update()`.

Cockpit SHALL additionally consume: `Ctrl+1` through `Ctrl+4` (page switch), `Tab` (focus toggle), `PageSelectedMsg` (sidebar selection), `Ctrl+P` (context panel toggle), `Ctrl+Y` (clipboard copy).

#### Scenario: Ctrl+2 switches to settings
- **WHEN** cockpit receives Ctrl+2
- **THEN** activePage SHALL become PageSettings, sidebar active item SHALL update, and SettingsPage.Activate() SHALL be called

#### Scenario: Tab toggles focus to sidebar
- **WHEN** cockpit receives Tab with sidebarFocused=false
- **THEN** sidebarFocused SHALL become true and sidebar.SetFocused(true) SHALL be called

#### Scenario: PageSelectedMsg from sidebar
- **WHEN** cockpit receives PageSelectedMsg{ID: "tools"}
- **THEN** activePage SHALL switch to PageTools and sidebarFocused SHALL become false

### Requirement: View dispatches to active page
View() SHALL dispatch to the active page: `child.View()` for PageChat, `pages[activePage].View()` for others. Layout composition (sidebar + main) remains unchanged.

#### Scenario: StatusPage active renders status view
- **WHEN** activePage is PageStatus
- **THEN** View() SHALL render sidebar + StatusPage.View()

### Requirement: 3-panel layout with context panel
When the context panel is visible, View() SHALL compose up to 3 panels: sidebar (left, optional), main content (center), context panel (right, optional). All components SHALL receive correct effective widths accounting for both sidebar and context panel.

#### Scenario: All three panels visible
- **WHEN** sidebarVisible=true and contextVisible=true and terminal width 120
- **THEN** child SHALL receive WindowSizeMsg{Width: 120 - 20 - 28 = 72, Height: terminalHeight}

#### Scenario: Only context panel visible
- **WHEN** sidebarVisible=false and contextVisible=true and terminal width 120
- **THEN** child SHALL receive WindowSizeMsg{Width: 120 - 28 = 92, Height: terminalHeight}

### Requirement: Mouse routing to sidebar
The cockpit SHALL route tea.MouseMsg to the sidebar when the click X coordinate is within the sidebar width. Mouse events SHALL be processed regardless of sidebar focus state.

#### Scenario: Click in sidebar region
- **WHEN** mouse click occurs at X < sidebarWidth
- **THEN** the event SHALL be forwarded to sidebar.Update(), not to child

#### Scenario: Click in content region
- **WHEN** mouse click occurs at X >= sidebarWidth
- **THEN** the event SHALL be forwarded to the active page or child

### Requirement: Clipboard copy
The cockpit SHALL support Ctrl+Y to copy the current active view content to the system clipboard.

#### Scenario: Copy chat view
- **WHEN** user presses Ctrl+Y with PageChat active
- **THEN** child.View() content SHALL be written to the system clipboard

#### Scenario: Copy other page view
- **WHEN** user presses Ctrl+Y with a non-chat page active
- **THEN** pages[activePage].View() content SHALL be written to the system clipboard

### Requirement: Context panel toggle with synthetic resize
When Ctrl+P toggles the context panel, cockpit SHALL send synthetic WindowSizeMsg to all components (child, all pages, contextPanel) with updated effective widths.

#### Scenario: Toggle context panel on
- **WHEN** user presses Ctrl+P to show context panel
- **THEN** contextPanel.Start() SHALL be called and all components SHALL receive reduced width

#### Scenario: Toggle context panel off
- **WHEN** user presses Ctrl+P to hide context panel
- **THEN** contextPanel.Stop() SHALL be called and all components SHALL receive increased width
