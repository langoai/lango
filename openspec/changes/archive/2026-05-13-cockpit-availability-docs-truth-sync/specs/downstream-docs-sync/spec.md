## MODIFIED Requirements

### Requirement: Public cockpit docs describe current page availability
Public cockpit documentation SHALL describe the current page roster and runtime availability behavior instead of older assumptions about always-live or disabled optional destinations.

#### Scenario: Cockpit feature page describes current degraded-page routing
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** the page SHALL describe the current 9-item sidebar roster headed by Mission Control
- **AND** it SHALL explain that cockpit pages remain routable and surface degraded in-page messaging when backing services are unavailable
- **AND** it SHALL describe Dead Letters as an always-registered degraded page rather than as a disabled destination
