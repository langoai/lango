## ADDED Requirements

### Requirement: Public cockpit docs describe the Mission Control composer with neutral request wording
Public cockpit documentation SHALL describe the default first-screen Mission Control composer hint using neutral request wording rather than chat-only wording.

#### Scenario: Cockpit feature page uses request wording
- **WHEN** a user reads the Mission Control composer description in `docs/features/cockpit.md`
- **THEN** it SHALL describe the hint as `Type a request here, or use lango chat for focused chat`
- **AND** it SHALL NOT describe the default first-screen hint as `Type to chat here`
