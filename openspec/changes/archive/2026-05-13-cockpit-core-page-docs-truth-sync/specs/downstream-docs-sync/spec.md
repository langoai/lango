## ADDED Requirements

### Requirement: README and CLI overview describe cockpit core pages as degraded surfaces when dependencies are absent
The public README and CLI overview SHALL describe cockpit core pages using the current degraded-page routing contract.

#### Scenario: README and CLI overview describe degraded Settings, Status, and Sessions pages
- **WHEN** a user reads the cockpit shortcut table in `README.md` or the `lango cockpit` overview in `docs/cli/core.md`
- **THEN** those docs SHALL explain that core cockpit pages remain routable and surface degraded/unavailable messaging when backing services are absent
- **AND** the README SHALL mention degraded-state behavior for Settings, Status, and Sessions specifically
