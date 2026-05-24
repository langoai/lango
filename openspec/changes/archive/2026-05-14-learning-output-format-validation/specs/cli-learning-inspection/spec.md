## ADDED Requirements
### Requirement: Learning CLI output format stays explicit and validated
`lango learning status` and `lango learning history` SHALL accept `--output table|json` and reject unknown values before config or bootstrap loading.

#### Scenario: Learning status rejects unknown output before config load
- **WHEN** user runs `lango learning status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Learning history rejects unknown output before bootstrap
- **WHEN** user runs `lango learning history --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the bootstrap loader
