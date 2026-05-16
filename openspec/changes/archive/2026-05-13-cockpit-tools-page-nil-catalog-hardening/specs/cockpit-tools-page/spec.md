## ADDED Requirements

### Requirement: Tools page degrades to an explicit empty state without a catalog
ToolsPage SHALL remain routable even when no `ToolCatalog` is available and SHALL render an explicit empty-state explanation instead of panicking or disappearing from cockpit routing.

#### Scenario: Nil catalog renders empty state
- **WHEN** ToolsPage is constructed with a nil tool catalog
- **THEN** category refresh and view rendering SHALL succeed without panic
- **AND** the page SHALL explain that the tool catalog is unavailable

#### Scenario: Cockpit can still register the tools page without a catalog
- **WHEN** the cockpit starts with no tool catalog wired
- **THEN** the Tools page route SHALL still be registerable
- **AND** activating that page SHALL render the empty-state contract instead of failing page routing
