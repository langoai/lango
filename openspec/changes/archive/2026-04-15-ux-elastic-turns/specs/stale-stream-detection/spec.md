## ADDED Requirements

### Requirement: Stale stream watchdog timer
The `Runner` SHALL maintain a configurable `staleTimeout` (default 30 seconds). A watchdog timer SHALL be started when the first streaming chunk arrives. Each subsequent chunk SHALL reset the timer. If the timer fires (no chunk for `staleTimeout`), the current attempt context SHALL be cancelled.

#### Scenario: Timer resets on chunk
- **WHEN** a streaming chunk arrives during an active turn
- **THEN** the stale watchdog timer SHALL be reset to `staleTimeout`

#### Scenario: Stale stream detected
- **WHEN** no streaming chunk arrives for `staleTimeout` after the first chunk
- **THEN** the watchdog SHALL cancel the current attempt context
- **AND** the retry loop SHALL treat this as a retryable failure

#### Scenario: Timer inactive before first chunk
- **WHEN** the agent is executing tools (no streaming chunks yet)
- **THEN** the stale watchdog SHALL NOT be active
- **AND** no stale detection SHALL occur

#### Scenario: Timer inactive after turn completes
- **WHEN** the turn completes successfully
- **THEN** the stale watchdog timer SHALL be stopped

### Requirement: Stale timeout configurable
The `staleTimeout` SHALL be configurable via `Runner` configuration. If not set, it SHALL default to 30 seconds.

#### Scenario: Custom stale timeout
- **WHEN** `Runner` is configured with `staleTimeout = 15s`
- **THEN** the watchdog SHALL fire after 15 seconds of no chunks

#### Scenario: Default stale timeout
- **WHEN** `Runner` is configured without a `staleTimeout` value
- **THEN** the watchdog SHALL use 30 seconds as the default
