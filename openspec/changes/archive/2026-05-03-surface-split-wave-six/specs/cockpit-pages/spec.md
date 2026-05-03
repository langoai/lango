## MODIFIED Requirements

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
