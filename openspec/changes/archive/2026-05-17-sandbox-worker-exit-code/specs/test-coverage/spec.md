## ADDED Requirements

### Requirement: Sandbox worker exit-code regressions stay executable

Executable tests SHALL cover sandbox worker exit-code behavior without intercepting `os.Exit`.

#### Scenario: Worker protocol exit-code paths are covered
- **WHEN** sandbox worker tests run
- **THEN** they SHALL exercise malformed input, unregistered tool, tool error, and successful tool paths
- **AND** they SHALL assert returned exit codes and JSON results without process-global exit interception
