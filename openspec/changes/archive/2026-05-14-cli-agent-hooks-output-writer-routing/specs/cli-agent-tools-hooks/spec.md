## MODIFIED Requirements

### Requirement: Agent hooks output routing
`lango agent hooks` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Agent hooks output uses the command writer
- **WHEN** `lango agent hooks` renders text or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
