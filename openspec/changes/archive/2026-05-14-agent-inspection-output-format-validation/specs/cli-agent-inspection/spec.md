## ADDED Requirements
### Requirement: Agent inspection output format stays explicit and validated
`lango agent status`, `lango agent list`, `lango agent tools`, and `lango agent hooks` SHALL accept `--output table|json` and reject unknown values before config loading.

#### Scenario: Agent inspection commands reject unknown output before config load
- **WHEN** the operator runs one of those commands with `--output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader
