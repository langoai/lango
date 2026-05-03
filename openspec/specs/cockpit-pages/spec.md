## Purpose

Capability spec for cockpit-pages. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Page interface with lifecycle
The cockpit SHALL define a `Page` interface extending `tea.Model` with `Title() string`, `ShortHelp() []key.Binding`, `Activate() tea.Cmd`, and `Deactivate()`.

#### Scenario: Page activation on switch
- **WHEN** cockpit switches from PageChat to PageStatus
- **THEN** cockpit SHALL call `StatusPage.Activate()` and execute the returned `tea.Cmd`

#### Scenario: Page deactivation on switch
- **WHEN** cockpit switches away from PageStatus
- **THEN** cockpit SHALL call `StatusPage.Deactivate()` before activating the new page

### Requirement: PageID routing

The cockpit SHALL define a `PageMissionControl` route alongside the existing cockpit pages and support round-trip routing between the page ID, sidebar entry, and registered page instance. This routing contract remains internal to cockpit; Wave 6 SHALL NOT use the cockpit page system as the bare-`lango` surface contract.

#### Scenario: Mission Control sidebar entry round-trips to page ID
- **WHEN** the sidebar emits the Mission Control item ID
- **THEN** `PageIDFromString(...)` SHALL resolve it to `PageMissionControl`
- **AND** `PageMissionControl.String()` SHALL return the same sidebar item ID

#### Scenario: Mission Control page activates through cockpit routing
- **WHEN** explicit cockpit switches to `PageMissionControl`
- **THEN** the registered Mission Control page SHALL receive `Activate()`

#### Scenario: Existing detail-page shortcuts remain stable
- **WHEN** the operator uses the existing cockpit global shortcuts for detail pages
- **THEN** the same detail pages as before SHALL still open
- **AND** Wave 6 SHALL NOT require shifting those shortcuts only to make the surface split happen

#### Scenario: Mission Control remains reachable inside cockpit
- **WHEN** the operator wants to return to Mission Control from another cockpit page
- **THEN** sidebar or page routing SHALL provide a direct path back to `PageMissionControl`

### Requirement: Focus ring between sidebar and content
Mission Control integration SHALL preserve the existing cockpit distinction between navigation focus and page-content focus. Adding the new page SHALL NOT collapse sidebar focus and content focus into a single routing model.

#### Scenario: Sidebar focus still routes navigation keys
- **WHEN** sidebar focus is active on a cockpit page
- **THEN** up/down/enter SHALL continue to route to the sidebar instead of the page content

#### Scenario: Mission Control content still receives keys when content-focused
- **WHEN** Mission Control is active and content focus is selected
- **THEN** page-local keys such as lane focus movement and composer typing SHALL route to the Mission Control page

### Requirement: Extended Deps struct
Cockpit Deps SHALL include `ToolCatalog`, `MetricsCollector`, `FeatureStatuses`, `ConfigStore`, and `ProfileName` in addition to existing fields.

#### Scenario: All Deps fields assignable from App
- **WHEN** cockpit.New(deps) is called
- **THEN** all fields SHALL be directly assignable from App struct public fields

