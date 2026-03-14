## MODIFIED Requirements

### Requirement: IdleTimeout config field
The `AgentConfig` SHALL include an `IdleTimeout` field of type `time.Duration`. When positive, idle timeout mode is active. When negative (-1), idle timeout is explicitly disabled. When zero (default), behavior depends on other config fields.

The `deadline` package SHALL provide a `ResolveTimeouts(cfg TimeoutConfig) (idleTimeout, hardCeiling time.Duration)` function that encapsulates all timeout resolution logic. Both channel handlers (`runAgent`) and the gateway (`initGateway`) SHALL use this function as the single source of truth for timeout computation.

#### Scenario: IdleTimeout set to 2m
- **WHEN** config has `idleTimeout: 2m` and `requestTimeout: 30m`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=2m, ceiling=30m

#### Scenario: IdleTimeout not set
- **WHEN** config has only `requestTimeout: 5m`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=0, ceiling=5m (fixed timeout, backward compatible)

#### Scenario: IdleTimeout set to -1
- **WHEN** config has `idleTimeout: -1`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=0 (disabled), using fixed RequestTimeout

#### Scenario: Legacy AutoExtendTimeout
- **WHEN** config has `autoExtendTimeout: true` with `requestTimeout: 5m` and `maxRequestTimeout: 15m`
- **THEN** `deadline.ResolveTimeouts()` SHALL return idle=5m, ceiling=15m

#### Scenario: Gateway uses same resolution as channels
- **WHEN** `initGateway()` constructs gateway config
- **THEN** it SHALL call `deadline.ResolveTimeouts()` with the same `TimeoutConfig` fields used by `runAgent()`

### Requirement: Idle timeout in gateway
The gateway `handleChatMessage()` SHALL support idle timeout via `Config.IdleTimeout` and `Config.MaxTimeout` fields. When `IdleTimeout > 0`, it SHALL use `ExtendableDeadline` and pass `WithOnActivity` to `RunStreaming`. Error types in gateway timeout handling SHALL use `string(deadline.ReasonXxx)` constants instead of raw string literals.

#### Scenario: Gateway idle timeout fires
- **WHEN** gateway idle timeout is active and no agent activity occurs
- **THEN** the gateway SHALL cancel the request and broadcast an "agent.error" event with type `string(deadline.ReasonIdle)` (value: "idle")

#### Scenario: Gateway max timeout fires
- **WHEN** gateway max timeout is reached
- **THEN** the gateway SHALL broadcast an "agent.error" event with type `string(deadline.ReasonMaxTimeout)` (value: "max_timeout")
