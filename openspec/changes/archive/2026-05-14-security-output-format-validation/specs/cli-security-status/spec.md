## ADDED Requirements
### Requirement: Security inspection output format stays explicit and validated
`lango security status`, `lango security keyring status`, `lango security kms status`, `lango security kms keys`, and `lango security secrets list` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: Security status rejects unknown output before bootstrap
- **WHEN** `lango security status --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

#### Scenario: Security secrets list rejects unknown output before bootstrap
- **WHEN** `lango security secrets list --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
