## ADDED Requirements
### Requirement: Status output format values are validated before work starts
`lango status` and its dead-letter subcommands SHALL accept only `table` or `json` for `--output`. Unknown values SHALL fail before bootstrap or dead-letter bridge work begins.

#### Scenario: Root status rejects an unknown output format before bootstrap
- **WHEN** the operator runs `lango status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the bootstrap loader

#### Scenario: Dead-letter status rejects an unknown output format before bridge loading
- **WHEN** the operator runs `lango status dead-letters --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the dead-letter bridge loader
