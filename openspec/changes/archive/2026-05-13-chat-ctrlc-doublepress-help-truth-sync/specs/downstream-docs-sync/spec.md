## MODIFIED Requirements
### Requirement: Cockpit feature docs describe the Chat operator surface
The public cockpit feature reference SHALL describe the current Chat page beyond simple roster availability.

#### Scenario: Cockpit feature page describes Chat double-press quit behavior
- **WHEN** a user reads the Chat section in `docs/features/cockpit.md`
- **THEN** it SHALL explain that idle-state `Ctrl+C` uses a double-press quit path
- **AND** it SHALL distinguish that `Ctrl+D` remains the immediate quit path

### Requirement: README and CLI overview describe cockpit core pages as degraded surfaces when dependencies are absent
The public README and CLI overview SHALL describe cockpit core pages using the current degraded-page routing contract.

#### Scenario: CLI overview describes Chat double-press quit behavior
- **WHEN** a user reads the `lango cockpit` overview in `docs/cli/core.md`
- **THEN** the Chat key-binding summary SHALL describe `Ctrl+C` as cancel-or-double-press-quit rather than a generic quit key
