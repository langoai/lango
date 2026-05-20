## Purpose

Capability spec for cockpit-pages. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Page interface with lifecycle
The cockpit SHALL define a `Page` interface extending `tea.Model` with `Title() string`, `ShortHelp() []key.Binding`, `Activate() tea.Cmd`, and `Deactivate()`.

#### Scenario: Sessions and Tools vertical navigation hints use cockpit-standard labels
- **WHEN** the Sessions or Tools page help is rendered
- **THEN** the vertical navigation bindings SHALL use `↑/k` and `↓/j` help labels rather than textual `up/k` or `down/j` forms

### Requirement: PageID routing

The cockpit SHALL define a `PageMissionControl` route alongside the existing cockpit pages and support round-trip routing between the page ID, sidebar entry, and registered page instance. This routing contract remains internal to cockpit; Slice 6 SHALL NOT use the cockpit page system as the bare-`lango` surface contract.

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
- **AND** Slice 6 SHALL NOT require shifting those shortcuts only to make the surface split happen

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

### Requirement: Status and Settings remain registered as cockpit core pages
The cockpit SHALL keep Status and Settings available as core page routes even when their optional backing dependencies are absent, relying on the page-level degraded states instead of suppressing page registration.

#### Scenario: Status page remains registered without metrics or feature providers
- **WHEN** cockpit startup wiring runs with no metrics collector and no feature-status provider
- **THEN** the Status page route SHALL still be registered

#### Scenario: Settings page remains registered without a config-profile store
- **WHEN** cockpit startup wiring runs with no config-profile store
- **THEN** the Settings page route SHALL still be registered

### Requirement: Dead Letters remains registered as a cockpit degraded surface
The cockpit SHALL keep Dead Letters available as a registered page route even when the dead-letter bridge is unavailable, relying on page-level unavailable messaging instead of suppressing page registration.

#### Scenario: Retry-running state hides inert reset help
- **WHEN** the Dead Letters help is rendered while a retry request is actively running
- **THEN** it SHALL omit the `Ctrl+R` reset binding

### Requirement: Mission Control surfaces unavailable projector state explicitly
The cockpit Mission Control page SHALL distinguish an unavailable mission-control projector from a normal empty mission state.

#### Scenario: Nil projector renders degraded note
- **WHEN** Mission Control renders after activation with no configured projector
- **THEN** the page SHALL still render its empty-state shell
- **AND** it SHALL include a degraded note explaining that Mission Control data is not configured

#### Scenario: Empty Mission Control help hides inert navigation keys
- **WHEN** Mission Control is in its true empty state with no missions, no pending decision, and no loops
- **THEN** the page help SHALL omit `↑/k` and `↓/j`
- **AND** it SHALL still advertise the focus and remaining actionable keys

#### Scenario: Empty workbench help labels starter seeding explicitly
- **WHEN** the standalone workbench surface is in the true empty state
- **AND** the default starter prompt is not yet staged
- **THEN** the `Enter` help label SHALL describe seeding the starter prompt

#### Scenario: Empty workbench help labels starter submission explicitly
- **WHEN** the standalone workbench surface is in the true empty state
- **AND** a starter prompt is already staged in the composer
- **THEN** the `Enter` help label SHALL describe running the starter prompt

#### Scenario: Populated Mission Control help hides inert navigation in single-row focus states
- **WHEN** Mission Control has content but the currently focused lane has fewer than two navigable rows
- **THEN** the page help SHALL omit `↑/k` and `↓/j`

#### Scenario: Composer/activity focus routes vertical navigation to the activity lane
- **WHEN** Mission Control focus is on the composer/activity lane
- **AND** multiple activity rows exist
- **THEN** pressing `↑/k` or `↓/j` SHALL move the activity cursor instead of being swallowed by the composer

### Requirement: Mission Control top-level submit fails closed without mission service
The cockpit Mission Control page SHALL preserve its durable-first submit contract by failing closed when no mission service is configured for an ordinary top-level composer submit.

#### Scenario: Ordinary submit returns explicit system message without mission service
- **WHEN** the Mission Control composer contains a non-slash request
- **AND** no mission service is configured
- **THEN** submit handling SHALL return an explicit system message explaining that Mission Control cannot start a durable mission
- **AND** it SHALL NOT fall through to ordinary shared-chat execution

#### Scenario: Slash submit still passes through without mission service
- **WHEN** the Mission Control composer contains a slash command
- **AND** no mission service is configured
- **THEN** the command SHALL still pass through to the shared chat path

### Requirement: Mission Control proposed-mission actions fail closed when backing services are absent
The cockpit Mission Control page SHALL return explicit system messages when the operator triggers proposed-mission actions that require unavailable backing services.

#### Scenario: Proposal row help exposes accept and dismiss keys
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** the missions lane is focused
- **THEN** the help bar SHALL label `Enter` as accepting the proposal
- **AND** it SHALL include the `d` dismiss binding

#### Scenario: Proposal row help preserves navigation when another row exists
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** the missions lane is focused
- **AND** another mission row exists
- **THEN** the help bar SHALL continue exposing the `↑/↓` navigation bindings
- **AND** SHALL keep the `Enter` accept and `d` dismiss bindings visible at the same time

#### Scenario: Dismiss help hides while another lane is focused
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** focus is not on the missions lane
- **THEN** the help bar SHALL omit the `d` dismiss binding

#### Scenario: Ordinary mission row hides inert Enter help
- **WHEN** the focused Mission Control lane has an ordinary mission row selected
- **AND** no starter-flow or composer submit path is active
- **THEN** the help bar SHALL omit the generic `Enter` binding

#### Scenario: Decisions lane help hides inert Enter key
- **WHEN** Mission Control focus is on the decisions lane
- **THEN** the help bar SHALL omit the generic `Enter` submit binding

#### Scenario: True empty cockpit hides inert Enter help
- **WHEN** Mission Control is in the true empty state on the ordinary cockpit surface
- **AND** no starter-flow path is available
- **THEN** the help bar SHALL omit the generic `Enter` binding

#### Scenario: Accept proposed mission fails closed without mission service
- **WHEN** the operator accepts a selected proposed mission
- **AND** no mission service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot accept the proposal into a durable mission row

#### Scenario: Dismiss proposed mission fails closed without proposal service
- **WHEN** the operator dismisses a selected proposed mission
- **AND** no proposal service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot dismiss the proposal
