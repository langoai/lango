## MODIFIED Requirements

### Requirement: Root entrypoint failure paths remain seam-aware
The top-level `lango` entrypoint SHALL route broker-mode and root-command failure messages through the configured stderr seam and produce deterministic exit codes under test. Sandbox worker mode SHALL return the worker seam's exit code without evaluating broker mode or constructing the root command.

#### Scenario: Worker mode returns worker exit code
- **WHEN** sandbox worker mode is active
- **THEN** the entrypoint SHALL invoke the sandbox worker seam
- **AND** it SHALL return the worker seam's exit code
- **AND** it SHALL NOT evaluate broker mode or construct the root command
