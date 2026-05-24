## MODIFIED Requirements
### Requirement: Public cockpit docs describe the Tools page as a degraded surface
Public cockpit documentation SHALL describe the Tools page using the same degraded-surface contract as the runtime.

#### Scenario: README describes the empty-catalog Tools state
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Tools row SHALL mention that a configured catalog with zero categories surfaces an explicit no-categories state
