## ADDED Requirements
### Requirement: Cockpit feature docs describe the Chat operator surface
The public cockpit feature reference SHALL describe the current Chat page beyond simple roster availability.

#### Scenario: Cockpit feature page includes Chat section
- **WHEN** `docs/features/cockpit.md` documents cockpit pages
- **THEN** it SHALL include a dedicated Chat section
- **AND** it SHALL describe the primary send/newline/quit key surface, slash-command discoverability, and inline approval-interrupt controls
