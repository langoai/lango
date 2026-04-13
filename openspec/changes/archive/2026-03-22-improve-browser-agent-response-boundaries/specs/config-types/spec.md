## MODIFIED Requirements

### Requirement: AgentConfig fields
`AgentConfig` SHALL include `MaxTurns int`, `ErrorCorrectionEnabled *bool`, and `MaxDelegationRounds int` fields with mapstructure/json tags.

#### Scenario: Zero-value defaults
- **WHEN** config omits `maxTurns`, `errorCorrectionEnabled`, and `maxDelegationRounds`
- **THEN** the zero values SHALL be interpreted as defaults by the wiring layer
- **AND** the effective defaults SHALL be 50 turns in single-agent mode, 75 turns in multi-agent mode, true for error correction, and 10 for max delegation rounds
