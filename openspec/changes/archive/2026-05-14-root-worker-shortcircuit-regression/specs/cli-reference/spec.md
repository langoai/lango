## MODIFIED Requirements

### Requirement: Root entrypoint failure paths remain seam-aware
The top-level `lango` entrypoint SHALL route broker-mode and root-command failure messages through the configured stderr seam and produce deterministic exit codes under test.

#### Scenario: Worker mode short-circuits before broker and root command setup
- **WHEN** sandbox worker mode is active
- **THEN** the entrypoint SHALL invoke the sandbox worker seam and return success without evaluating broker mode or constructing the root command
