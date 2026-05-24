## MODIFIED Requirements

### Requirement: Root entrypoint failure paths remain seam-aware
The top-level `lango` entrypoint SHALL route broker-mode and root-command failure messages through the configured stderr seam and produce deterministic exit codes under test.

#### Scenario: Broker mode forwards injected stdin and stdout
- **WHEN** broker mode is active and the broker server starts successfully
- **THEN** the entrypoint SHALL pass the configured stdin and stdout seams into the broker server
