## ADDED Requirements
### Requirement: Agent diagnostics output format stays explicit and validated
`lango agent trace list`, `lango agent trace show`, `lango agent graph`, and `lango agent trace metrics` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: Agent diagnostics reject unknown output before bootstrap
- **WHEN** the operator runs one of those commands with `--output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
