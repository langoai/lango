## MODIFIED Requirements
### Requirement: Tools page degrades to an explicit empty state without a catalog
The cockpit Tools page SHALL render a stable empty state when no tool catalog is configured.

#### Scenario: Nil catalog renders explicit unavailable message
- **WHEN** ToolsPage is constructed with a nil tool catalog
- **THEN** category refresh and view rendering SHALL succeed without panic
- **AND** both panels SHALL explain that the tool catalog is not available

#### Scenario: Empty catalog renders explicit no-categories message
- **WHEN** ToolsPage is constructed with a configured catalog that has zero categories
- **THEN** both panes SHALL explain that no categories are registered

#### Scenario: Cockpit can register Tools route without catalog
- **WHEN** the cockpit starts with no tool catalog wired
- **THEN** the Tools page route SHALL still be registerable
