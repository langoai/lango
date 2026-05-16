## MODIFIED Requirements

### Requirement: Graph export command output routing
`lango graph export [--format json|csv]` SHALL route JSON and CSV output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Graph export output uses the command writer
- **WHEN** `lango graph export` renders JSON or CSV output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
