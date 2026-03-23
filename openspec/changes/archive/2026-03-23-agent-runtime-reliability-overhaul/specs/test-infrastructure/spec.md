## ADDED Requirements

### Requirement: Transcript replay fixtures for multi-agent runtime failures
The test infrastructure SHALL provide sanitized transcript replay fixtures for real multi-agent runtime failures so that end-to-end harness regressions can be reproduced without live network or external model dependencies.

#### Scenario: Vault balance loop fixture reproduces the failure shape
- **WHEN** the transcript replay harness runs the sanitized vault-balance-loop fixture
- **THEN** the test SHALL reproduce repeated same-signature specialist tool calls and a missing visible completion
- **AND** the runtime assertions SHALL evaluate the resulting classified outcome

#### Scenario: Replay fixture avoids external dependencies
- **WHEN** transcript replay tests execute in CI
- **THEN** they SHALL run without live Telegram, RPC, or external LLM access
- **AND** they SHALL rely only on local fixtures and test doubles

### Requirement: End-to-end assertions cover isolation and outcome parity
Replay-driven integration tests SHALL assert both persistence invariants and user-facing outcome parity across channel and gateway entrypoints.

#### Scenario: Isolated raw turns do not leak into parent history
- **WHEN** a replay fixture exercises an isolated specialist loop
- **THEN** the resulting persisted parent history SHALL contain only summary/discard entries
- **AND** raw specialist assistant/tool turns SHALL remain absent

#### Scenario: Channel and gateway classify the same failure identically
- **WHEN** the same replay fixture is executed through channel-style and gateway-style turn runners
- **THEN** both paths SHALL report the same terminal classification (for example `loop_detected` or `empty_after_tool_use`)
- **AND** both paths SHALL reference the same trace-backed root-cause summary semantics
