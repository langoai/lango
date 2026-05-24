## MODIFIED Requirements
### Requirement: README and CLI overview describe cockpit core pages as degraded surfaces when dependencies are absent
The public README and CLI overview SHALL describe cockpit core pages using the current degraded-page routing contract.

#### Scenario: README and CLI overview describe Settings and Status interaction semantics
- **WHEN** a user reads the cockpit shortcut table in `README.md` or the `lango cockpit` overview in `docs/cli/core.md`
- **THEN** the Settings description SHALL mention the embedded inline help/footer model
- **AND** the Status description SHALL mention that the page is read-only and refreshes automatically while active
