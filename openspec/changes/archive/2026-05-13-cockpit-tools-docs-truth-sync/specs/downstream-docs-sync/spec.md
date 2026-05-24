## ADDED Requirements

### Requirement: Public cockpit docs describe the Tools page as a degraded surface
Public cockpit documentation SHALL describe the Tools page using the same degraded-surface contract as the runtime.

#### Scenario: Cockpit feature page and README describe the Tools page degraded state
- **WHEN** a user reads `docs/features/cockpit.md` or the cockpit shortcut table in `README.md`
- **THEN** those docs SHALL describe the Tools page as always available
- **AND** they SHALL explain that the page surfaces unavailable/degraded messaging when the tool catalog is absent
