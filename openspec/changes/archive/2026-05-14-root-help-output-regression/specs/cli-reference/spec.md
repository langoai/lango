## MODIFIED Requirements

### Requirement: Root entrypoint failure paths remain seam-aware
The top-level `lango` entrypoint SHALL route broker-mode and root-command failure messages through the configured stderr seam and produce deterministic exit codes under test.

#### Scenario: Non-interactive bare root writes help to command output
- **WHEN** bare `lango` is executed without arguments in a non-interactive environment
- **THEN** the root help text SHALL be written through the Cobra command output stream
