## ADDED Requirements
### Requirement: Cockpit feature docs describe the Dead Letters operator surface
The public cockpit feature reference SHALL describe the current Dead Letters page beyond simple roster availability.

#### Scenario: Cockpit feature page includes Dead Letters section
- **WHEN** `docs/features/cockpit.md` documents cockpit pages
- **THEN** it SHALL include a dedicated Dead Letters section
- **AND** it SHALL describe filter controls, retry request flow, and degraded/unavailable or load-failure states
