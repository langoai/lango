## ADDED Requirements

### Requirement: README describes Dead Letters as a degraded cockpit page
The public README SHALL describe the cockpit Dead Letters page using the same degraded-page contract as the runtime and feature docs.

#### Scenario: README cockpit shortcut table uses degraded-page wording
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Dead Letters row SHALL describe the page as available with degraded unavailable messaging until the dead-letter bridge is ready
- **AND** it SHALL NOT imply that the page appears only when the bridge is available
