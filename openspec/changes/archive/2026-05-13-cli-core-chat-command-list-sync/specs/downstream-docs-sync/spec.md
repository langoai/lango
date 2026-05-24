## MODIFIED Requirements
### Requirement: README and CLI overview describe cockpit core pages as degraded surfaces when dependencies are absent
The public README and CLI overview SHALL describe cockpit core pages using the current degraded-page routing contract.

#### Scenario: CLI overview lists the current built-in chat slash commands
- **WHEN** a user reads the `lango cockpit` overview in `docs/cli/core.md`
- **THEN** the slash-command summary SHALL include `/mode` and `/cost` alongside the existing built-in chat commands
