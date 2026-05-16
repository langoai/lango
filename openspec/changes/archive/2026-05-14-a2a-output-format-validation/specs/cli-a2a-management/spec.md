## ADDED Requirements
### Requirement: A2A CLI output format stays explicit and validated
`lango a2a card` and `lango a2a check` SHALL accept `--output table|json` and reject unknown values before config loading or outbound fetch work.

#### Scenario: A2A card rejects unknown output before config load
- **WHEN** user runs `lango a2a card --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: A2A check rejects unknown output before fetch
- **WHEN** user runs `lango a2a check <url> --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT issue the outbound fetch request
