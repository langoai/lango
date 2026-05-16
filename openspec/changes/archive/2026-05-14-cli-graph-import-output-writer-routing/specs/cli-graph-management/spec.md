## MODIFIED Requirements

### Requirement: Graph import command output routing
`lango graph import <file> [--json]` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Graph import output uses the command writer
- **WHEN** `lango graph import` renders text or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
