## ADDED Requirements

### Requirement: Public cockpit docs describe current page availability
Public cockpit documentation SHALL describe the current page roster and availability behavior instead of older assumptions about always-live or absent optional pages.

#### Scenario: Cockpit feature page describes disabled optional destinations
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** the page SHALL describe the current 9-item sidebar roster headed by Mission Control
- **AND** it SHALL explain that unavailable optional pages remain visible as disabled destinations
- **AND** it SHALL describe Dead Letters as disabled until its bridge is ready rather than as an always-live page
