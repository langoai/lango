## ADDED Requirements
### Requirement: Additional P2P read-only inspection commands stay explicit and validated
`lango p2p discover`, `pricing`, `reputation`, and `session list` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: Discover rejects unknown output before bootstrap
- **WHEN** user runs `lango p2p discover --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

#### Scenario: Pricing rejects unknown output before bootstrap
- **WHEN** user runs `lango p2p pricing --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

#### Scenario: Reputation rejects unknown output before bootstrap
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123 --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

#### Scenario: Session list rejects unknown output before bootstrap
- **WHEN** user runs `lango p2p session list --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
