## ADDED Requirements

### Requirement: Page interface with lifecycle
The cockpit SHALL define a `Page` interface extending `tea.Model` with `Title() string`, `ShortHelp() []key.Binding`, `Activate() tea.Cmd`, and `Deactivate()`.

#### Scenario: Page activation on switch
- **WHEN** cockpit switches from PageChat to PageStatus
- **THEN** cockpit SHALL call `StatusPage.Activate()` and execute the returned `tea.Cmd`

#### Scenario: Page deactivation on switch
- **WHEN** cockpit switches away from PageStatus
- **THEN** cockpit SHALL call `StatusPage.Deactivate()` before activating the new page

### Requirement: PageID routing
The cockpit SHALL define PageID constants: `PageChat`, `PageSettings`, `PageTools`, `PageStatus`. Ctrl+1 through Ctrl+4 SHALL switch to the corresponding page.

#### Scenario: Ctrl+3 switches to tools
- **WHEN** user presses Ctrl+3
- **THEN** cockpit SHALL set activePage to PageTools and call ToolsPage.Activate()

### Requirement: Focus ring between sidebar and content
Tab SHALL toggle `sidebarFocused` between true and false. When sidebar is focused, up/down/enter SHALL be routed to sidebar. When content is focused, keys SHALL be routed to the active page.

#### Scenario: Tab toggles focus
- **WHEN** user presses Tab with sidebarFocused=false
- **THEN** sidebarFocused SHALL become true and sidebar SHALL receive subsequent key events

### Requirement: Extended Deps struct
Cockpit Deps SHALL include `ToolCatalog`, `MetricsCollector`, `FeatureStatuses`, `ConfigStore`, and `ProfileName` in addition to existing fields.

#### Scenario: All Deps fields assignable from App
- **WHEN** cockpit.New(deps) is called
- **THEN** all fields SHALL be directly assignable from App struct public fields
