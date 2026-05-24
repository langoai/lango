## ADDED Requirements
### Requirement: Cockpit feature docs describe the Settings and Status operator surfaces
The public cockpit feature reference SHALL describe the current Settings and Status pages beyond simple roster availability.

#### Scenario: Cockpit feature page includes Settings and Status sections
- **WHEN** `docs/features/cockpit.md` documents cockpit pages
- **THEN** it SHALL include dedicated Settings and Status sections
- **AND** it SHALL describe save-unavailable behavior for Settings
- **AND** it SHALL describe unavailable provider/collector states for Status
