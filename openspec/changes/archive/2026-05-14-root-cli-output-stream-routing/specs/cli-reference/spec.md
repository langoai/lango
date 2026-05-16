## ADDED Requirements

### Requirement: Top-level utility commands write success output through Cobra streams
Top-level utility commands under `lango` SHALL route successful human-readable output through the Cobra command output stream so wrappers and command-level tests can capture it without intercepting process-global stdout.

#### Scenario: Version command writes to command output
- **WHEN** `lango version` succeeds
- **THEN** the version string SHALL be written through the Cobra command output stream
