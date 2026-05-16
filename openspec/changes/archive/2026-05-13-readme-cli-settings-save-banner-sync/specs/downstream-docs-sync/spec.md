## MODIFIED Requirements
### Requirement: README and CLI overview describe cockpit core pages as degraded surfaces when dependencies are absent
The public README and CLI overview SHALL describe cockpit core pages using the current degraded-page routing contract.

#### Scenario: README and CLI overview describe Settings save feedback
- **WHEN** a user reads the cockpit shortcut table in `README.md` or the `lango cockpit` overview in `docs/cli/core.md`
- **THEN** the Settings description SHALL mention that embedded saves surface inline success or failure feedback
