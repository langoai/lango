## ADDED Requirements

### Requirement: README describes Tasks and Approvals as degraded cockpit pages when dependencies are absent
The public README SHALL describe the cockpit Tasks and Approvals pages using the same degraded-surface contract as the runtime and cockpit feature docs.

#### Scenario: README cockpit shortcut table uses degraded-page wording for Tasks and Approvals
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Tasks row SHALL mention unavailable/degraded messaging when the background task manager is absent
- **AND** the Approvals row SHALL mention unavailable/degraded messaging when approval stores are absent
