## MODIFIED Requirements

### Requirement: Consume-or-forward message delegation
The cockpit `Update()` SHALL route messages based on `sidebarFocused` and `activePage`. When sidebar is focused, key events SHALL go to sidebar. When content is focused and activePage is PageChat, keys SHALL go to child via existing consume-or-forward. For non-chat pages, keys SHALL go to `pages[activePage].Update()`.

Cockpit SHALL additionally consume: `Ctrl+1` through `Ctrl+4` (page switch), `Tab` (focus toggle), `PageSelectedMsg` (sidebar selection).

#### Scenario: Ctrl+2 switches to settings
- **WHEN** cockpit receives Ctrl+2
- **THEN** activePage SHALL become PageSettings, sidebar active item SHALL update, and SettingsPage.Activate() SHALL be called

#### Scenario: Tab toggles focus to sidebar
- **WHEN** cockpit receives Tab with sidebarFocused=false
- **THEN** sidebarFocused SHALL become true and sidebar.SetFocused(true) SHALL be called

#### Scenario: PageSelectedMsg from sidebar
- **WHEN** cockpit receives PageSelectedMsg{ID: "tools"}
- **THEN** activePage SHALL switch to PageTools and sidebarFocused SHALL become false

### Requirement: Cockpit root model orchestrates 2-panel layout
View() SHALL dispatch to the active page: `child.View()` for PageChat, `pages[activePage].View()` for others. Layout composition (sidebar + main) remains unchanged.

#### Scenario: StatusPage active renders status view
- **WHEN** activePage is PageStatus
- **THEN** View() SHALL render sidebar + StatusPage.View()
