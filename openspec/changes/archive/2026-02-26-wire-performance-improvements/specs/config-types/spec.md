## MODIFIED Requirements

### Requirement: AgentConfig fields
`AgentConfig` SHALL include `MaxTurns int`, `ErrorCorrectionEnabled *bool`, and `MaxDelegationRounds int` fields with mapstructure/json tags.

#### Scenario: Zero-value defaults
- **WHEN** config omits `maxTurns`, `errorCorrectionEnabled`, and `maxDelegationRounds`
- **THEN** the zero values (0, nil, 0) SHALL be interpreted as defaults (25, true, 10) by the wiring layer

### Requirement: ObservationalMemoryConfig fields
`ObservationalMemoryConfig` SHALL include `MemoryTokenBudget int` and `ReflectionConsolidationThreshold int` fields with mapstructure/json tags.

#### Scenario: Zero-value defaults
- **WHEN** config omits `memoryTokenBudget` and `reflectionConsolidationThreshold`
- **THEN** the zero values SHALL be interpreted as defaults (4000, 5) by the wiring layer
