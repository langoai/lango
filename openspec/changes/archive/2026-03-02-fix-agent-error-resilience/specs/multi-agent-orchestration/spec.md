## ADDED Requirements

### Requirement: Multi-agent default turn limit
When `agent.multiAgent` is true and no explicit `MaxTurns` is configured, the system SHALL default to 50 turns instead of the standard 25. This provides sufficient headroom for multi-agent workflows with delegation overhead.

#### Scenario: Multi-agent mode with no explicit MaxTurns
- **WHEN** `agent.multiAgent` is true AND `agent.maxTurns` is zero or unset
- **THEN** the system SHALL use 50 as the maximum turn limit

#### Scenario: Multi-agent mode with explicit MaxTurns
- **WHEN** `agent.multiAgent` is true AND `agent.maxTurns` is set to a positive value
- **THEN** the system SHALL use the explicitly configured value, not the multi-agent default

#### Scenario: Single-agent mode unaffected
- **WHEN** `agent.multiAgent` is false
- **THEN** the system SHALL use the standard default of 25 turns (unchanged behavior)
