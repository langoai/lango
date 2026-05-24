## ADDED Requirements

### Requirement: Root entrypoint failure paths remain seam-aware
The top-level `lango` entrypoint SHALL route broker-mode and root-command failure messages through the configured stderr seam and produce deterministic exit codes under test.

#### Scenario: Broker-mode failure writes to injected stderr
- **WHEN** broker mode is active and the broker server returns an error
- **THEN** the error SHALL be written through the injected stderr seam
- **AND** the entrypoint SHALL return exit code `1`

#### Scenario: Root command failure writes to injected stderr
- **WHEN** the root Cobra command returns an error
- **THEN** the error SHALL be written through the injected stderr seam
- **AND** the entrypoint SHALL return exit code `1`
