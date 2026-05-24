## MODIFIED Requirements
### Requirement: Dead Letters remains registered as a cockpit degraded surface
The cockpit SHALL keep Dead Letters available as a registered page route even when the dead-letter bridge is unavailable, relying on page-level unavailable messaging instead of suppressing page registration.

#### Scenario: Dead Letters page remains registered without bridge callbacks
- **WHEN** cockpit startup wiring runs without a ready dead-letter bridge
- **THEN** the Dead Letters page route SHALL still be registered

#### Scenario: Dead Letters degraded page reports missing list function immediately
- **WHEN** the registered Dead Letters page renders without a configured list callback
- **THEN** the page SHALL explain that the dead-letter backlog is not configured

#### Scenario: Dead Letters activation still surfaces missing callback error
- **WHEN** the registered Dead Letters page is activated without a configured list callback
- **THEN** activation SHALL yield a load error that explains the dead-letter list function is not configured

#### Scenario: Dead Letters backlog load failure renders explicit failure message
- **WHEN** the registered Dead Letters page renders after a configured backlog load failed
- **THEN** the page SHALL explain that dead letters failed to load
- **AND** it SHALL include the underlying error text

#### Scenario: Dead Letters detail load failure renders explicit failure message
- **WHEN** the registered Dead Letters page renders with a selected row whose detail query failed
- **THEN** the detail pane SHALL explain that detail failed to load
- **AND** it SHALL include the underlying error text
