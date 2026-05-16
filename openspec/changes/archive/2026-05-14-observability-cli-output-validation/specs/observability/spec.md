## ADDED Requirements
### Requirement: Metrics output format values are validated before gateway calls
`lango metrics` and its subcommands SHALL accept only `table` or `json` for `--output`. Unknown values SHALL fail before the command contacts the gateway.

#### Scenario: Metrics summary rejects an unknown output format before fetch
- **WHEN** the operator runs `lango metrics --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT contact the gateway
