## MODIFIED Requirements

### Requirement: Context panel displays live system metrics
The ContextPanel SHALL display token usage, tool stats, runtime status (when active), channel statuses (when channels exist), and system uptime. The runtime section SHALL appear between Tool Stats and Channels.

#### Scenario: Runtime section when turn is active
- **WHEN** `SetRuntimeStatus` is called with `IsRunning=true`, `ActiveAgent="operator"`, `DelegationCount=3`, `TurnTokens=1234`
- **THEN** the context panel SHALL display a "Runtime" section with a green running indicator, agent name, delegation count, and formatted token count

#### Scenario: Runtime section hidden when idle
- **WHEN** `SetRuntimeStatus` is called with `IsRunning=false`
- **THEN** the "Runtime" section SHALL NOT be rendered (graceful degradation)

#### Scenario: Runtime status refreshed on tick
- **WHEN** a contextTickMsg fires and a RuntimeTracker is available
- **THEN** the cockpit SHALL push `runtimeTracker.Snapshot()` to the context panel alongside channel statuses

#### Scenario: Zero delegations not displayed
- **WHEN** `DelegationCount=0` in the runtime status
- **THEN** the delegation line SHALL NOT be rendered

#### Scenario: Zero tokens not displayed
- **WHEN** `TurnTokens=0` in the runtime status
- **THEN** the token line SHALL NOT be rendered
