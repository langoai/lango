## ADDED Requirements

### Requirement: Public cockpit docs describe Dead Letters as a degraded page
Public cockpit documentation SHALL describe Dead Letters as an always-registered cockpit page that degrades to unavailable messaging when its bridge is absent.

#### Scenario: Cockpit feature page describes degraded Dead Letters availability
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** it SHALL describe Dead Letters as always available in the cockpit roster
- **AND** it SHALL explain that the page surfaces unavailable/degraded messaging until the dead-letter bridge is ready
