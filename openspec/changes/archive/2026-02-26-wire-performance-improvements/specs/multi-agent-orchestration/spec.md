## MODIFIED Requirements

### Requirement: Configurable delegation rounds
The orchestrator SHALL use `cfg.Agent.MaxDelegationRounds` instead of hardcoded `5`. When the config value is 0, the orchestrator default (10) SHALL be used.

#### Scenario: Custom delegation rounds from config
- **WHEN** config sets `agent.maxDelegationRounds: 8`
- **THEN** the orchestrator SHALL limit delegation to 8 rounds per turn

#### Scenario: Default delegation rounds
- **WHEN** config omits `agent.maxDelegationRounds`
- **THEN** the orchestrator SHALL use its default of 10 rounds
