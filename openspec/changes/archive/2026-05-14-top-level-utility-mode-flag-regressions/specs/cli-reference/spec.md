## MODIFIED Requirements

### Requirement: Top-level utility commands write success output through Cobra streams
Top-level utility commands under `lango` SHALL route successful human-readable output through the Cobra command output stream so wrappers and command-level tests can capture it without intercepting process-global stdout.

#### Scenario: Utility subcommands ignore the root mode flag
- **WHEN** `lango version` or `lango health` is executed with the root-level `--mode` flag
- **THEN** the utility subcommand SHALL still complete normally
