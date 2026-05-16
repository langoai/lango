## ADDED Requirements
### Requirement: P2P inspection output format stays explicit and validated
`lango p2p status`, `peers`, and `identity` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: P2P status rejects unknown output before bootstrap
- **WHEN** user runs `lango p2p status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

#### Scenario: P2P peers reject unknown output before bootstrap
- **WHEN** user runs `lango p2p peers --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

#### Scenario: P2P identity rejects unknown output before bootstrap
- **WHEN** user runs `lango p2p identity --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
