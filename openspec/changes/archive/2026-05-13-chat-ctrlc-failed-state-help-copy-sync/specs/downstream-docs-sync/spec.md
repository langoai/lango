## MODIFIED Requirements
### Requirement: README and CLI overview describe cockpit core pages as degraded surfaces when dependencies are absent
The public README and CLI overview SHALL describe cockpit core pages using the current degraded-page routing contract.

#### Scenario: CLI overview describes Chat idle and failed quit semantics
- **WHEN** a user reads the `lango cockpit` overview in `docs/cli/core.md`
- **THEN** the Chat key-binding summary SHALL describe `Ctrl+C` as a double-press quit path for idle or failed states
