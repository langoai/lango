## MODIFIED Requirements

### Requirement: PageID routing includes Mission Control
The cockpit SHALL define a `PageMissionControl` route alongside the existing cockpit pages and support round-trip routing between the page ID, sidebar entry, and registered page instance.

#### Scenario: Mission Control sidebar entry round-trips to page ID
- **WHEN** the sidebar emits the Mission Control item ID
- **THEN** `PageIDFromString(...)` SHALL resolve it to `PageMissionControl`
- **AND** `PageMissionControl.String()` SHALL return the same sidebar item ID

#### Scenario: Mission Control page activates through routing
- **WHEN** cockpit switches to `PageMissionControl`
- **THEN** the registered Mission Control page SHALL receive `Activate()`

### Requirement: Default page migration minimizes shortcut churn
Mission Control SHALL become the default landing page without forcing a broad remap of the existing cockpit global shortcuts in Wave 1. Existing page shortcuts remain stable unless a specific route is intentionally reassigned.

#### Scenario: Existing detail-page shortcuts remain stable
- **WHEN** the operator uses the existing cockpit global shortcuts for detail pages
- **THEN** the same detail pages as before SHALL still open
- **AND** Wave 1 SHALL NOT require shifting every existing shortcut only to make room for Mission Control

#### Scenario: Mission Control remains reachable without shortcut remap
- **WHEN** the operator wants to return to Mission Control from another cockpit page
- **THEN** sidebar or page routing SHALL provide a direct path back to `PageMissionControl`

### Requirement: Focus model still separates navigation from page content
Mission Control integration SHALL preserve the existing cockpit distinction between navigation focus and page-content focus. Adding the new page SHALL NOT collapse sidebar focus and content focus into a single routing model.

#### Scenario: Sidebar focus still routes navigation keys
- **WHEN** sidebar focus is active on a cockpit page
- **THEN** up/down/enter SHALL continue to route to the sidebar instead of the page content

#### Scenario: Mission Control content still receives keys when content-focused
- **WHEN** Mission Control is active and content focus is selected
- **THEN** page-local keys such as lane focus movement and composer typing SHALL route to the Mission Control page
