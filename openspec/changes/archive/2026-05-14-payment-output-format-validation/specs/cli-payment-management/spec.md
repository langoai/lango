## ADDED Requirements
### Requirement: Payment CLI output format stays explicit and validated
`lango payment balance`, `history`, `limits`, `info`, `send`, and `x402` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: Payment balance rejects unknown output before bootstrap
- **WHEN** user runs `lango payment balance --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

#### Scenario: Payment send rejects unknown output before bootstrap
- **WHEN** user runs `lango payment send --to 0x... --amount 1.00 --purpose "test" --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
