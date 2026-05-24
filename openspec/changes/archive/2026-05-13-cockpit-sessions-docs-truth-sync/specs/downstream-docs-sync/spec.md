## ADDED Requirements

### Requirement: Public cockpit docs describe Sessions page behavior
Public cockpit documentation SHALL describe the Sessions page using the current runtime contract.

#### Scenario: Cockpit feature page describes sessions ordering and unavailable-state split
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** the page SHALL describe the Sessions page as a newest-first session summary list
- **AND** it SHALL explain that the page distinguishes an unavailable session-list source from an empty configured session history
