## MODIFIED Requirements

### Requirement: Performance fields in agent status
`lango agent status` SHALL display MaxTurns, ErrorCorrectionEnabled, and MaxDelegationRounds (multi-agent only) with their effective values (config or default).

#### Scenario: Default values displayed
- **WHEN** user runs `lango agent status` with no performance config
- **THEN** output SHALL show Max Turns: 25, Error Correction: true

#### Scenario: Multi-agent delegation rounds
- **WHEN** user runs `lango agent status` with `agent.multiAgent: true`
- **THEN** output SHALL include Delegation Rounds field

#### Scenario: JSON output includes new fields
- **WHEN** user runs `lango agent status --json`
- **THEN** JSON output SHALL include `max_turns`, `error_correction_enabled`, and `max_delegation_rounds` fields
