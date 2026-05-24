## MODIFIED Requirements

### Requirement: Graph add command output routing
`lango graph add --subject <s> --predicate <p> --object <o> [--json]` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Graph add output uses the command writer
- **WHEN** `lango graph add` renders text or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
