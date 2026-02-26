## MODIFIED Requirements

### Requirement: Configurable error correction
The wiring layer SHALL wire `learning.Engine` as the agent's `ErrorFixProvider` when `errorCorrectionEnabled` is true (default) and the knowledge system is available.

#### Scenario: Error correction enabled by default
- **WHEN** config omits `agent.errorCorrectionEnabled` and knowledge system is enabled
- **THEN** the agent SHALL have error correction wired

#### Scenario: Error correction explicitly disabled
- **WHEN** config sets `agent.errorCorrectionEnabled: false`
- **THEN** the agent SHALL NOT have error correction wired regardless of knowledge system state

#### Scenario: Knowledge system unavailable
- **WHEN** knowledge system is disabled
- **THEN** error correction SHALL NOT be wired even if `errorCorrectionEnabled` is true
